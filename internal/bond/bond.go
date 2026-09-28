// Package bond defines normalized instruments, quotes, and explicit domain errors.
package bond

import (
	"math"
	"time"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	InvalidInput            Error = "invalid_input"
	MissingQuote            Error = "missing_quote"
	Unsupported             Error = "unsupported"
	SourceUnavailable       Error = "source_unavailable"
	SyncFailed              Error = "sync_failed"
	SourceMismatch          Error = "source_mismatch"
	SchemaChanged           Error = "source_schema_changed"
	CalendarOutOfRange      Error = "calendar_out_of_range"
	NoRemainingTerm         Error = "no_remaining_term"
	CalculationOutOfRange   Error = "calculation_out_of_range"
	CalculationNotValidated Error = "calculation_not_validated"
)

func Message(err error) string {
	switch err {
	case MissingQuote:
		return "No official quote is available for this date and context."
	case Unsupported:
		return "This instrument or feature is not supported in this milestone."
	case SourceUnavailable:
		return "The official source is unavailable. The offline demo remains available with tesouro-lab demo."
	case CalculationNotValidated:
		return "Synchronized early-redemption scenarios are unavailable while the official settlement convention is being validated. Purchase and official base PU scenarios remain available."
	case SyncFailed:
		return "Synchronization failed. Previously stored quotes are preserved; inspect the sync report and retry."
	case SourceMismatch:
		return "The requested source does not match this server. Start tesouro-lab demo for demo data or tesouro-lab for persistent data."
	case CalendarOutOfRange:
		return "The requested dates are outside the selected calendar's verified coverage."
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
	ImportedAt            string // UTC synchronization timestamp; empty for demo or legacy rows.
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
