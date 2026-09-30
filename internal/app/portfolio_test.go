package app

import (
	"context"
	"math"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
	"github.com/ViniciusReno/tlab/internal/portfolio"
	"github.com/ViniciusReno/tlab/internal/storage/sqlite"
)

func portfolioService(t *testing.T) *Service {
	t.Helper()
	s, err := OpenPersistent(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	for _, path := range []string{"../pricing/testdata/m3-quotes.csv", "../pricing/testdata/redemption-transition.csv"} {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		data, err := tesouro.ParseDataset(f, IPCASource)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if err = s.store.Upsert(context.Background(), data.Quotes); err != nil {
			t.Fatal(err)
		}
	}
	return s
}
func assertNumber(t *testing.T, got *float64, want float64) {
	t.Helper()
	if got == nil || !bond.Finite(*got) || math.Abs(*got-want) > 1e-9 {
		t.Fatalf("got %v want %g", got, want)
	}
}
func TestPortfolioOfficialValuationCompletenessAndMaturity(t *testing.T) {
	ctx := context.Background()
	s := portfolioService(t)
	for _, p := range []portfolio.Position{
		{BondID: "ipca:2029-05-15", Quantity: 2},
		{BondID: "selic:2029-03-01", Quantity: .5, InvestedBRL: floatPtr(9000)},
		{BondID: "prefixado:2024-07-01", Quantity: 2},
	} {
		if err := s.SavePosition(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	now, _ := bond.ParseDate("2026-09-29")
	view, err := s.Portfolio(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if view.Valued != 2 || view.Unavailable != 1 {
		t.Fatalf("counts: %+v", view)
	}
	assertNumber(t, view.Total, 17629.14)
	for _, p := range view.Positions {
		switch p.Quote.Bond.Kind {
		case "ipca":
			assertNumber(t, p.Value, 7731.5)
			if p.Change != nil || p.Completeness != "valuation_only" {
				t.Fatal("inferred acquisition")
			}
		case "selic":
			assertNumber(t, p.Value, 9897.64)
			assertNumber(t, p.Change, 897.64)
			if p.ComparisonError != "unsupported" {
				t.Fatal("Selic comparison")
			}
		case "prefixado":
			if p.Value != nil || p.Status != "matured" || p.Comparison != nil {
				t.Fatal("matured valued")
			}
		}
	}
	// The quote has a distinct BuyPU; missing BasePU cannot fall back to it.
	q, err := s.store.Quote(ctx, "ipca:2029-05-15", "")
	if err != nil {
		t.Fatal(err)
	}
	q.BasePU = nil
	if err = s.store.Upsert(ctx, []bond.Quote{q}); err != nil {
		t.Fatal(err)
	}
	view, err = s.Portfolio(ctx, now)
	if err != nil || view.Valued != 1 || view.Unavailable != 2 {
		t.Fatalf("missing base %+v %v", view, err)
	}
	assertNumber(t, view.Total, 9897.64)
}
func floatPtr(n float64) *float64 { return &n }
func TestPortfolioComparisonGateAndExactStoredQuantity(t *testing.T) {
	s := portfolioService(t)
	ctx := context.Background()
	now, _ := bond.ParseDate("2021-09-13")
	if err := s.SavePosition(ctx, portfolio.Position{BondID: "prefixado:2024-07-01", Quantity: 2.5}); err != nil {
		t.Fatal(err)
	}
	view, err := s.Portfolio(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	p := view.Positions[0]
	if p.Comparison == nil || p.Comparison.Basis.BusinessDays != 701 {
		t.Fatalf("%+v", p)
	}
	assertNumber(t, &p.Comparison.EarlyExitAmount, 1906.65)
	assertNumber(t, &p.Comparison.BreakEvenRate, .10230155977571715)
	result, amount, err := s.PositionScenario(ctx, p.ID, "2021-09-13", .0923)
	if err != nil {
		t.Fatal(err)
	}
	// Fixed official fixture scenario PU 782.241348739505 × stored 2.5.
	assertNumber(t, &amount, 1955.6033718487625)
	if result.Basis != "mark_to_market" || result.Amount != nil {
		t.Fatal("quantity reconstructed through amount")
	}
	q, _ := s.store.Quote(ctx, p.BondID, "")
	q.SellPU = floatPtr(750)
	if err = s.store.Upsert(ctx, []bond.Quote{q}); err != nil {
		t.Fatal(err)
	}
	view, err = s.Portfolio(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if view.Positions[0].Comparison != nil || view.Positions[0].ComparisonError != "calculation_not_validated" {
		t.Fatal("gate bypassed")
	}
	assertNumber(t, view.Positions[0].Value, 1906.65)
}
func TestPortfolioDemoRestartAndPersistentCRUD(t *testing.T) {
	ctx := context.Background()
	s := portfolioService(t)
	p := portfolio.Position{BondID: "selic:2029-03-01", Quantity: 1, Note: "private"}
	if err := s.SavePosition(ctx, p); err != nil {
		t.Fatal(err)
	}
	positions, _ := s.store.Positions(ctx)
	p = positions[0]
	p.Quantity = 3
	if err := s.SavePosition(ctx, p); err != nil {
		t.Fatal(err)
	}
	// Reopen the same persistent database, independently of private demo stores.
	reopened, err := OpenPersistent(ctx, s.directory)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	a, err := OpenDemo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.DeletePosition(ctx, "demo-ipca"); err != nil {
		t.Fatal(err)
	}
	if err := a.SavePosition(ctx, portfolio.Position{BondID: DemoBond, Quantity: 4}); err != nil {
		t.Fatal(err)
	}
	a.Close()
	b, err := OpenDemo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	demo, _ := b.store.Positions(ctx)
	if len(demo) != 1 || demo[0].ID != "demo-ipca" {
		t.Fatal("demo did not reset")
	}
	got, _ := reopened.store.Positions(ctx)
	if len(got) != 1 || got[0].Note != "private" || math.Abs(got[0].Quantity-3) > 1e-12 {
		t.Fatal("persistent data changed")
	}
	bad := got[0]
	bad.Quantity = -1
	if err := s.SavePosition(ctx, bad); err == nil {
		t.Fatal("accepted invalid update")
	}
	if err := s.DeletePosition(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeletePosition(ctx, p.ID); err != bond.InvalidInput {
		t.Fatal("missing delete accepted")
	}
	got, _ = reopened.store.Positions(ctx)
	if len(got) != 0 {
		t.Fatal("deletion not persisted")
	}
}
func TestPortfolioFormNormalization(t *testing.T) {
	v := url.Values{"bond": {"ipca:2029-05-15"}, "purchase_pu": {"700"}, "invested_brl": {"1400"}, "purchase_yield": {"6.5"}, "purchase_date": {"2026-09-04"}}
	p, err := ParsePosition(v)
	if err != nil || math.Abs(p.Quantity-2) > 1e-12 || math.Abs(*p.PurchaseYield-.065) > 1e-12 {
		t.Fatalf("%+v %v", p, err)
	}
	for _, tc := range []struct{ k, v string }{{"quantity", "0"}, {"quantity", "NaN"}, {"quantity", "1,5"}, {"purchase_date", "04/09/2026"}, {"source", "demo"}, {"purchase_pu", "-1"}} {
		values, _ := url.ParseQuery(v.Encode())
		values.Set(tc.k, tc.v)
		if _, err := ParsePosition(values); err == nil {
			t.Fatalf("accepted %s %s", tc.k, tc.v)
		}
	}
	v.Add("bond", "other")
	if _, err := ParsePosition(v); err == nil {
		t.Fatal("accepted duplicate")
	}
}

func TestPortfolioLocalDateBoundary(t *testing.T) {
	s := portfolioService(t)
	ctx := context.Background()
	if err := s.SavePosition(ctx, portfolio.Position{BondID: "prefixado:2024-07-01", Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		now     time.Time
		matured bool
	}{
		{time.Date(2024, 6, 30, 23, 59, 0, 0, time.FixedZone("local", -3*3600)), false},
		{time.Date(2024, 7, 1, 0, 0, 0, 0, time.FixedZone("local", -3*3600)), true},
	} {
		v, err := s.Portfolio(ctx, tc.now)
		if err != nil || (v.Positions[0].Status == "matured") != tc.matured {
			t.Fatalf("%+v %v", v, err)
		}
	}
}

func TestPortfolioOlderMissingAndFutureQuoteStates(t *testing.T) {
	s := portfolioService(t)
	ctx := context.Background()
	q, err := s.store.Quote(ctx, "ipca:2029-05-15", "")
	if err != nil {
		t.Fatal(err)
	}
	run := sqlite.SyncRun{ID: "portfolio-freshness", Source: IPCASource, StartedAt: "2026-09-29T00:00:00Z", Status: "success", RecordsWritten: 1, ImportedAt: "2026-09-29T00:00:00Z", DatasetMaxQuoteDate: "2026-09-28"}
	if err = s.store.StartSync(ctx, run); err != nil {
		t.Fatal(err)
	}
	if err = s.store.CommitSync(ctx, &run, []bond.Quote{q}); err != nil {
		t.Fatal(err)
	}
	if err = s.SavePosition(ctx, portfolio.Position{BondID: q.Bond.ID, Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	now, _ := bond.ParseDate("2026-09-29")
	v, err := s.Portfolio(ctx, now)
	if err != nil || v.Positions[0].Status != "older_quote" || v.LastQuoteDate != "2026-09-04" || v.DatasetDate != "2026-09-28" {
		t.Fatalf("older quote %+v %v", v, err)
	}
	assertNumber(t, v.Total, 3865.75)
	before, _ := bond.ParseDate("2026-09-03")
	v, err = s.Portfolio(ctx, before)
	if err != nil || v.Total != nil || v.Positions[0].Status != "future_quote_date" {
		t.Fatalf("future quote %+v %v", v, err)
	}
	q.BasePU = nil
	if err = s.store.Upsert(ctx, []bond.Quote{q}); err != nil {
		t.Fatal(err)
	}
	v, err = s.Portfolio(ctx, now)
	if err != nil || v.Total != nil || v.Positions[0].Status != "missing_base_quote" {
		t.Fatalf("missing base %+v %v", v, err)
	}
}
