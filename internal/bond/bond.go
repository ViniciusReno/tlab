// Package bond defines normalized instruments, quotes, and explicit domain errors.
package bond

import (
	"math"
	"time"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	InvalidInput          Error = "invalid_input"
	MissingQuote          Error = "missing_quote"
	Unsupported           Error = "unsupported"
	SourceUnavailable     Error = "source_unavailable"
	SourceMismatch        Error = "source_mismatch"
	SchemaChanged         Error = "source_schema_changed"
	CalendarOutOfRange    Error = "calendar_out_of_range"
	NoRemainingTerm       Error = "no_remaining_term"
	CalculationOutOfRange Error = "calculation_out_of_range"
)

func Message(err error) string {
	switch err {
	case MissingQuote:
		return "No official quote is available for this date and context."
	case Unsupported:
		return "This instrument or feature is not supported in this milestone."
	case SourceUnavailable:
		return "Official synchronization and synchronized analysis are not implemented yet. Start the offline demo with tesouro-lab demo."
	case SourceMismatch:
		return "The requested source does not match this server. Start tesouro-lab demo for demo data or tesouro-lab for persistent data."
	case CalendarOutOfRange:
		return "The requested dates are outside verified calendar coverage (2012–2015)."
	case NoRemainingTerm:
		return "There are no remaining business days between settlement and maturity."
	case CalculationOutOfRange:
		return "The result is outside the supported numeric range."
	case SchemaChanged:
		return "The source columns do not match the expected format."
	default:
		return "Invalid input. Use ISO dates, decimal-point numbers, positive amounts, and yields greater than -100%."
	}
}

type Bond struct {
	ID       string
	Kind     string
	Name     string
	Maturity time.Time
}

type Quote struct {
	Bond                  Bond
	Date                  time.Time
	BuyYield, SellYield   *float64
	BuyPU, SellPU, BasePU *float64
	Source                string
}

func ParseDate(value string) (time.Time, error) {
	d, err := time.Parse(time.DateOnly, value)
	if err != nil || d.Format(time.DateOnly) != value {
		return time.Time{}, InvalidInput
	}
	return d, nil
}

func Finite(n float64) bool     { return !math.IsNaN(n) && !math.IsInf(n, 0) }
func Positive(n float64) bool   { return Finite(n) && n > 0 }
func ValidYield(n float64) bool { return Finite(n) && n > -1 }
