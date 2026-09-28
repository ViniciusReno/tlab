package app

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/storage/sqlite"
)

func TestMarketFreshnessMaturityAndHistory(t *testing.T) {
	ctx := context.Background()
	s, err := OpenPersistent(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	date := func(s string) time.Time {
		d, err := bond.ParseDate(s)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	now := time.Date(2026, 9, 15, 23, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	if m, err := s.Market(ctx, now, false); err != nil || len(m.Quotes) != 0 || m.ImportedAt != "" {
		t.Fatalf("empty: %+v %v", m, err)
	}
	pu, yield := 500.0, 0.12
	quotes := []bond.Quote{}
	for i, maturity := range []string{"2026-09-15", "2026-09-16", "2030-01-01", "2031-01-01", "2032-01-01"} {
		quotes = append(quotes, bond.Quote{Bond: bond.Bond{ID: "prefixado:" + maturity, Kind: "prefixado", Name: "Synthetic test bond", Maturity: date(maturity)}, Date: date("2026-09-14"), BuyPU: &pu, BuyYield: &yield, Source: "synthetic-test"})
		if i == 2 {
			quotes[i].Date = date("2026-09-11")
		}
		if i == 3 {
			quotes[i].BuyPU = nil
		}
	}
	if err := s.store.Upsert(ctx, quotes); err != nil {
		t.Fatal(err)
	}
	m, err := s.Market(ctx, now, false)
	if err != nil || len(m.Quotes) != 4 || m.Quotes[0].Status != "freshness_unknown" {
		t.Fatalf("legacy/local date: %+v %v", m, err)
	}
	// The source maximum includes excluded instruments, not just imported Prefixado.
	run := sqlite.SyncRun{ID: "success", Source: "synthetic-test", StartedAt: "2026-09-15T10:00:00Z", Status: "success", RecordsWritten: len(quotes), ImportedAt: "2026-09-15T10:01:00Z", DatasetMaxQuoteDate: "2026-09-15"}
	if err := s.store.StartSync(ctx, run); err != nil {
		t.Fatal(err)
	}
	if err := s.store.CommitSync(ctx, &run, quotes); err != nil {
		t.Fatal(err)
	}
	m, err = s.Market(ctx, now, false)
	if err != nil || m.Quotes[0].Status != "older_quote" || m.DatasetMaxQuoteDate != "2026-09-15" {
		t.Fatalf("full dataset date: %+v %v", m, err)
	}
	run.ID, run.DatasetMaxQuoteDate = "new-success", "2026-09-14"
	if err := s.store.StartSync(ctx, run); err != nil {
		t.Fatal(err)
	}
	if err := s.store.CommitSync(ctx, &run, quotes); err != nil {
		t.Fatal(err)
	}
	failed := sqlite.SyncRun{ID: "failed", Source: "synthetic-test", Status: "failed", FinishedAt: "2099-01-01T00:00:00Z"}
	if err := s.store.StartSync(ctx, failed); err != nil {
		t.Fatal(err)
	}
	if err := s.store.FailSync(ctx, failed); err != nil {
		t.Fatal(err)
	}
	m, err = s.Market(ctx, now, true)
	if err != nil || len(m.Quotes) != 5 || m.DatasetMaxQuoteDate != run.DatasetMaxQuoteDate || m.ImportedAt != run.ImportedAt {
		t.Fatalf("metadata: %+v %v", m, err)
	}
	for i, want := range []string{"matured", "latest", "older_quote", "no_quote", "latest"} {
		if m.Quotes[i].Status != want {
			t.Fatalf("row %d: got %s, want %s", i, m.Quotes[i].Status, want)
		}
	}
	if m.Quotes[2].Date.Format(time.DateOnly) != "2026-09-11" || math.Abs(*m.Quotes[2].BuyPU-500) > 1e-9 {
		t.Fatal("older quote was substituted")
	}

	// More than one page, with missing context in the newest record.
	id := quotes[4].Bond.ID
	history := []bond.Quote{}
	for i := 0; i < 102; i++ {
		q := quotes[4]
		q.Date = date("2026-01-01").AddDate(0, 0, i)
		history = append(history, q)
	}
	quotes[4].BuyPU = nil
	history = append(history, quotes[4])
	if err := s.store.Upsert(ctx, history); err != nil {
		t.Fatal(err)
	}
	h, err := s.History(ctx, id, "")
	if err != nil || len(h.Quotes) != 100 || h.Before == "" || h.Quotes[0].BuyPU != nil {
		t.Fatalf("first page: %+v %v", h, err)
	}
	next, err := s.History(ctx, id, h.Before)
	if err != nil || len(next.Quotes) != 3 || next.Before != "" || !next.Quotes[0].Date.Before(h.Quotes[99].Date) {
		t.Fatalf("second page: %+v %v", next, err)
	}
	for _, tc := range []struct {
		id, before string
		want       error
	}{
		{"", "", bond.InvalidInput}, {id, "2026-02-30", bond.InvalidInput}, {"unknown", "", bond.MissingQuote}, {id, "2000-01-01", bond.MissingQuote},
	} {
		if _, err := s.History(ctx, tc.id, tc.before); err != tc.want {
			t.Fatalf("history %s: %v", fmt.Sprint(tc), err)
		}
	}
	demo, err := OpenDemo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer demo.Close()
	if _, err := demo.Market(ctx, now, true); err != bond.SourceMismatch {
		t.Fatalf("demo market: %v", err)
	}
	if _, err := demo.History(ctx, id, ""); err != bond.SourceMismatch {
		t.Fatalf("demo history: %v", err)
	}
}
