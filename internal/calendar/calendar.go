// Package calendar counts business days against an explicitly bounded holiday fixture.
package calendar

import (
	"encoding/json"
	"io"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
)

type Calendar struct {
	Version    string
	start, end time.Time
	holidays   map[string]bool
}

func Load(r io.Reader) (*Calendar, error) {
	var data struct {
		Version, Start, End string
		Holidays            []string
	}
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return nil, err
	}
	start, err := bond.ParseDate(data.Start)
	if err != nil {
		return nil, err
	}
	end, err := bond.ParseDate(data.End)
	if err != nil {
		return nil, err
	}
	if data.Version == "" || end.Before(start) {
		return nil, bond.InvalidInput
	}
	c := &Calendar{Version: data.Version, start: start, end: end, holidays: make(map[string]bool)}
	for _, s := range data.Holidays {
		d, err := bond.ParseDate(s)
		if err != nil || !c.covered(d) || c.holidays[s] {
			return nil, bond.InvalidInput
		}
		c.holidays[s] = true
	}
	return c, nil
}

func day(d time.Time) time.Time              { return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC) }
func (c *Calendar) covered(d time.Time) bool { return !d.Before(c.start) && !d.After(c.end) }
func (c *Calendar) business(d time.Time) bool {
	return d.Weekday() != time.Saturday && d.Weekday() != time.Sunday && !c.holidays[d.Format(time.DateOnly)]
}

func (c *Calendar) Next(d time.Time) (time.Time, error) {
	d = day(d)
	if !c.covered(d) {
		return time.Time{}, bond.CalendarOutOfRange
	}
	for {
		d = d.AddDate(0, 0, 1)
		if !c.covered(d) {
			return time.Time{}, bond.CalendarOutOfRange
		}
		if c.business(d) {
			return d, nil
		}
	}
}

// Count includes settlement and excludes maturity. Both endpoints must be covered.
func (c *Calendar) Count(settlement, maturity time.Time) (int, error) {
	settlement, maturity = day(settlement), day(maturity)
	if !c.covered(settlement) || !c.covered(maturity) {
		return 0, bond.CalendarOutOfRange
	}
	if !settlement.Before(maturity) {
		return 0, bond.NoRemainingTerm
	}
	count := 0
	for d := settlement; d.Before(maturity); d = d.AddDate(0, 0, 1) {
		if c.business(d) {
			count++
		}
	}
	if count == 0 {
		return 0, bond.NoRemainingTerm
	}
	return count, nil
}
