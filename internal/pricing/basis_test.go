package pricing

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/calendar"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
)

func expandedCalendar(t *testing.T) *calendar.Calendar {
	t.Helper()
	f, err := assets.Files.Open("data/calendar/anbima-2002-2032-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c, err := calendar.Load(f)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func contextQuotes(t *testing.T) map[string]bond.Quote {
	t.Helper()
	f, err := os.Open("testdata/quote-contexts.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var provenance struct{ Source string }
	data, err := os.ReadFile("testdata/provenance.json")
	if err != nil {
		t.Fatal(err)
	}
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
	return quotes
}

func TestOfficialQuoteContextsAgainstIndependentExpectations(t *testing.T) {
	cal, quotes := expandedCalendar(t), contextQuotes(t)
	data, err := os.ReadFile("testdata/scenarios.json")
	if err != nil {
		t.Fatal(err)
	}
	var examples []struct {
		BondID              string  `json:"bond_id"`
		QuoteDate           string  `json:"quote_date"`
		Basis               string  `json:"basis"`
		SettlementDate      string  `json:"settlement_date"`
		BusinessDays        int     `json:"business_days"`
		CalendarVersion     string  `json:"calendar_version"`
		BasePU              float64 `json:"base_pu"`
		BaseYield           float64 `json:"base_yield"`
		TheoreticalPU       float64 `json:"theoretical_pu"`
		TruncatedDifference float64 `json:"truncated_difference"`
		ValidationStatus    string  `json:"validation_status"`
		Shocks              []struct{ Yield, PU, Variation float64 }
	}
	if err := json.Unmarshal(data, &examples); err != nil {
		t.Fatal(err)
	}
	if len(examples) != 12 {
		t.Fatalf("expected four quotes with all three contexts, got %d", len(examples))
	}
	for _, e := range examples {
		t.Run(e.QuoteDate+"/"+e.Basis, func(t *testing.T) {
			q, ok := quotes[e.BondID+"/"+e.QuoteDate]
			if !ok {
				t.Fatal("missing raw official quote")
			}
			b, err := ResolveBasis(q, e.Basis, cal)
			if err != nil {
				t.Fatal(err)
			}
			if b.Settlement.Format(time.DateOnly) != e.SettlementDate || b.BusinessDays != e.BusinessDays || cal.Version != e.CalendarVersion {
				t.Fatalf("basis: %+v", b)
			}
			near(t, b.PU, e.BasePU, 1e-9)
			near(t, b.Yield, e.BaseYield, 1e-12)
			theoretical, err := PrefixadoPU(b.Yield, b.BusinessDays)
			if err != nil {
				t.Fatal(err)
			}
			near(t, theoretical, e.TheoreticalPU, 1e-9)
			difference := math.Trunc(theoretical*100)/100 - b.PU
			near(t, difference, e.TruncatedDifference, 1e-9)
			if e.ValidationStatus == "calculation_not_validated" {
				// Regression evidence of a source/spec disagreement, not a relaxed
				// tolerance or a fixture accepted as valid D+1 standalone pricing.
				if math.Abs(difference) <= 0.01 || len(e.Shocks) != 0 {
					t.Fatal("known discrepancy was silently accepted")
				}
				return
			}
			if e.ValidationStatus != "validated" || len(e.Shocks) != 3 {
				t.Fatal("incomplete validation fixture")
			}
			near(t, math.Trunc(theoretical*100)/100, b.PU, 0.01)
			previous := math.Inf(1)
			for _, expected := range e.Shocks {
				in := Input{BasePU: b.PU, BaseYield: b.Yield, ScenarioYield: expected.Yield, BusinessDays: b.BusinessDays}
				r, err := Scenario(in)
				if err != nil {
					t.Fatal(err)
				}
				near(t, r.ScenarioPU, expected.PU, 1e-9)
				near(t, r.Variation, expected.Variation, 1e-12)
				if r.ScenarioPU >= previous {
					t.Fatal("price must fall when yield rises")
				}
				previous = r.ScenarioPU
				if math.Abs(expected.Yield-b.Yield) <= 1e-12 {
					near(t, r.ScenarioPU, b.PU, 1e-9)
					continue
				}
				for _, offset := range []int{-1, 1} {
					wrong := in
					wrong.BusinessDays += offset
					other, err := Scenario(wrong)
					if err != nil || math.Abs(other.ScenarioPU-expected.PU) < 1e-6 || math.Abs(other.Variation-expected.Variation) < 1e-9 {
						t.Fatal("expectation cannot detect a one-day error")
					}
				}
			}
		})
	}
}

func TestQuoteContextMissingPairsAndInvalidNumbers(t *testing.T) {
	cal := expandedCalendar(t)
	q := contextQuotes(t)["prefixado:2018-01-01/2016-02-05"]
	for _, context := range []string{"purchase", "mark_to_market", "early_exit"} {
		for _, field := range []string{"pu", "yield"} {
			for _, invalid := range []*float64{nil, floatPointer(0), floatPointer(-1), floatPointer(math.NaN()), floatPointer(math.Inf(1)), floatPointer(math.Inf(-1))} {
				if field == "yield" && invalid != nil && bond.ValidYield(*invalid) {
					continue
				}
				changed := q
				if field == "yield" {
					if context == "purchase" {
						changed.BuyYield = invalid
					} else {
						changed.SellYield = invalid
					}
				} else {
					switch context {
					case "purchase":
						changed.BuyPU = invalid
					case "mark_to_market":
						changed.BasePU = invalid
					case "early_exit":
						changed.SellPU = invalid
					}
				}
				want := bond.InvalidInput
				if invalid == nil {
					want = bond.MissingQuote
				}
				if _, err := ResolveBasis(changed, context, cal); err != want {
					t.Fatalf("%s %s: %v want %v", context, field, err, want)
				}
			}
		}
	}
	for _, kind := range []string{"selic", "ipca", "prefixado_coupon", "unsupported"} {
		changed := q
		changed.Bond.Kind = kind
		if _, err := ResolveBasis(changed, "purchase", cal); err != bond.Unsupported {
			t.Fatalf("accepted %s: %v", kind, err)
		}
	}
	if _, err := ResolveBasis(q, "unknown", cal); err != bond.InvalidInput {
		t.Fatal(err)
	}
	if _, err := ResolveBasis(q, "purchase", nil); err != bond.InvalidInput {
		t.Fatal(err)
	}
}

func floatPointer(n float64) *float64 { return &n }

func TestQuoteContextMaturityAndCalendarBounds(t *testing.T) {
	cal := expandedCalendar(t)
	q := contextQuotes(t)["prefixado:2018-01-01/2016-02-05"]
	for _, tc := range []struct {
		date, maturity, context string
		want                    error
	}{
		{"2016-02-05", "2016-02-10", "purchase", bond.NoRemainingTerm},
		{"2026-09-04", "2026-09-08", "early_exit", nil},
		{"2026-09-08", "2026-09-08", "early_exit", bond.NoRemainingTerm},
		{"2026-09-05", "2026-09-08", "early_exit", bond.NoRemainingTerm},
		{"2016-02-05", "2016-02-10", "early_exit", bond.NoRemainingTerm},
		{"2016-02-05", "2016-02-10", "mark_to_market", nil},
		{"2016-02-05", "2016-02-09", "purchase", bond.NoRemainingTerm},
		{"2016-02-06", "2016-02-10", "mark_to_market", bond.NoRemainingTerm},
		{"2018-01-01", "2018-01-01", "purchase", bond.NoRemainingTerm},
		{"2018-01-02", "2018-01-01", "mark_to_market", bond.NoRemainingTerm},
		{"2001-12-31", "2018-01-01", "purchase", bond.CalendarOutOfRange},
		{"2001-12-31", "2001-12-31", "purchase", bond.CalendarOutOfRange},
		{"2033-01-02", "2033-01-01", "mark_to_market", bond.CalendarOutOfRange},
		{"2016-02-05", "2033-01-01", "purchase", bond.CalendarOutOfRange},
		{"2016-02-05", "2033-01-01", "mark_to_market", bond.CalendarOutOfRange},
		{"2032-12-31", "2033-01-01", "early_exit", bond.CalendarOutOfRange},
	} {
		t.Run(tc.date+"/"+tc.maturity+"/"+tc.context, func(t *testing.T) {
			q.Date, _ = bond.ParseDate(tc.date)
			q.Bond.Maturity, _ = bond.ParseDate(tc.maturity)
			_, err := ResolveBasis(q, tc.context, cal)
			if err != tc.want {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
}

func TestRedemptionValidationInputsAndTolerance(t *testing.T) {
	for _, tc := range []struct {
		basis Basis
		want  error
	}{
		{Basis{PU: 1000, Yield: 0, BusinessDays: 1}, nil},
		{Basis{PU: 999.99, Yield: 0, BusinessDays: 1}, nil},
		{Basis{PU: 999.98, Yield: 0, BusinessDays: 1}, bond.CalculationNotValidated},
		{Basis{PU: 0, Yield: 0, BusinessDays: 1}, bond.InvalidInput},
		{Basis{PU: math.NaN(), Yield: 0, BusinessDays: 1}, bond.InvalidInput},
		{Basis{PU: math.Inf(1), Yield: 0, BusinessDays: 1}, bond.InvalidInput},
		{Basis{PU: 1000, Yield: math.NaN(), BusinessDays: 1}, bond.InvalidInput},
		{Basis{PU: 1000, Yield: 0, BusinessDays: 0}, bond.NoRemainingTerm},
		{Basis{PU: 1000, Yield: -0.9999999999999999, BusinessDays: 10000}, bond.CalculationOutOfRange},
	} {
		if err := ValidateRedemption(tc.basis); err != tc.want {
			t.Fatalf("%+v: got %v want %v", tc.basis, err, tc.want)
		}
	}
}
