package calendar

import (
	"testing"
	"time"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/bond"
)

func expanded(t *testing.T) *Calendar {
	t.Helper()
	f, err := assets.Files.Open("data/calendar/anbima-2002-2032-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c, err := Load(f)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestExpandedCalendarProvenanceAndDemoOverlap(t *testing.T) {
	c, old := expanded(t), fixture(t)
	if c.Version != "anbima-2002-2032-v1" || len(c.holidays) != 381 || c.start.Format(time.DateOnly) != "2002-01-01" || c.end.Format(time.DateOnly) != "2032-12-31" {
		t.Fatal("unexpected expanded calendar")
	}
	for d := old.start; !d.After(old.end); d = d.AddDate(0, 0, 1) {
		if c.business(d) != old.business(d) {
			t.Fatalf("changed demo calendar day: %s", d)
		}
	}
	// Official M1 methodology example must still match after calendar expansion.
	got, err := c.Count(date(t, "2012-01-04"), date(t, "2015-01-01"))
	if err != nil || got != 755 {
		t.Fatalf("official day count changed: %d %v", got, err)
	}
}

func TestExpandedSettlementHolidaysAndYearEnd(t *testing.T) {
	c := expanded(t)
	for _, tc := range []struct{ start, want string }{
		{"2016-02-05", "2016-02-10"},
		{"2023-11-17", "2023-11-20"},
		{"2024-11-19", "2024-11-21"},
		{"2026-09-04", "2026-09-08"},
		{"2026-12-30", "2026-12-31"},
		{"2026-12-31", "2027-01-04"},
		{"2032-03-25", "2032-03-29"},
	} {
		got, err := c.Next(date(t, tc.start))
		if err != nil || got.Format(time.DateOnly) != tc.want {
			t.Fatalf("%+v: %v %v", tc, got, err)
		}
	}
	for _, day := range []string{"2001-12-31", "2032-12-31", "2033-01-01"} {
		if _, err := c.Next(date(t, day)); err != bond.CalendarOutOfRange {
			t.Fatalf("%s: %v", day, err)
		}
	}
	for _, tc := range []struct {
		start, end string
		want       int
	}{
		{"2002-01-01", "2002-01-03", 1},
		{"2016-02-10", "2018-01-01", 475},
		{"2016-02-05", "2018-01-01", 476},
		{"2024-11-21", "2027-01-01", 529},
		{"2024-11-19", "2027-01-01", 530},
		{"2026-09-08", "2032-01-01", 1331},
		{"2026-09-04", "2032-01-01", 1332},
		{"2032-12-30", "2032-12-31", 1},
	} {
		got, err := c.Count(date(t, tc.start), date(t, tc.end))
		if err != nil || got != tc.want {
			t.Fatalf("%+v: %d %v", tc, got, err)
		}
	}
	if _, err := c.Count(date(t, "2032-01-01"), date(t, "2033-01-01")); err != bond.CalendarOutOfRange {
		t.Fatal(err)
	}
}

func TestExpandedCalendarAnnualCounts(t *testing.T) {
	c := expanded(t)
	// Fixed independent full-week/remainder counts minus weekday holidays.
	// The last entry ends on 2032-12-31 exclusive to stay inside verified bounds.
	counts := map[int]int{
		2002: 253, 2003: 253, 2004: 252, 2005: 251, 2006: 249, 2007: 250,
		2008: 254, 2009: 250, 2010: 251, 2011: 251, 2012: 251, 2013: 253,
		2014: 253, 2015: 250, 2016: 251, 2017: 249, 2018: 250, 2019: 253,
		2020: 251, 2021: 251, 2022: 251, 2023: 249, 2024: 253, 2025: 252,
		2026: 249, 2027: 251, 2028: 248, 2029: 249, 2030: 252, 2031: 252, 2032: 251,
	}
	for year, want := range counts {
		start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(1, 0, 0)
		if year == 2032 {
			end = c.end
		}
		got, err := c.Count(start, end)
		if err != nil || got != want {
			t.Fatalf("year %d: %d want %d, %v", year, got, want, err)
		}
	}
}

func TestPreviousVerifiedBusinessDay(t *testing.T) {
	c := expanded(t)
	for _, tc := range []struct{ local, want string }{
		{"2016-02-10", "2016-02-05"},
		{"2024-11-21", "2024-11-19"},
		{"2026-09-07", "2026-09-04"},
		{"2026-09-08", "2026-09-04"},
		{"2026-09-09", "2026-09-08"},
		{"2027-01-04", "2026-12-31"},
	} {
		got, err := c.Previous(date(t, tc.local))
		if err != nil || got.Format(time.DateOnly) != tc.want {
			t.Fatalf("%+v: %v %v", tc, got, err)
		}
	}
	for _, local := range []string{"2001-12-31", "2002-01-01", "2002-01-02", "2033-01-01"} {
		if _, err := c.Previous(date(t, local)); err != bond.CalendarOutOfRange {
			t.Fatalf("%s: %v", local, err)
		}
	}
}
