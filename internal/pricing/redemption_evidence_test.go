package pricing

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
)

// This test preserves contradictory evidence; it does not authorize a new
// runtime settlement convention or certify the mismatching official rows.
func TestRedemptionTransitionEvidence(t *testing.T) {
	cal := expandedCalendar(t)
	f, err := os.Open("testdata/redemption-transition.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	data, err := os.ReadFile("testdata/redemption-transition-provenance.json")
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
	quotes := map[string]bond.Quote{}
	for _, q := range dataset.Quotes {
		quotes[q.Bond.ID+"/"+q.Date.Format(time.DateOnly)] = q
	}
	data, err = os.ReadFile("testdata/redemption-transition.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		BondID               string `json:"bond_id"`
		QuoteDate            string `json:"quote_date"`
		Convention           string
		SettlementDate       string  `json:"settlement_date"`
		BusinessDays         int     `json:"business_days"`
		OfficialPU           float64 `json:"official_pu"`
		BaseYield            float64 `json:"base_yield"`
		TheoreticalTruncated float64 `json:"theoretical_truncated"`
		Difference           float64
		Matches              bool
		Shocks               []struct{ Yield, PU, Variation float64 }
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 10 || len(cases) != 20 {
		t.Fatal("incomplete transition evidence")
	}
	for _, tc := range cases {
		t.Run(tc.BondID+"/"+tc.QuoteDate+"/"+tc.Convention, func(t *testing.T) {
			q, ok := quotes[tc.BondID+"/"+tc.QuoteDate]
			if !ok || q.SellPU == nil || q.SellYield == nil {
				t.Fatal("missing official pair")
			}
			settlement := q.Date
			if tc.Convention == "D+1" {
				settlement, err = cal.Next(q.Date)
				if err != nil {
					t.Fatal(err)
				}
			}
			days, err := cal.Count(settlement, q.Bond.Maturity)
			if err != nil || days != tc.BusinessDays || settlement.Format(time.DateOnly) != tc.SettlementDate {
				t.Fatalf("term: %d %v", days, err)
			}
			// Absolute tolerances: PU BRL 1e-9; decimal yield/variation 1e-12.
			if math.Abs(*q.SellPU-tc.OfficialPU) > 1e-9 || math.Abs(*q.SellYield-tc.BaseYield) > 1e-12 {
				t.Fatal("official inputs changed")
			}
			theoretical, err := PrefixadoPU(*q.SellYield, days)
			if err != nil {
				t.Fatal(err)
			}
			truncated := math.Trunc(theoretical*100) / 100
			difference := truncated - *q.SellPU
			if math.Abs(truncated-tc.TheoreticalTruncated) > 1e-9 || math.Abs(difference-tc.Difference) > 1e-9 || (math.Abs(difference) <= 0.01) != tc.Matches {
				t.Fatalf("evidence changed: %v", difference)
			}
			for _, shock := range tc.Shocks {
				out, err := Scenario(Input{BasePU: *q.SellPU, BaseYield: *q.SellYield, ScenarioYield: shock.Yield, BusinessDays: days})
				if err != nil || math.Abs(out.ScenarioPU-shock.PU) > 1e-9 || math.Abs(out.Variation-shock.Variation) > 1e-12 {
					t.Fatalf("independent hypothetical result: %+v %v", out, err)
				}
			}
		})
	}
}
