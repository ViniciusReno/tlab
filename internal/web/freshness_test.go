package web

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/app"
	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/storage/sqlite"
)

func TestMarketDatasetAgeWarning(t *testing.T) {
	ctx, dir := context.Background(), t.TempDir()
	store, err := sqlite.OpenPersistent(ctx, dir, assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	s, err := app.OpenPersistent(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	local := time.Date(2026, 9, 10, 23, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	h, err := newHandler(s, func() time.Time { return local })
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, quote, want string }{
		{"empty", "", "Dataset age assessment unavailable (missing_metadata)"},
		{"one-day", "2026-09-08", "No dataset-age warning at this time"},
		{"two-days", "2026-09-04", "Stored dataset appears stale."},
		{"outside-calendar", "2001-12-31", "Dataset age assessment unavailable (calendar_out_of_range)"},
		{"future", "2026-09-11", "Dataset age assessment unavailable (future_quote_date)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.quote != "" {
				qdate, _ := bond.ParseDate(tc.quote)
				maturity, _ := bond.ParseDate("2032-01-01")
				pu, yield := 500.0, 0.12
				q := bond.Quote{Bond: bond.Bond{ID: "prefixado:2032-01-01", Kind: "prefixado", Name: "Synthetic test bond", Maturity: maturity}, Date: qdate, BuyPU: &pu, BuyYield: &yield, Source: "synthetic-test"}
				run := sqlite.SyncRun{ID: tc.name, Source: "synthetic-test", Status: "success", RecordsWritten: 1, ImportedAt: "2026-09-10T12:00:00Z", DatasetMaxQuoteDate: tc.quote}
				if err := store.StartSync(ctx, run); err != nil {
					t.Fatal(err)
				}
				if err := store.CommitSync(ctx, &run, []bond.Quote{q}); err != nil {
					t.Fatal(err)
				}
				// A more recent failed import must not hide the last successful age.
				failed := sqlite.SyncRun{ID: tc.name + "-failed", Source: "synthetic-test", Status: "failed"}
				if err := store.StartSync(ctx, failed); err != nil {
					t.Fatal(err)
				}
				if err := store.FailSync(ctx, failed); err != nil {
					t.Fatal(err)
				}
			}
			r := httptest.NewRecorder()
			h.ServeHTTP(r, httptest.NewRequest("GET", "/market", nil))
			body := r.Body.String()
			if r.Code != 200 || !strings.Contains(body, tc.want) {
				t.Fatalf("%d: %s", r.Code, body)
			}
			if tc.quote != "" && (!strings.Contains(body, "Dataset maximum official quote date: "+tc.quote) || !strings.Contains(body, "2026-09-10T12:00:00Z")) {
				t.Fatal("warning lost actual source/import dates")
			}
			if tc.name == "two-days" && !strings.Contains(body, "2 financial-market business days behind the expected base date 2026-09-09") {
				t.Fatal("incorrect warning threshold or local reference date")
			}
			if tc.name != "two-days" && strings.Contains(body, "Stored dataset appears stale.") {
				t.Fatal("unexpected stale warning")
			}
		})
	}
}
