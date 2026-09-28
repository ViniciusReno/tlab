package sqlite

import (
	"context"
	"math"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
)

func TestSyncAtomicStatusRollbackAndReopen(t *testing.T) {
	ctx, dir := context.Background(), t.TempDir()
	s, err := OpenPersistent(ctx, dir, assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { s.Close() }()
	body, err := assets.Files.ReadFile("data/demo/prefixado.csv")
	if err != nil {
		t.Fatal(err)
	}
	quotes, err := tesouro.Parse(strings.NewReader(string(body)), "storage-test")
	if err != nil {
		t.Fatal(err)
	}
	run := SyncRun{ID: "first", Source: tesouro.PackageURL, StartedAt: "2026-09-09T12:00:00Z", FinishedAt: "2026-09-09T12:00:01Z",
		Status: "success", RecordsRead: 1, RecordsWritten: 1, ImportedAt: "2026-09-09T12:00:01Z", DatasetMaxQuoteDate: "2012-01-03", Unsupported: map[string]int{}}
	quotes[0].ImportedAt = run.ImportedAt
	for _, id := range []string{"first", "repeat"} {
		run.ID = id
		if err := s.StartSync(ctx, run); err != nil {
			t.Fatal(err)
		}
		if err := s.CommitSync(ctx, &run, quotes); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := s.db.QueryRow("SELECT count(*) FROM market_quotes").Scan(&count); err != nil || count != 1 {
		t.Fatalf("duplicates: %d %v", count, err)
	}
	pu := 800.0
	changed := quotes[0]
	changed.BuyPU = &pu
	run.ID = "rollback"
	if err := s.StartSync(ctx, run); err != nil {
		t.Fatal(err)
	}
	// Force an actual SQL failure after the quote updates, while recording success.
	if _, err := s.db.Exec(`CREATE TRIGGER reject_success BEFORE UPDATE ON sync_runs WHEN NEW.status='success' BEGIN SELECT RAISE(ABORT,'test failure'); END`); err != nil {
		t.Fatal(err)
	}
	if err := s.CommitSync(ctx, &run, []bond.Quote{changed}); err == nil {
		t.Fatal("accepted failed status transaction")
	}
	stored, err := s.Quote(ctx, quotes[0].Bond.ID, "2012-01-03")
	if err != nil || math.Abs(*stored.BuyPU-733.86) > 1e-9 {
		t.Fatalf("rollback lost original quote: %+v %v", stored, err)
	}
	status, err := s.SyncRun(ctx, run.ID)
	if err != nil || status.Status != "running" || status.RecordsWritten != 0 {
		t.Fatalf("status escaped transaction: %+v %v", status, err)
	}
	run.Status, run.RecordsWritten, run.ImportedAt, run.DatasetMaxQuoteDate = "failed", 0, "", ""
	run.ErrorMessage = "Local database write failed."
	if err := s.FailSync(ctx, run); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec("DROP TRIGGER reject_success"); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = OpenPersistent(ctx, dir, assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	status, err = s.SyncRun(ctx, "rollback")
	if err != nil || status.Status != "failed" || status.ErrorMessage == "" || status.ImportedAt != "" {
		t.Fatalf("failed status not persisted: %+v %v", status, err)
	}
	status, err = s.SyncRun(ctx, "first")
	if err != nil || status.Status != "success" || status.DatasetMaxQuoteDate != "2012-01-03" {
		t.Fatalf("lost prior successful sync: %+v %v", status, err)
	}
	stored, err = s.Quote(ctx, quotes[0].Bond.ID, "2012-01-03")
	if err != nil || stored.ImportedAt != "2026-09-09T12:00:01Z" {
		t.Fatalf("lost import provenance: %+v %v", stored, err)
	}
	// A source correction replaces the same natural key; missing values stay NULL.
	run.ID, run.Status, run.RecordsWritten, run.ImportedAt, run.DatasetMaxQuoteDate = "correction", "success", 1, "2026-09-10T12:00:00Z", "2012-01-03"
	run.ErrorMessage = ""
	changed.ImportedAt, changed.BuyYield = run.ImportedAt, nil
	if err := s.StartSync(ctx, run); err != nil {
		t.Fatal(err)
	}
	if err := s.CommitSync(ctx, &run, []bond.Quote{changed}); err != nil {
		t.Fatal(err)
	}
	stored, err = s.Quote(ctx, changed.Bond.ID, "")
	if err != nil || stored.BuyYield != nil || math.Abs(*stored.BuyPU-800) > 1e-9 || stored.ImportedAt != run.ImportedAt {
		t.Fatalf("correction failed: %+v %v", stored, err)
	}
}

func TestSyncMigrationPreservesLegacyQuoteWithoutInventingImportDate(t *testing.T) {
	ctx, dir := context.Background(), t.TempDir()
	initial, err := assets.Files.ReadFile("migrations/001_initial.sql")
	if err != nil {
		t.Fatal(err)
	}
	legacy := fstest.MapFS{"migrations/001_initial.sql": &fstest.MapFile{Data: initial}}
	s, err := OpenPersistent(ctx, dir, legacy)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.db.Exec(`INSERT INTO bonds VALUES ('prefixado:2015-01-01','prefixado','Tesouro Prefixado','2015-01-01');
	INSERT INTO market_quotes (bond_id,quote_date,buy_pu,source) VALUES ('prefixado:2015-01-01','2012-01-03',733.86,'legacy');`)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = OpenPersistent(ctx, dir, assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	q, err := s.Quote(ctx, "prefixado:2015-01-01", "2012-01-03")
	if err != nil || q.Source != "legacy" || q.ImportedAt != "" || q.Date.Format(time.DateOnly) != "2012-01-03" || math.Abs(*q.BuyPU-733.86) > 1e-9 {
		t.Fatalf("legacy quote changed: %+v %v", q, err)
	}
}
