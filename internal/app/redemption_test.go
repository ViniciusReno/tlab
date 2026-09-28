package app

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
	"github.com/ViniciusReno/tlab/internal/pricing"
)

func TestApprovedRedemptionTransitionAndPerRecordGate(t *testing.T) {
	s, _ := syncedQuotes(t)
	ctx := context.Background()
	f, err := os.Open("../pricing/testdata/redemption-transition.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dataset, err := tesouro.ParseDataset(f, "https://www.tesourotransparente.gov.br/ckan/")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.store.Upsert(ctx, dataset.Quotes); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../pricing/testdata/redemption-transition.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		BondID         string `json:"bond_id"`
		QuoteDate      string `json:"quote_date"`
		Convention     string
		SettlementDate string `json:"settlement_date"`
		BusinessDays   int    `json:"business_days"`
		Matches        bool
		Shocks         []struct{ Yield, PU, Variation float64 }
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	accepted, blocked := 0, 0
	for _, tc := range cases {
		want := "D0"
		if tc.QuoteDate < "2021-09-13" {
			want = "D+1"
		}
		if tc.Convention != want {
			continue
		}
		if tc.Matches {
			accepted++
		} else {
			blocked++
		}
		for _, shock := range tc.Shocks {
			out, err := s.Analyze(ctx, Request{BondID: tc.BondID, Source: "synced", Basis: "early_exit", Date: tc.QuoteDate, Yield: shock.Yield})
			if !tc.Matches {
				if err != bond.CalculationNotValidated {
					t.Fatalf("mismatch not blocked: %s %s %v", tc.BondID, tc.QuoteDate, err)
				}
				continue
			}
			// Independent fixed expectations: BRL 1e-9; fractional variation 1e-12.
			if err != nil || math.Abs(out.ScenarioPU-shock.PU) > 1e-9 || math.Abs(out.Variation-shock.Variation) > 1e-12 || out.BusinessDays != tc.BusinessDays || out.SettlementDate != tc.SettlementDate || out.SettlementConvention != want || out.SettlementVersion != pricing.SettlementVersion {
				t.Fatalf("approved transition: %+v %v", out, err)
			}
		}
	}
	if accepted != 6 || blocked != 4 {
		t.Fatalf("evidence coverage: %d accepted, %d blocked", accepted, blocked)
	}
	// Missing SellPU must not fall back to BasePU, even for matching modern fields.
	q := dataset.Quotes[0]
	q.SellPU = nil
	if err := s.store.Upsert(ctx, []bond.Quote{q}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Analyze(ctx, Request{BondID: q.Bond.ID, Source: "synced", Basis: "early_exit", Date: q.Date.Format("2006-01-02"), Yield: 0.1}); err != bond.MissingQuote {
		t.Fatalf("missing sell pair: %v", err)
	}
}
