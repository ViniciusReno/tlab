package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
	"github.com/ViniciusReno/tlab/internal/storage/sqlite"
)

// Sync is explicit and permitted only for a persistent service. The client uses
// fixed official endpoints; its replaceable transport allows offline I/O tests.
func (s *Service) Sync(ctx context.Context, client tesouro.Client) (sqlite.SyncRun, error) {
	if s.source != "synced" {
		return sqlite.SyncRun{}, bond.SourceMismatch
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return sqlite.SyncRun{}, err
	}
	run := sqlite.SyncRun{ID: hex.EncodeToString(id[:]), Source: tesouro.PackageURL,
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano), Status: "running", Unsupported: map[string]int{}}
	if err := s.store.StartSync(ctx, run); err != nil {
		return sqlite.SyncRun{}, bond.SyncFailed
	}
	data, err := client.Fetch(ctx)
	run.ResourceID, run.SourceURL = data.ResourceID, data.SourceURL
	run.RecordsRead = data.RecordsRead
	if data.Unsupported != nil {
		run.Unsupported = data.Unsupported
	}
	if err == nil {
		run.ImportedAt = time.Now().UTC().Format(time.RFC3339Nano)
		run.DatasetMaxQuoteDate = data.MaxQuoteDate.Format(time.DateOnly)
		for i := range data.Quotes {
			data.Quotes[i].ImportedAt = run.ImportedAt
		}
		run.Status, run.RecordsWritten = "success", len(data.Quotes)
		if err = s.store.CommitSync(ctx, &run, data.Quotes); err != nil {
			run.ErrorMessage = "Local database write failed; the quote transaction was rolled back."
		}
	} else {
		run.ErrorMessage = err.Error()
	}
	if err == nil {
		return run, nil
	}
	run.Status, run.RecordsWritten = "failed", 0
	run.ImportedAt, run.DatasetMaxQuoteDate = "", ""
	run.FinishedAt = time.Now().UTC().Format(time.RFC3339Nano)
	// Ctrl+C cancels the download/write, but still permits a bounded status write.
	statusCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := s.store.FailSync(statusCtx, run); err != nil {
		run.ErrorMessage += " The failure status could not be saved; its stored run may remain running."
	}
	return run, bond.SyncFailed
}
