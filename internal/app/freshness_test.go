package app

import (
	"context"
	"testing"
	"time"
)

func TestDatasetFreshnessThresholdAndVerifiedDates(t *testing.T) {
	s, err := OpenPersistent(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, tc := range []struct {
		name, quote, local, state, reason, expected string
		lag                                         int
	}{
		{"same date", "2026-09-10", "2026-09-10", "within_window", "", "2026-09-09", 0},
		{"expected date", "2026-09-09", "2026-09-10", "within_window", "", "2026-09-09", 0},
		{"one day tolerated", "2026-09-08", "2026-09-10", "within_window", "", "2026-09-09", 1},
		{"two days warns", "2026-09-04", "2026-09-10", "appears_stale", "", "2026-09-09", 2},
		{"holiday Monday", "2026-09-04", "2026-09-07", "within_window", "", "2026-09-04", 0},
		{"after holiday", "2026-09-04", "2026-09-08", "within_window", "", "2026-09-04", 0},
		{"Saturday", "2026-09-09", "2026-09-12", "appears_stale", "", "2026-09-11", 2},
		{"Sunday", "2026-09-09", "2026-09-13", "appears_stale", "", "2026-09-11", 2},
		{"Monday", "2026-09-09", "2026-09-14", "appears_stale", "", "2026-09-11", 2},
		{"Carnival", "2016-02-05", "2016-02-10", "within_window", "", "2016-02-05", 0},
		{"November holiday", "2024-11-19", "2024-11-21", "within_window", "", "2024-11-19", 0},
		{"missing metadata", "", "2026-09-10", "unavailable", "missing_metadata", "", 0},
		{"malformed metadata", "invalid", "2026-09-10", "unavailable", "invalid_quote_date", "", 0},
		{"future metadata", "2026-09-11", "2026-09-10", "unavailable", "future_quote_date", "", 0},
		{"old quote outside coverage", "2001-12-31", "2026-09-10", "unavailable", "calendar_out_of_range", "", 0},
		{"local date outside coverage", "2032-12-30", "2033-01-01", "unavailable", "calendar_out_of_range", "", 0},
		{"no verified predecessor", "2002-01-01", "2002-01-01", "unavailable", "calendar_out_of_range", "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			local, err := time.ParseInLocation(time.DateOnly, tc.local, time.FixedZone("UTC-3", -3*60*60))
			if err != nil {
				t.Fatal(err)
			}
			// At 23:30, UTC has advanced a date; classification must stay local.
			out := s.datasetFreshness(tc.quote, local.Add(23*time.Hour+30*time.Minute))
			if out.State != tc.state || out.Reason != tc.reason || out.ExpectedDate != tc.expected || out.BusinessDaysBehind != tc.lag || out.CalendarVersion != "anbima-2002-2032-v1" {
				t.Fatalf("got %+v; want %+v", out, tc)
			}
		})
	}
}
