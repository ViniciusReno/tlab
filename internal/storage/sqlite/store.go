// Package sqlite persists normalized data without implementing financial formulas.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

// OpenMemory creates a connection-private database; it never reads a user data path.
func OpenMemory(ctx context.Context, assets fs.FS) (*Store, error) {
	return open(ctx, ":memory:", assets)
}

// OpenPersistent opens a fixed database filename in a locally selected directory.
func OpenPersistent(ctx context.Context, directory string, assets fs.FS) (*Store, error) {
	if directory == "" {
		return nil, fmt.Errorf("data directory is required")
	}
	directory, err := filepath.Abs(directory)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	path := filepath.Join(directory, "tesouro-lab.db")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	// Encode reserved URI characters so directory names cannot become SQLite options.
	uriPath := filepath.ToSlash(path)
	if !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}
	uri := url.URL{Scheme: "file", Path: uriPath}
	return open(ctx, uri.String(), assets)
}

func open(ctx context.Context, dsn string, assets fs.FS) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if _, err = db.ExecContext(ctx, "PRAGMA busy_timeout = 5000"); err == nil {
		_, err = db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
	}
	if err == nil {
		err = s.migrate(ctx, assets)
	}
	if err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) HasQuotes(ctx context.Context) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM market_quotes)").Scan(&exists)
	return exists, err
}

func (s *Store) migrate(ctx context.Context, assets fs.FS) error {
	if _, err := s.db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY)"); err != nil {
		return err
	}
	entries, err := fs.ReadDir(assets, "migrations")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		version, err := strconv.Atoi(strings.SplitN(entry.Name(), "_", 2)[0])
		if err != nil {
			return err
		}
		var count int
		if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM schema_migrations WHERE version = ?", version).Scan(&count); err != nil {
			return err
		}
		if count != 0 {
			continue
		}
		content, err := fs.ReadFile(assets, "migrations/"+entry.Name())
		if err != nil {
			return err
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, string(content)); err == nil {
			_, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES (?)", version)
		}
		if err != nil {
			tx.Rollback()
			return err
		}
		if err = tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Upsert(ctx context.Context, quotes []bond.Quote) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range quotes {
		if q.Bond.ID == "" || q.Source == "" || q.Date.IsZero() || q.Bond.Maturity.IsZero() {
			return bond.InvalidInput
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO bonds(id,kind,name,maturity) VALUES (?,?,?,?)
            ON CONFLICT(id) DO UPDATE SET kind=excluded.kind,name=excluded.name,maturity=excluded.maturity`,
			q.Bond.ID, q.Bond.Kind, q.Bond.Name, q.Bond.Maturity.Format(time.DateOnly))
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO market_quotes(bond_id,quote_date,buy_yield,sell_yield,buy_pu,sell_pu,base_pu,source)
            VALUES (?,?,?,?,?,?,?,?) ON CONFLICT(bond_id,quote_date) DO UPDATE SET
            buy_yield=excluded.buy_yield,sell_yield=excluded.sell_yield,buy_pu=excluded.buy_pu,
            sell_pu=excluded.sell_pu,base_pu=excluded.base_pu,source=excluded.source`,
			q.Bond.ID, q.Date.Format(time.DateOnly), q.BuyYield, q.SellYield, q.BuyPU, q.SellPU, q.BasePU, q.Source)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Quote(ctx context.Context, id, date string) (bond.Quote, error) {
	query := `SELECT b.id,b.kind,b.name,b.maturity,q.quote_date,q.buy_yield,q.sell_yield,q.buy_pu,q.sell_pu,q.base_pu,q.source
        FROM market_quotes q JOIN bonds b ON b.id=q.bond_id WHERE b.id=?`
	args := []any{id}
	if date != "" {
		query += " AND q.quote_date=?"
		args = append(args, date)
	}
	query += " ORDER BY q.quote_date DESC LIMIT 1"
	var q bond.Quote
	var maturity, quoteDate string
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&q.Bond.ID, &q.Bond.Kind, &q.Bond.Name, &maturity, &quoteDate,
		&q.BuyYield, &q.SellYield, &q.BuyPU, &q.SellPU, &q.BasePU, &q.Source)
	if err == sql.ErrNoRows {
		return q, bond.MissingQuote
	}
	if err != nil {
		return q, err
	}
	q.Bond.Maturity, err = bond.ParseDate(maturity)
	if err != nil {
		return q, fmt.Errorf("stored maturity: %w", err)
	}
	q.Date, err = bond.ParseDate(quoteDate)
	return q, err
}
