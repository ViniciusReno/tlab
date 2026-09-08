// Package pricing contains deterministic financial calculations, independent of I/O.
package pricing

import (
	"math"

	"github.com/ViniciusReno/tlab/internal/bond"
)

const Version = "anchored-zero-coupon-v1"

type Input struct {
	BasePU        float64  `json:"base_pu"`
	BaseYield     float64  `json:"base_yield"`
	ScenarioYield float64  `json:"scenario_yield"`
	BusinessDays  int      `json:"business_days"`
	Amount        *float64 `json:"amount,omitempty"`
}

type Result struct {
	Input
	ScenarioPU     float64  `json:"scenario_pu"`
	Variation      float64  `json:"variation"`
	Quantity       *float64 `json:"quantity,omitempty"`
	ScenarioAmount *float64 `json:"scenario_amount,omitempty"`
}

func Scenario(in Input) (Result, error) {
	if !bond.Positive(in.BasePU) || !bond.ValidYield(in.BaseYield) || !bond.ValidYield(in.ScenarioYield) ||
		(in.Amount != nil && !bond.Positive(*in.Amount)) {
		return Result{}, bond.InvalidInput
	}
	if in.BusinessDays <= 0 {
		return Result{}, bond.NoRemainingTerm
	}
	// Log1p/Exp avoids overflowing the intermediate yield ratio near -100%.
	factor := math.Exp((math.Log1p(in.BaseYield) - math.Log1p(in.ScenarioYield)) * float64(in.BusinessDays) / 252)
	price := in.BasePU * factor
	variation := price/in.BasePU - 1
	if !bond.Positive(price) || !bond.Finite(variation) {
		return Result{}, bond.CalculationOutOfRange
	}
	out := Result{Input: in, ScenarioPU: price, Variation: variation}
	if in.Amount != nil {
		amount := *in.Amount
		out.Amount = &amount
		quantity := amount / in.BasePU
		simulated := quantity * price
		if !bond.Positive(quantity) || !bond.Positive(simulated) {
			return Result{}, bond.CalculationOutOfRange
		}
		out.Quantity, out.ScenarioAmount = &quantity, &simulated
	}
	return out, nil
}

// PrefixadoPU is the standalone theoretical price, before presentation/truncation.
func PrefixadoPU(yield float64, days int) (float64, error) {
	if !bond.ValidYield(yield) {
		return 0, bond.InvalidInput
	}
	if days <= 0 {
		return 0, bond.NoRemainingTerm
	}
	pu := 1000 * math.Exp(-math.Log1p(yield)*float64(days)/252)
	if !bond.Positive(pu) {
		return 0, bond.CalculationOutOfRange
	}
	return pu, nil
}
