package sqlite

import (
	"context"
	"math"
	"strings"
	"testing"
	"testing/fstest"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
	"github.com/ViniciusReno/tlab/internal/portfolio"
)

func TestPortfolioMigrationFromM3PreservesQuotesAndPersistsPositions(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	old := fstest.MapFS{}
	for _, name := range []string{"migrations/001_initial.sql", "migrations/002_sync.sql"} {
		data, err := assets.Files.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		old[name] = &fstest.MapFile{Data: data}
	}
	s, err := OpenPersistent(ctx, dir, old)
	if err != nil {
		t.Fatal(err)
	}
	data, err := assets.Files.ReadFile("data/demo/ipca.csv")
	if err != nil {
		t.Fatal(err)
	}
	quotes, err := tesouro.Parse(strings.NewReader(string(data)), "migration-test")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Upsert(ctx, quotes); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s, err = OpenPersistent(ctx, dir, assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	p := portfolio.Position{ID: "position", BondID: quotes[0].Bond.ID, Quantity: 2}
	if err = s.SavePosition(ctx, p, true); err != nil {
		t.Fatal(err)
	}
	if err = s.SavePosition(ctx, p, true); err == nil {
		t.Fatal("duplicate insert accepted")
	}
	invalid := p
	invalid.BondID = "nonexistent"
	if err = s.SavePosition(ctx, invalid, false); err == nil {
		t.Fatal("foreign key not enforced")
	}
	s.Close()
	s, err = OpenPersistent(ctx, dir, assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rows, err := s.Positions(ctx)
	if err != nil || len(rows) != 1 || rows[0].BondID != p.BondID {
		t.Fatalf("positions %+v %v", rows, err)
	}
	q, err := s.Quote(ctx, p.BondID, "")
	if err != nil || q.BasePU == nil || math.Abs(*q.BasePU-1839.28) > 1e-9 {
		t.Fatalf("quote changed %+v %v", q, err)
	}
	// Reimport changes official data without replacing the user's position.
	if err = s.Upsert(ctx, quotes); err != nil {
		t.Fatal(err)
	}
	rows, err = s.Positions(ctx)
	if err != nil || len(rows) != 1 {
		t.Fatal("sync removed position")
	}
}
