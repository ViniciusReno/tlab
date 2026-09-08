package calendar

import (
	"strings"
	"testing"
	"time"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/bond"
)

func fixture(t *testing.T) *Calendar {
	t.Helper()
	f, err := assets.Files.Open("data/demo/calendar.json")
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
func date(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := bond.ParseDate(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestOfficialPricingDayCount(t *testing.T) {
	c := fixture(t)
	for _, tc := range []struct {
		start, end string
		want       int
	}{
		{"2012-01-04", "2015-01-01", 755},
		{"2012-01-03", "2015-01-01", 756},
		{"2012-02-17", "2012-02-23", 2},
		{"2012-12-31", "2013-01-02", 1},
	} {
		got, err := c.Count(date(t, tc.start), date(t, tc.end))
		if err != nil || got != tc.want {
			t.Fatalf("%+v: count=%d err=%v", tc, got, err)
		}
	}
}

func TestSettlementAndBounds(t *testing.T) {
	c := fixture(t)
	for _, tc := range []struct{ start, want string }{
		{"2012-01-03", "2012-01-04"},
		{"2012-02-17", "2012-02-22"},
		{"2012-04-05", "2012-04-09"},
		{"2014-12-31", "2015-01-02"},
	} {
		got, err := c.Next(date(t, tc.start))
		if err != nil || got.Format(time.DateOnly) != tc.want {
			t.Fatalf("%+v: %v %v", tc, got, err)
		}
	}
	for _, s := range []string{"2011-12-31", "2015-12-31", "2016-01-01"} {
		if _, err := c.Next(date(t, s)); err != bond.CalendarOutOfRange {
			t.Fatalf("%s: %v", s, err)
		}
	}
	for _, tc := range []struct {
		start, end string
		want       error
	}{
		{"2015-01-01", "2015-01-01", bond.NoRemainingTerm},
		{"2015-01-02", "2015-01-01", bond.NoRemainingTerm},
		{"2012-02-18", "2012-02-22", bond.NoRemainingTerm},
		{"2012-01-04", "2016-01-01", bond.CalendarOutOfRange},
		{"2011-12-31", "2015-01-01", bond.CalendarOutOfRange},
	} {
		if _, err := c.Count(date(t, tc.start), date(t, tc.end)); err != tc.want {
			t.Fatalf("%+v: %v", tc, err)
		}
	}
}

func TestInvalidCalendar(t *testing.T) {
	for _, s := range []string{`{}`, `{"version":"x","start":"2015-01-01","end":"2014-01-01"}`,
		`{"version":"x","start":"2012-01-01","end":"2012-12-31","holidays":["2013-01-01"]}`,
		`{"version":"x","start":"2012-01-01","end":"2012-12-31","holidays":["2012-01-01","2012-01-01"]}`} {
		if _, err := Load(strings.NewReader(s)); err == nil {
			t.Fatalf("accepted %s", s)
		}
	}
}
