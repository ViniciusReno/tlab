package app

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
)

func syncedQuotes(t *testing.T) (*Service, []bond.Quote) {
	t.Helper()
	s, err := OpenPersistent(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	f, err := os.Open("../pricing/testdata/quote-contexts.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	data, err := os.ReadFile("../pricing/testdata/provenance.json")
	if err != nil {
		t.Fatal(err)
	}
	var provenance struct{ Source string }
	if err := json.Unmarshal(data, &provenance); err != nil {
		t.Fatal(err)
	}
	dataset, err := tesouro.ParseDataset(f, provenance.Source)
	if err != nil {
		t.Fatal(err)
	}
	for i := range dataset.Quotes {
		dataset.Quotes[i].ImportedAt = "2026-09-15T12:00:00Z"
	}
	if err := s.store.Upsert(context.Background(), dataset.Quotes); err != nil {
		t.Fatal(err)
	}
	return s, dataset.Quotes
}

func TestSyncedScenariosAgainstIndependentFixtures(t *testing.T) {
	s, _ := syncedQuotes(t)
	data, err := os.ReadFile("../pricing/testdata/scenarios.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Result
		Shocks []struct{ Yield, PU, Variation float64 }
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, tc := range cases {
		t.Run(tc.BondID+"/"+tc.Basis, func(t *testing.T) {
			r := Request{BondID: tc.BondID, Source: "synced", Basis: tc.Basis, Date: tc.QuoteDate, Yield: tc.BaseYield}
			for _, shock := range tc.Shocks {
				r.Yield = shock.Yield
				out, err := s.Analyze(ctx, r)
				if err != nil {
					t.Fatal(err)
				}
				// Fixed independent PU tolerance BRL 1e-9; yields/variation 1e-12.
				if math.Abs(out.ScenarioPU-shock.PU) > 1e-9 || math.Abs(out.Variation-shock.Variation) > 1e-12 || math.Abs(out.BasePU-tc.BasePU) > 1e-9 || math.Abs(out.BaseYield-tc.BaseYield) > 1e-12 {
					t.Fatalf("independent expectation: %+v, want %+v", out, shock)
				}
				if out.BusinessDays != tc.BusinessDays || out.SettlementDate != tc.SettlementDate || out.QuoteDate != tc.QuoteDate || out.CalendarVersion != tc.CalendarVersion || out.FixtureVersion != "" || out.ImportedAt != "2026-09-15T12:00:00Z" {
					t.Fatalf("resolved metadata: %+v", out)
				}
				replay, err := ParseRequest(out.Values(), "synced")
				if err != nil {
					t.Fatal(err)
				}
				again, err := s.Analyze(ctx, replay)
				if err != nil || math.Abs(again.ScenarioPU-out.ScenarioPU) > 1e-9 {
					t.Fatalf("replay: %+v %v", again, err)
				}
				if strings.Contains(out.Values().Encode(), "data-dir") || !strings.Contains(out.Command(), "--data-dir '") {
					t.Fatal("local directory missing from command or leaked into URL")
				}
			}
		})
	}
}

func TestSyncedSelectionAndBoundaries(t *testing.T) {
	s, quotes := syncedQuotes(t)
	ctx := context.Background()
	q := quotes[0]
	r := Request{BondID: q.Bond.ID, Source: "synced", Basis: "purchase", Yield: 0.12}
	latest, err := s.Analyze(ctx, r)
	if err != nil || latest.QuoteDate != q.Date.Format(time.DateOnly) {
		t.Fatalf("latest: %+v %v", latest, err)
	}
	// A newer incomplete record must not cause a fallback to the previous pair.
	newer := q
	newer.Date = q.Date.AddDate(0, 0, 1)
	newer.BuyPU, newer.BasePU = nil, nil
	if err := s.store.Upsert(ctx, []bond.Quote{newer}); err != nil {
		t.Fatal(err)
	}
	for _, basis := range []string{"purchase", "mark_to_market"} {
		r.Basis = basis
		if _, err := s.Analyze(ctx, r); err != bond.MissingQuote {
			t.Fatalf("latest incomplete %s: %v", basis, err)
		}
		r.Date = newer.Date.Format(time.DateOnly)
		if _, err := s.Analyze(ctx, r); err != bond.MissingQuote {
			t.Fatalf("exact incomplete %s: %v", basis, err)
		}
		r.Date = q.Date.Format(time.DateOnly)
		if _, err := s.Analyze(ctx, r); err != nil {
			t.Fatalf("historical original: %v", err)
		}
		r.Date = ""
	}
	r.Basis, r.Date = "purchase", q.Date.AddDate(0, 0, 2).Format(time.DateOnly)
	if _, err := s.Analyze(ctx, r); err != bond.MissingQuote {
		t.Fatalf("missing exact date: %v", err)
	}
	r.Date = q.Date.Format(time.DateOnly)
	r.Yield = math.NaN()
	if _, err := s.Analyze(ctx, r); err != bond.InvalidInput {
		t.Fatalf("nonfinite: %v", err)
	}
	r.Yield = -0.9999999999999999
	if out, err := s.Analyze(ctx, r); err != nil || !bond.Positive(out.ScenarioPU) {
		t.Fatalf("representable near-boundary yield: %+v %v", out, err)
	}
	// The unit price is still representable for this short historical term.
	// A finite maximum gross amount makes the derived scenario amount overflow.
	amount := math.MaxFloat64
	r.Amount = &amount
	if _, err := s.Analyze(ctx, r); err != bond.CalculationOutOfRange {
		t.Fatalf("numeric range: %v", err)
	}
	r.Amount = nil
	r.Yield, r.BondID = 0.12, "selic:2032-01-01"
	if _, err := s.Analyze(ctx, r); err != bond.Unsupported {
		t.Fatalf("unsupported: %v", err)
	}
	for _, tc := range []struct {
		date, maturity string
		want           error
	}{
		{"2026-09-04", "2026-09-08", bond.NoRemainingTerm},
		{"2026-09-08", "2026-09-08", bond.NoRemainingTerm},
		{"2026-09-04", "2033-01-01", bond.CalendarOutOfRange},
	} {
		changed := q
		changed.Date, _ = bond.ParseDate(tc.date)
		changed.Bond.Maturity, _ = bond.ParseDate(tc.maturity)
		changed.Bond.ID = "prefixado:" + tc.maturity
		if err := s.store.Upsert(ctx, []bond.Quote{changed}); err != nil {
			t.Fatal(err)
		}
		r.BondID, r.Date = changed.Bond.ID, tc.date
		if _, err := s.Analyze(ctx, r); err != tc.want {
			t.Fatalf("boundary %+v: %v", tc, err)
		}
	}
}

func TestSyncedLatestMovesWhileExactDateRemainsReproducible(t *testing.T) {
	s, quotes := syncedQuotes(t)
	ctx := context.Background()
	q := quotes[0]
	r := Request{BondID: q.Bond.ID, Source: "synced", Basis: "purchase", Yield: 0.12}
	original, err := s.Analyze(ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	// Synthetic later record exercises selection, not official pricing validation.
	newer := q
	newer.Date = q.Date.AddDate(0, 0, 1)
	newer.Source = "synthetic-test"
	if err := s.store.Upsert(ctx, []bond.Quote{newer}); err != nil {
		t.Fatal(err)
	}
	latest, err := s.Analyze(ctx, r)
	if err != nil || latest.QuoteDate != newer.Date.Format(time.DateOnly) || latest.BusinessDays != original.BusinessDays-1 {
		t.Fatalf("latest selection: %+v %v", latest, err)
	}
	r.Date = original.QuoteDate
	fixed, err := s.Analyze(ctx, r)
	// Exact-date replay keeps the unrounded result within BRL 1e-9.
	if err != nil || fixed.QuoteDate != original.QuoteDate || fixed.BusinessDays != original.BusinessDays || math.Abs(fixed.ScenarioPU-original.ScenarioPU) > 1e-9 {
		t.Fatalf("fixed-date replay: %+v %v", fixed, err)
	}
}
