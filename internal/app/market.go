package app

import (
	"context"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
)

type MarketQuote struct {
	bond.Quote
	Status string
}

type Market struct {
	Quotes                          []MarketQuote
	DatasetMaxQuoteDate, ImportedAt string
	Freshness                       DatasetFreshness
}

// DatasetFreshness is a conservative local-data warning, not a source SLA.
type DatasetFreshness struct {
	State, Reason, ExpectedDate, CalendarVersion string
	BusinessDaysBehind                           int
}

func (s *Service) datasetFreshness(quoteDate string, now time.Time) DatasetFreshness {
	f := DatasetFreshness{State: "unavailable", Reason: "missing_metadata", CalendarVersion: s.calendar.Version}
	if quoteDate == "" {
		return f
	}
	quote, err := bond.ParseDate(quoteDate)
	if err != nil {
		f.Reason = "invalid_quote_date"
		return f
	}
	// Preserve the user's local calendar date instead of converting it to UTC.
	localDate, _ := bond.ParseDate(now.Format(time.DateOnly))
	if !s.calendar.Covers(quote) || !s.calendar.Covers(localDate) {
		f.Reason = "calendar_out_of_range"
		return f
	}
	if quote.After(localDate) {
		f.Reason = "future_quote_date"
		return f
	}
	expected, err := s.calendar.Previous(localDate)
	if err != nil {
		f.Reason = "calendar_out_of_range"
		return f
	}
	f.ExpectedDate = expected.Format(time.DateOnly)
	if quote.Before(expected) {
		// Count business dates after the stored quote through the expected date.
		next, err := s.calendar.Next(quote)
		if err != nil {
			f.Reason = "calendar_out_of_range"
			return f
		}
		f.BusinessDaysBehind, err = s.calendar.Count(next, localDate)
		if err != nil {
			f.Reason = string(bond.CalendarOutOfRange)
			return f
		}
	}
	f.State, f.Reason = "within_window", ""
	if f.BusinessDaysBehind >= 2 {
		f.State = "appears_stale"
	}
	return f
}

// Market compares calendar dates in the user's local timezone, not UTC instants.
// Missing successful-sync metadata stays unknown, including for legacy databases.
func (s *Service) Market(ctx context.Context, now time.Time, includeMatured bool) (Market, error) {
	if s.source != "synced" {
		return Market{}, bond.SourceMismatch
	}
	data, err := s.store.Market(ctx)
	if err != nil {
		return Market{}, err
	}
	m := Market{DatasetMaxQuoteDate: data.DatasetMaxQuoteDate, ImportedAt: data.ImportedAt}
	m.Freshness = s.datasetFreshness(data.DatasetMaxQuoteDate, now)
	for _, q := range data.Quotes {
		status := "freshness_unknown"
		switch {
		case now.Format(time.DateOnly) >= q.Bond.Maturity.Format(time.DateOnly):
			if !includeMatured {
				continue
			}
			status = "matured"
		case q.Date.IsZero() || q.BuyPU == nil || q.BuyYield == nil:
			status = "no_quote"
		case q.Date.Format(time.DateOnly) == data.DatasetMaxQuoteDate:
			status = "latest"
		case data.DatasetMaxQuoteDate != "" && q.Date.Format(time.DateOnly) < data.DatasetMaxQuoteDate:
			status = "older_quote"
		}
		m.Quotes = append(m.Quotes, MarketQuote{Quote: q, Status: status})
	}
	return m, nil
}

type History struct {
	Quotes []bond.Quote
	Before string
}

func (s *Service) History(ctx context.Context, id, before string) (History, error) {
	if s.source != "synced" {
		return History{}, bond.SourceMismatch
	}
	if id == "" {
		return History{}, bond.InvalidInput
	}
	if before != "" {
		if _, err := bond.ParseDate(before); err != nil {
			return History{}, err
		}
	}
	quotes, err := s.store.History(ctx, id, before)
	if err != nil {
		return History{}, err
	}
	if len(quotes) == 0 {
		return History{}, bond.MissingQuote
	}
	h := History{Quotes: quotes}
	if len(quotes) > 100 {
		h.Quotes = quotes[:100]
		h.Before = h.Quotes[99].Date.Format(time.DateOnly)
	}
	return h, nil
}
