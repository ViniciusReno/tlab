package pricing

import (
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/calendar"
)

// Basis is the official PU/yield pair and verified term for a scenario context.
// It does not replace an official price with a standalone theoretical price.
type Basis struct {
	PU           float64
	Yield        float64
	Settlement   time.Time
	BusinessDays int
}

// ResolveBasis applies the locked V1 D0/D+1 contract at the selected historical
// quote date. It performs no I/O and does not consult the machine's current date.
func ResolveBasis(q bond.Quote, context string, cal *calendar.Calendar) (Basis, error) {
	if q.Bond.Kind != "prefixado" {
		return Basis{}, bond.Unsupported
	}
	var pu, yield *float64
	switch context {
	case "purchase":
		pu, yield = q.BuyPU, q.BuyYield
	case "mark_to_market":
		pu, yield = q.BasePU, q.SellYield
	case "early_exit":
		pu, yield = q.SellPU, q.SellYield
	default:
		return Basis{}, bond.InvalidInput
	}
	if pu == nil || yield == nil {
		return Basis{}, bond.MissingQuote
	}
	if !bond.Positive(*pu) || !bond.ValidYield(*yield) || q.Date.IsZero() || q.Bond.Maturity.IsZero() || cal == nil {
		return Basis{}, bond.InvalidInput
	}
	if !cal.Covers(q.Date) || !cal.Covers(q.Bond.Maturity) {
		return Basis{}, bond.CalendarOutOfRange
	}
	settlement := q.Date
	if !settlement.Before(q.Bond.Maturity) {
		return Basis{}, bond.NoRemainingTerm
	}
	var err error
	if context != "mark_to_market" {
		settlement, err = cal.Next(q.Date)
		if err != nil {
			return Basis{}, err
		}
	}
	days, err := cal.Count(settlement, q.Bond.Maturity)
	if err != nil {
		return Basis{}, err
	}
	return Basis{PU: *pu, Yield: *yield, Settlement: settlement, BusinessDays: days}, nil
}
