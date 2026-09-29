package calendar

import (
	"testing"
	"time"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/bond"
)

func TestM3CalendarExtensionPreservesVerifiedDates(t *testing.T) {
	f, err := assets.Files.Open("data/calendar/anbima-2002-2050-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c, err := Load(f)
	if err != nil {
		t.Fatal(err)
	}
	old := expanded(t)
	if c.Version != "anbima-2002-2050-v1" || len(c.holidays) != 615 || c.end.Format(time.DateOnly) != "2050-12-31" {
		t.Fatal("unexpected M3 calendar")
	}
	for d := old.start; !d.After(old.end); d = d.AddDate(0, 0, 1) {
		if c.business(d) != old.business(d) {
			t.Fatalf("changed prior date: %s", d)
		}
	}
	// Fixed independent full-week/remainder counts minus official weekday holidays.
	// The final interval ends at the coverage boundary (2050-12-31 exclusive).
	counts := map[int]int{2033: 251, 2034: 248, 2035: 249, 2036: 253, 2037: 249, 2038: 251, 2039: 251, 2040: 250, 2041: 252, 2042: 252, 2043: 249, 2044: 251, 2045: 248, 2046: 249, 2047: 252, 2048: 250, 2049: 251, 2050: 251}
	for year, want := range counts {
		start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(1, 0, 0)
		if year == 2050 {
			end = c.end
		}
		if got, err := c.Count(start, end); err != nil || got != want {
			t.Fatalf("annual %d: %d %v", year, got, err)
		}
	}
	// Published 2040 table uses a four-digit year for Christmas; do not omit it.
	if c.business(date(t, "2040-12-25")) {
		t.Fatal("missed four-digit source date")
	}
	for _, d := range []string{"2001-12-31", "2051-01-01"} {
		if _, err := c.Next(date(t, d)); err != bond.CalendarOutOfRange {
			t.Fatalf("outside coverage: %v", err)
		}
	}
	if got, err := c.Next(date(t, "2050-02-18")); err != nil || got.Format(time.DateOnly) != "2050-02-23" {
		t.Fatalf("Carnival 2050: %s %v", got, err)
	}
}
