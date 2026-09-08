package sqlite

import (
	"context"
	"math"
	"strings"
	"testing"

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
