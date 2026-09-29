package app

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
	"github.com/ViniciusReno/tlab/internal/pricing"
)

func TestIPCAIndependentScenariosDemoAndSynced(t *testing.T) {
	ctx := context.Background()
	synced, err := OpenPersistent(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer synced.Close()
	demo, err := OpenDemo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer demo.Close()
	f, err := os.Open("../pricing/testdata/m3-quotes.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := tesouro.ParseDataset(f, IPCASource)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := synced.store.Upsert(ctx, rows.Quotes); err != nil {
			t.Fatal(err)
		}
	}
	market, err := synced.Market(ctx, time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC), true)
	if err != nil || len(market.Quotes) != 4 {
		t.Fatalf("M3 storage idempotency: %+v %v", market, err)
	}
	data, err := os.ReadFile("../pricing/testdata/ipca-scenarios.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		BondID         string `json:"bond_id"`
		QuoteDate      string `json:"quote_date"`
		Basis          string
		SettlementDate string  `json:"settlement_date"`
		BusinessDays   int     `json:"business_days"`
		BasePU         float64 `json:"base_pu"`
		BaseYield      float64 `json:"base_yield"`
		Shocks         []struct{ Yield, PU, Variation float64 }
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 9 {
		t.Fatal("missing IPCA context evidence")
	}
	for _, s := range []*Service{synced, demo} {
		for _, tc := range cases {
			if s.Source() == "demo" && tc.BondID != DemoIPCABond {
				continue
			}
			t.Run(s.Source()+"/"+tc.BondID+"/"+tc.Basis, func(t *testing.T) {
				previous := math.Inf(1)
				for _, shock := range tc.Shocks {
					req := Request{BondID: tc.BondID, Source: s.Source(), Basis: tc.Basis, Date: tc.QuoteDate, Yield: shock.Yield}
					out, err := s.Analyze(ctx, req)
					// Fixed independent tolerances: BRL 1e-9; decimal yields/variation 1e-12.
					if err != nil || math.Abs(out.ScenarioPU-shock.PU) > 1e-9 || math.Abs(out.Variation-shock.Variation) > 1e-12 || math.Abs(out.BasePU-tc.BasePU) > 1e-9 || math.Abs(out.BaseYield-tc.BaseYield) > 1e-12 || out.BusinessDays != tc.BusinessDays || out.SettlementDate != tc.SettlementDate || out.Kind != "ipca" || out.YieldType != "real" {
						t.Fatalf("independent result: %+v %v", out, err)
					}
					if out.ScenarioPU >= previous {
						t.Fatal("IPCA price/yield direction")
					}
					previous = out.ScenarioPU
					if s.Source() == "demo" && (out.FixtureVersion != IPCAFixtureVersion || out.CalendarVersion != "anbima-2012-2015-v1" || out.Provenance != IPCASource) {
						t.Fatalf("demo provenance: %+v", out)
					}
					replay, err := ParseRequest(out.Values(), s.Source())
					if err != nil {
						t.Fatal(err)
					}
					again, err := s.Analyze(ctx, replay)
					if err != nil || math.Abs(again.ScenarioPU-out.ScenarioPU) > 1e-9 {
						t.Fatal("IPCA replay")
					}
					if math.Abs(shock.Yield-tc.BaseYield) > 1e-12 {
						for _, offset := range []int{-1, 1} {
							wrong := out.Input
							wrong.BusinessDays += offset
							changed, err := pricing.Scenario(wrong)
							if err != nil || math.Abs(changed.ScenarioPU-shock.PU) < 1e-6 {
								t.Fatal("fixture cannot detect a one-day error")
							}
						}
					} else if math.Abs(out.ScenarioPU-out.BasePU) > 1e-9 {
						t.Fatal("zero shock")
					}
				}
			})
		}
	}
	// Selic is preserved for display, but no context reaches the scenario engine.
	for _, basis := range []string{"purchase", "mark_to_market", "early_exit"} {
		if _, err := synced.Analyze(ctx, Request{BondID: "selic:2029-03-01", Source: "synced", Basis: basis, Yield: 0.1}); err != bond.Unsupported {
			t.Fatalf("Selic %s: %v", basis, err)
		}
	}
	selic, err := synced.store.Quote(ctx, "selic:2029-03-01", "2026-09-04")
	if err != nil || selic.BasePU == nil || math.Abs(*selic.BasePU-19795.28) > 1e-9 {
		t.Fatalf("Selic official MTM quote: %+v %v", selic, err)
	}
}

func TestIPCAMissingInputsAndLimits(t *testing.T) {
	s, err := OpenPersistent(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	date, _ := bond.ParseDate("2026-09-04")
	maturity, _ := bond.ParseDate("2029-05-15")
	pu, yield := 3876.96, 0.078
	q := bond.Quote{Bond: bond.Bond{ID: "ipca:2029-05-15", Kind: "ipca", Name: "Synthetic test IPCA+", Maturity: maturity}, Date: date, BuyPU: &pu, BuyYield: &yield, Source: "synthetic-test"}
	if err := s.store.Upsert(ctx, []bond.Quote{q}); err != nil {
		t.Fatal(err)
	}
	r := Request{BondID: q.Bond.ID, Source: "synced", Basis: "mark_to_market", Date: "2026-09-04", Yield: 0.068}
	if _, err := s.Analyze(ctx, r); err != bond.MissingQuote {
		t.Fatal(err)
	}
	r.Basis = "purchase"
	r.Date = "2026-09-05"
	if _, err := s.Analyze(ctx, r); err != bond.MissingQuote {
		t.Fatal(err)
	}
	r.Date = "2026-09-04"
	r.Yield = math.NaN()
	if _, err := s.Analyze(ctx, r); err != bond.InvalidInput {
		t.Fatal(err)
	}
	r.Yield = -0.9999999999999999
	amount := math.MaxFloat64
	r.Amount = &amount
	if _, err := s.Analyze(ctx, r); err != bond.CalculationOutOfRange {
		t.Fatal(err)
	}
	r.Yield = 0.068
	r.Amount = nil
	for _, tc := range []struct {
		maturity string
		want     error
	}{{"2026-09-08", bond.NoRemainingTerm}, {"2026-09-04", bond.NoRemainingTerm}, {"2051-05-15", bond.CalendarOutOfRange}} {
		q.Bond.Maturity, _ = bond.ParseDate(tc.maturity)
		q.Bond.ID = "ipca:" + tc.maturity
		r.BondID = q.Bond.ID
		if err := s.store.Upsert(ctx, []bond.Quote{q}); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Analyze(ctx, r); err != tc.want {
			t.Fatalf("%s: %v", tc.maturity, err)
		}
	}
}
