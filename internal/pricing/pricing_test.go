package pricing

import (
	"encoding/json"
	"math"
	"testing"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/bond"
)

func near(t *testing.T, got, want, tolerance float64) {
	t.Helper()
	if !bond.Finite(got) || math.Abs(got-want) > tolerance {
		t.Fatalf("got %.16g want %.16g tolerance %.3g", got, want, tolerance)
	}
}

func TestIndependentOfficialFixture(t *testing.T) {
	bytes, err := assets.Files.ReadFile("data/demo/scenarios.json")
	if err != nil {
		t.Fatal(err)
	}
	var examples []struct{ Yield, PU, Variation float64 }
	if err := json.Unmarshal(bytes, &examples); err != nil {
		t.Fatal(err)
	}
	amount := 10000.0
	for _, e := range examples {
		out, err := Scenario(Input{BasePU: 733.86, BaseYield: 0.1088, ScenarioYield: e.Yield, BusinessDays: 755, Amount: &amount})
		if err != nil {
			t.Fatal(err)
		}
		near(t, out.ScenarioPU, e.PU, 1e-9)
		near(t, out.Variation, e.Variation, 1e-12)
		near(t, *out.ScenarioAmount, amount*(1+e.Variation), 1e-8)
		if math.Abs(e.Yield-0.1088) > 1e-12 {
			wrong, err := Scenario(Input{BasePU: 733.86, BaseYield: 0.1088, ScenarioYield: e.Yield, BusinessDays: 754})
			if err != nil || math.Abs(wrong.ScenarioPU-e.PU) < 1e-6 {
				t.Fatal("fixture cannot detect a one-day error")
			}
		}
	}
	theoretical, err := PrefixadoPU(0.1088, 755)
	if err != nil {
		t.Fatal(err)
	}
	near(t, math.Trunc(theoretical*100)/100, 733.86, 0.01)
}

func TestScenarioInvariants(t *testing.T) {
	previous := math.Inf(1)
	for _, yield := range []float64{-0.2, 0, 0.0688, 0.1088, 0.1488, 1} {
		r, err := Scenario(Input{BasePU: 733.86, BaseYield: 0.1088, ScenarioYield: yield, BusinessDays: 755})
		if err != nil {
			t.Fatal(err)
		}
		if r.ScenarioPU >= previous {
			t.Fatal("price must decrease as yield increases")
		}
		previous = r.ScenarioPU
	}
	for _, pu := range []float64{0.001, 733.86, 10000000} {
		r, err := Scenario(Input{BasePU: pu, BaseYield: 0.1088, ScenarioYield: 0.1088, BusinessDays: 755})
		if err != nil {
			t.Fatal(err)
		}
		near(t, r.ScenarioPU, pu, 1e-9)
	}
}

func TestInvalidScenarios(t *testing.T) {
	valid := Input{BasePU: 733.86, BaseYield: 0.1088, ScenarioYield: 0.0888, BusinessDays: 755}
	for _, n := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), 0, -1} {
		in := valid
		in.BasePU = n
		if _, err := Scenario(in); err != bond.InvalidInput {
			t.Fatalf("PU %v: %v", n, err)
		}
		in = valid
		in.Amount = &n
		if _, err := Scenario(in); err != bond.InvalidInput {
			t.Fatalf("amount %v: %v", n, err)
		}
	}
	for _, n := range []float64{-1, -2, math.NaN(), math.Inf(1)} {
		in := valid
		in.ScenarioYield = n
		if _, err := Scenario(in); err != bond.InvalidInput {
			t.Fatalf("yield %v: %v", n, err)
		}
		in = valid
		in.BaseYield = n
		if _, err := Scenario(in); err != bond.InvalidInput {
			t.Fatalf("base yield %v: %v", n, err)
		}
		if _, err := PrefixadoPU(n, 755); err != bond.InvalidInput {
			t.Fatalf("standalone yield %v: %v", n, err)
		}
	}
	for _, days := range []int{0, -1} {
		in := valid
		in.BusinessDays = days
		if _, err := Scenario(in); err != bond.NoRemainingTerm {
			t.Fatal(err)
		}
		if _, err := PrefixadoPU(0.1, days); err != bond.NoRemainingTerm {
			t.Fatal(err)
		}
	}
	for _, yield := range []float64{-0.9999999999999999, 1e100} {
		in := valid
		in.ScenarioYield = yield
		in.BusinessDays = 100000
		if _, err := Scenario(in); err != bond.CalculationOutOfRange {
			t.Fatalf("numeric range: %v", err)
		}
	}
}
