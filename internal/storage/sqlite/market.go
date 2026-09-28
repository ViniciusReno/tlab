package sqlite

import (
	"context"
	"database/sql"

	"github.com/ViniciusReno/tlab/internal/bond"
)

type Market struct {
	Quotes              []bond.Quote
	DatasetMaxQuoteDate string
	ImportedAt          string
}

const quoteColumns = `b.id,b.kind,b.name,b.maturity,COALESCE(q.quote_date,''),
	q.buy_yield,q.sell_yield,q.buy_pu,q.sell_pu,q.base_pu,COALESCE(q.source,''),COALESCE(q.imported_at,'')`

// Market reads quotes and successful-sync metadata from one SQLite snapshot.
func (s *Store) Market(ctx context.Context) (Market, error) {
	var market Market
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return market, err
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(ctx, `SELECT dataset_max_quote_date,imported_at FROM sync_runs
		WHERE status='success' ORDER BY finished_at DESC,id DESC LIMIT 1`).Scan(&market.DatasetMaxQuoteDate, &market.ImportedAt)
	if err != nil && err != sql.ErrNoRows {
		return market, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT `+quoteColumns+` FROM bonds b LEFT JOIN market_quotes q
		ON q.bond_id=b.id AND q.quote_date=(SELECT MAX(quote_date) FROM market_quotes WHERE bond_id=b.id)
		ORDER BY b.maturity,b.id`)
	if err != nil {
		return market, err
	}
	defer rows.Close()
	for rows.Next() {
		q, err := scanQuote(rows)
		if err != nil {
			return market, err
		}
		market.Quotes = append(market.Quotes, q)
	}
	if err := rows.Err(); err != nil {
		return market, err
	}
	if err := rows.Close(); err != nil {
		return market, err
	}
	return market, tx.Commit()
}

// History returns at most 101 rows: 100 visible rows and one pagination probe.
// The exclusive date cursor never fills missing dates or context fields.
func (s *Store) History(ctx context.Context, id, before string) ([]bond.Quote, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+quoteColumns+` FROM market_quotes q
		JOIN bonds b ON b.id=q.bond_id WHERE b.id=? AND (?='' OR q.quote_date < ?)
		ORDER BY q.quote_date DESC LIMIT 101`, id, before, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var quotes []bond.Quote
	for rows.Next() {
		q, err := scanQuote(rows)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, q)
	}
	return quotes, rows.Err()
}
