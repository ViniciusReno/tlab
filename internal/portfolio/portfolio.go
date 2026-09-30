// Package portfolio implements deterministic position normalization and gross math.
package portfolio

import (
	"math"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/calendar"
	"github.com/ViniciusReno/tlab/internal/pricing"
)

type Position struct {
	ID, BondID                             string
	Quantity                               float64
	PurchaseDate                           *time.Time
	PurchasePU, PurchaseYield, InvestedBRL *float64
	Note                                   string
}

const InconsistentCost bond.Error = "inconsistent_acquisition_cost"

// Normalize derives only mathematically determined fields, never market inputs.
func Normalize(p Position) (Position, error) {
	if p.BondID == "" || len(p.Note) > 2000 || !bond.Finite(p.Quantity) || p.Quantity < 0 {
		return p, bond.InvalidInput
	}
	for _, n := range []*float64{p.PurchasePU, p.InvestedBRL} {
		if n != nil && !bond.Positive(*n) {
			return p, bond.InvalidInput
		}
	}
	if p.PurchaseYield != nil && !bond.ValidYield(*p.PurchaseYield) {
		return p, bond.InvalidInput
	}
	if p.PurchaseDate != nil && p.PurchaseDate.IsZero() {
		return p, bond.InvalidInput
	}
	if p.Quantity == 0 {
		if p.InvestedBRL == nil || p.PurchasePU == nil {
			return p, bond.InvalidInput
		}
		p.Quantity = *p.InvestedBRL / *p.PurchasePU
		if !bond.Positive(p.Quantity) {
			return p, bond.CalculationOutOfRange
		}
	}
	if p.PurchasePU != nil {
		amount, err := Value(p.Quantity, *p.PurchasePU)
		if err != nil {
			return p, err
		}
		if p.InvestedBRL == nil {
			p.InvestedBRL = &amount
		} else if math.Abs(*p.InvestedBRL-amount) > math.Max(.02, *p.InvestedBRL*.0001) {
			return p, InconsistentCost
		}
	}
	return p, nil
}

func Value(quantity, pu float64) (float64, error) {
	if !bond.Positive(quantity) || !bond.Positive(pu) {
		return 0, bond.InvalidInput
	}
	n := quantity * pu
	if !bond.Positive(n) {
		return 0, bond.CalculationOutOfRange
	}
	return n, nil
}

func Change(value, cost float64) (float64, float64, error) {
	if !bond.Positive(value) || !bond.Positive(cost) {
		return 0, 0, bond.InvalidInput
	}
	delta, rate := value-cost, value/cost-1
	if !bond.Finite(delta) || !bond.Finite(rate) {
		return 0, 0, bond.CalculationOutOfRange
	}
	return delta, rate, nil
}

type Comparison struct {
	MaturityAmount, EarlyExitAmount, BreakEvenRate float64
	Basis                                          pricing.Basis
}

// HoldExit validates the official Prefixado sell context before comparing paths.
func HoldExit(q bond.Quote, quantity float64, cal *calendar.Calendar) (Comparison, error) {
	if q.Bond.Kind != "prefixado" {
		return Comparison{}, bond.Unsupported
	}
	b, err := pricing.ResolveBasis(q, "early_exit", cal)
	if err != nil {
		return Comparison{}, err
	}
	if err = pricing.ValidateRedemption(b); err != nil {
		return Comparison{}, err
	}
	out, err := Compare(quantity, b.PU, b.BusinessDays)
	out.Basis = b
	return out, err
}

func Compare(quantity, sellPU float64, days int) (Comparison, error) {
	if days <= 0 {
		return Comparison{}, bond.NoRemainingTerm
	}
	maturity, err := Value(quantity, 1000)
	if err != nil {
		return Comparison{}, err
	}
	early, err := Value(quantity, sellPU)
	if err != nil {
		return Comparison{}, err
	}
	rate := math.Expm1((math.Log(maturity) - math.Log(early)) * 252 / float64(days))
	if !bond.ValidYield(rate) {
		return Comparison{}, bond.CalculationOutOfRange
	}
	return Comparison{MaturityAmount: maturity, EarlyExitAmount: early, BreakEvenRate: rate}, nil
}

// Allocation is a gross subtotal at the input quotes' dates, not a target weight.
type Allocation struct {
	Kind         string
	Value, Share float64
}

func Aggregate(values []Allocation) (float64, []Allocation, error) {
	total := 0.0
	amounts := map[string]float64{}
	for _, v := range values {
		if !bond.Positive(v.Value) {
			return 0, nil, bond.InvalidInput
		}
		total += v.Value
		amounts[v.Kind] += v.Value
	}
	if len(values) == 0 {
		return 0, nil, nil
	}
	if !bond.Positive(total) {
		return 0, nil, bond.CalculationOutOfRange
	}
	var out []Allocation
	for _, kind := range []string{"prefixado", "ipca", "selic"} {
		if n := amounts[kind]; n > 0 {
			out = append(out, Allocation{Kind: kind, Value: n, Share: n / total})
		}
	}
	return total, out, nil
}
