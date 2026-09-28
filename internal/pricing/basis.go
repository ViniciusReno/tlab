package pricing

import (
	"math"
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
	Convention   string
}

const SettlementVersion = "morning-redemption-2021-v1"

// ResolveBasis applies the approved V1 D0/D+1 contract at the selected historical
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
	convention := "D0"
	if context == "purchase" || (context == "early_exit" && q.Date.Format(time.DateOnly) < "2021-09-13") {
		convention = "D+1"
		settlement, err = cal.Next(q.Date)
		if err != nil {
			return Basis{}, err
		}
	}
	days, err := cal.Count(settlement, q.Bond.Maturity)
	if err != nil {
		return Basis{}, err
	}
	return Basis{PU: *pu, Yield: *yield, Settlement: settlement, BusinessDays: days, Convention: convention}, nil
}

// ValidateRedemption keeps unresolved source/calendar discrepancies unavailable.
// It never chooses a settlement convention by fitting the official price.
func ValidateRedemption(b Basis) error {
	if !bond.Positive(b.PU) {
		return bond.InvalidInput
	}
	pu, err := PrefixadoPU(b.Yield, b.BusinessDays)
	if err != nil {
		return err
	}
	if !bond.Finite(pu * 100) {
		return bond.CalculationOutOfRange
	}
	if math.Abs(math.Trunc(pu*100)/100-b.PU) > 0.01 {
		return bond.CalculationNotValidated
	}
	return nil
}
