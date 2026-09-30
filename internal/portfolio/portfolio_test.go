package portfolio

import (
	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/calendar"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
	"math"
	"os"
	"testing"

	"github.com/ViniciusReno/tlab/internal/bond"
)

func ptr(n float64) *float64 { return &n }
func closeTo(t *testing.T, got, want, tolerance float64) {
	t.Helper()
	if !bond.Finite(got) || math.Abs(got-want) > tolerance {
		t.Fatalf("got %.17g want %.17g", got, want)
	}
}
func TestNormalizeCompletenessAndConsistency(t *testing.T) {
	for _, tc := range []struct {
		p        Position
		quantity float64
		cost     *float64
	}{
		{Position{BondID: "x", Quantity: 2}, 2, nil},
		{Position{BondID: "x", Quantity: 2, InvestedBRL: ptr(1400)}, 2, ptr(1400)},
		{Position{BondID: "x", Quantity: 2, PurchasePU: ptr(700)}, 2, ptr(1400)},
		{Position{BondID: "x", PurchasePU: ptr(700), InvestedBRL: ptr(1400)}, 2, ptr(1400)},
		{Position{BondID: "x", Quantity: 2, PurchasePU: ptr(700), InvestedBRL: ptr(1400.13)}, 2, ptr(1400.13)},
	} {
		got, err := Normalize(tc.p)
		if err != nil {
			t.Fatal(err)
		}
		closeTo(t, got.Quantity, tc.quantity, 1e-12)
		if tc.cost == nil {
			if got.InvestedBRL != nil {
				t.Fatal("invented cost")
			}
		} else {
			closeTo(t, *got.InvestedBRL, *tc.cost, 1e-9)
		}
	}
	for _, tc := range []struct {
		p   Position
		err error
	}{
		{Position{BondID: "x"}, bond.InvalidInput},
		{Position{BondID: "x", Quantity: math.NaN()}, bond.InvalidInput},
		{Position{BondID: "x", Quantity: math.Inf(1)}, bond.InvalidInput},
		{Position{BondID: "x", Quantity: -1}, bond.InvalidInput},
		{Position{BondID: "x", Quantity: 1, PurchasePU: ptr(0)}, bond.InvalidInput},
		{Position{BondID: "x", Quantity: 1, InvestedBRL: ptr(math.Inf(1))}, bond.InvalidInput},
		{Position{BondID: "x", Quantity: 1, PurchaseYield: ptr(-1)}, bond.InvalidInput},
		{Position{BondID: "x", Quantity: 2, PurchasePU: ptr(700), InvestedBRL: ptr(1401)}, InconsistentCost},
		{Position{BondID: "x", Quantity: math.MaxFloat64, PurchasePU: ptr(2)}, bond.CalculationOutOfRange},
		{Position{BondID: "x", PurchasePU: ptr(math.SmallestNonzeroFloat64), InvestedBRL: ptr(1)}, bond.CalculationOutOfRange},
	} {
		if _, err := Normalize(tc.p); err != tc.err {
			t.Fatalf("%+v: %v want %v", tc.p, err, tc.err)
		}
	}
}
func TestGrossMathAndIndependentBreakEven(t *testing.T) {
	// Fixed expectations from Python Decimal (precision 60), using independently
	// verified DU=701 from the official 2021-09-13 transition fixture. BRL 1e-9;
	// annual rate tolerance 1e-12. Expectations do not call production math.
	out, err := Compare(2.5, 762.66, 701)
	if err != nil {
		t.Fatal(err)
	}
	closeTo(t, out.MaturityAmount, 2500, 1e-9)
	closeTo(t, out.EarlyExitAmount, 1906.65, 1e-9)
	closeTo(t, out.BreakEvenRate, .10230155977571715, 1e-12)
	delta, rate, err := Change(1906.65, 2000)
	if err != nil {
		t.Fatal(err)
	}
	closeTo(t, delta, -93.35, 1e-9)
	closeTo(t, rate, -.046675, 1e-12)
	for _, tc := range []struct {
		q, pu float64
		days  int
		err   error
	}{
		{1, 1000, 0, bond.NoRemainingTerm}, {1, 1000, -1, bond.NoRemainingTerm},
		{0, 1000, 1, bond.InvalidInput}, {1, math.NaN(), 1, bond.InvalidInput},
		{math.MaxFloat64, 1000, 252, bond.CalculationOutOfRange},
		{1, math.SmallestNonzeroFloat64, 1, bond.CalculationOutOfRange},
	} {
		if _, err := Compare(tc.q, tc.pu, tc.days); err != tc.err {
			t.Fatalf("%+v: %v", tc, err)
		}
	}
	if _, err := Value(math.SmallestNonzeroFloat64, .1); err != bond.CalculationOutOfRange {
		t.Fatal(err)
	}
	if _, _, err := Change(math.MaxFloat64, math.SmallestNonzeroFloat64); err != bond.CalculationOutOfRange {
		t.Fatal(err)
	}
}

func TestAggregatePartialInputsAndNumericBounds(t *testing.T) {
	total, parts, err := Aggregate([]Allocation{{Kind: "ipca", Value: 600}, {Kind: "selic", Value: 400}})
	if err != nil || len(parts) != 2 {
		t.Fatal(err)
	}
	closeTo(t, total, 1000, 1e-9)
	closeTo(t, parts[0].Share, .6, 1e-12)
	if _, _, err := Aggregate([]Allocation{{Kind: "ipca", Value: math.MaxFloat64}, {Kind: "ipca", Value: math.MaxFloat64}}); err != bond.CalculationOutOfRange {
		t.Fatal(err)
	}
}

func TestHoldExitOfficialSettlementTransition(t *testing.T) {
	f, err := assets.Files.Open("data/calendar/anbima-2002-2050-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	cal, err := calendar.Load(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.Open("../pricing/testdata/redemption-transition.csv")
	if err != nil {
		t.Fatal(err)
	}
	data, err := tesouro.ParseDataset(raw, "official-fixture")
	raw.Close()
	if err != nil {
		t.Fatal(err)
	}
	matched := 0
	for _, q := range data.Quotes {
		if q.Bond.ID != "prefixado:2024-07-01" {
			continue
		}
		out, err := HoldExit(q, 2.5, cal)
		if err != nil {
			t.Fatal(err)
		}
		want := .10230155977571715
		convention := "D0"
		amount := 1906.65
		if q.Date.Format("2006-01-02") == "2021-09-10" {
			want = .10090206178697514
			convention = "D+1"
			amount = 1913.4
		}
		if out.Basis.BusinessDays != 701 || out.Basis.Convention != convention || out.Basis.Settlement.Format("2006-01-02") != "2021-09-13" {
			t.Fatalf("%+v", out)
		}
		closeTo(t, out.BreakEvenRate, want, 1e-12)
		closeTo(t, out.EarlyExitAmount, amount, 1e-9)
		matched++
		q.SellPU = nil
		if _, err := HoldExit(q, 2.5, cal); err != bond.MissingQuote {
			t.Fatal("missing sell replaced")
		}
	}
	if matched != 2 {
		t.Fatal("missing transition evidence")
	}
	for _, kind := range []string{"ipca", "selic", "coupon"} {
		if _, err := HoldExit(bond.Quote{Bond: bond.Bond{Kind: kind}}, 1, cal); err != bond.Unsupported {
			t.Fatal(err)
		}
	}
}
