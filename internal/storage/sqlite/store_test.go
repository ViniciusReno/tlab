package sqlite

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
)

func TestMigrationsUpsertRollbackAndIsolation(t *testing.T) {
	ctx := context.Background()
	s, err := OpenMemory(ctx, assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.migrate(ctx, assets.Files); err != nil {
		t.Fatal(err)
	}
	var migrations int
	if err := s.db.QueryRow("SELECT count(*) FROM schema_migrations").Scan(&migrations); err != nil || migrations != 1 {
		t.Fatalf("migrations %d %v", migrations, err)
	}
	data, err := assets.Files.ReadFile("data/demo/prefixado.csv")
	if err != nil {
		t.Fatal(err)
	}
	quotes, err := tesouro.Parse(strings.NewReader(string(data)), "test")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := s.Upsert(ctx, quotes); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := s.db.QueryRow("SELECT count(*) FROM market_quotes").Scan(&count); err != nil || count != 1 {
		t.Fatalf("quotes %d %v", count, err)
	}
	updated := quotes[0]
	pu := 800.0
	updated.BuyPU = &pu
	invalid := quotes[0]
	invalid.Source = ""
	if err := s.Upsert(ctx, []bond.Quote{updated, invalid}); err == nil {
		t.Fatal("invalid batch accepted")
	}
	stored, err := s.Quote(ctx, quotes[0].Bond.ID, "2012-01-03")
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(*stored.BuyPU-*quotes[0].BuyPU) > 1e-9 {
		t.Fatal("failed transaction changed existing quote")
	}
	if stored.BasePU != nil {
		t.Fatal("NULL lost")
	}
	other, err := OpenMemory(ctx, assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if _, err := other.Quote(ctx, quotes[0].Bond.ID, ""); err != bond.MissingQuote {
		t.Fatalf("memory databases shared: %v", err)
	}
	if _, err := s.Quote(ctx, quotes[0].Bond.ID, "2012-01-04"); err != bond.MissingQuote {
		t.Fatalf("date fallback: %v", err)
	}
}

func TestPersistentReopenAndMigrationRollback(t *testing.T) {
	ctx := context.Background()
	dir := filepath.Join(t.TempDir(), "data # percent%")
	s, err := OpenPersistent(ctx, dir, assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	if has, err := s.HasQuotes(ctx); err != nil || has {
		t.Fatalf("new database must be empty: %v %v", has, err)
	}
	data, err := assets.Files.ReadFile("data/demo/prefixado.csv")
	if err != nil {
		t.Fatal(err)
	}
	quotes, err := tesouro.Parse(strings.NewReader(string(data)), "persistence-test")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Upsert(ctx, quotes); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = OpenPersistent(ctx, dir, assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	broken := fstest.MapFS{"migrations/002_broken.sql": &fstest.MapFile{Data: []byte("CREATE TABLE rollback_probe(id INTEGER); INVALID SQL;")}}
	if err := s.migrate(ctx, broken); err == nil {
		t.Fatal("accepted broken migration")
	}
	var probe int
	if err := s.db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name='rollback_probe'").Scan(&probe); err != nil || probe != 0 {
		t.Fatalf("failed migration left partial schema: %d %v", probe, err)
	}
	q, err := s.Quote(ctx, quotes[0].Bond.ID, "2012-01-03")
	if err != nil || q.Source != "persistence-test" || q.BuyPU == nil || math.Abs(*q.BuyPU-733.86) > 1e-9 {
		t.Fatalf("quote not preserved: %+v %v", q, err)
	}
	var count int
	if err := s.db.QueryRow("SELECT count(*) FROM schema_migrations").Scan(&count); err != nil || count != 1 {
		t.Fatalf("migration reapplied: %d %v", count, err)
	}
	if has, err := s.HasQuotes(ctx); err != nil || !has {
		t.Fatalf("persisted database must contain quotes: %v %v", has, err)
	}
}

func TestPersistentRejectsInvalidDirectoryAndDatabase(t *testing.T) {
	ctx := context.Background()
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("sentinel"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"", file, filepath.Join(file, "child")} {
		if s, err := OpenPersistent(ctx, dir, assets.Files); err == nil {
			s.Close()
			t.Fatalf("accepted invalid directory %q", dir)
		}
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "tesouro-lab.db")
	if err := os.WriteFile(path, []byte("not a SQLite database"), 0600); err != nil {
		t.Fatal(err)
	}
	if s, err := OpenPersistent(ctx, dir, assets.Files); err == nil {
		s.Close()
		t.Fatal("accepted invalid database")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "not a SQLite database" {
		t.Fatal("invalid database was replaced")
	}
}
