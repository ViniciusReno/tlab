package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
)

// SyncRun separates local import timestamps from the source's official market dates.
type SyncRun struct {
	ID                  string         `json:"id"`
	Source              string         `json:"source"`
	StartedAt           string         `json:"started_at"`
	FinishedAt          string         `json:"finished_at"`
	Status              string         `json:"status"`
	RecordsRead         int            `json:"records_read"`
	RecordsWritten      int            `json:"records_written"`
	ErrorMessage        string         `json:"error_message,omitempty"`
	ResourceID          string         `json:"resource_id,omitempty"`
	SourceURL           string         `json:"source_url,omitempty"`
	ImportedAt          string         `json:"imported_at,omitempty"`
	DatasetMaxQuoteDate string         `json:"dataset_max_quote_date,omitempty"`
	Unsupported         map[string]int `json:"unsupported_rows"`
}

func (s *Store) StartSync(ctx context.Context, run SyncRun) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO sync_runs(id,source,started_at,status) VALUES (?,?,?,'running')`, run.ID, run.Source, run.StartedAt)
	return err
}

// CommitSync commits the full normalized batch and successful status together.
// Failure leaves both old quotes and the original running record untouched.
func (s *Store) CommitSync(ctx context.Context, run *SyncRun, quotes []bond.Quote) error {
	if run.Status != "success" || len(quotes) == 0 || run.RecordsWritten != len(quotes) || run.ImportedAt == "" || run.DatasetMaxQuoteDate == "" {
		return bond.InvalidInput
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := upsert(ctx, tx, quotes); err != nil {
		return err
	}
	run.FinishedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err := finishSync(ctx, tx, *run); err != nil {
		return err
	}
	return tx.Commit()
}

// FailSync runs separately after any normalized write transaction has rolled back.
func (s *Store) FailSync(ctx context.Context, run SyncRun) error {
	if run.Status != "failed" || run.RecordsWritten != 0 || run.ImportedAt != "" {
		return bond.InvalidInput
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := finishSync(ctx, tx, run); err != nil {
		return err
	}
	return tx.Commit()
}

func finishSync(ctx context.Context, tx *sql.Tx, run SyncRun) error {
	excluded, err := json.Marshal(run.Unsupported)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE sync_runs SET finished_at=?,status=?,records_read=?,records_written=?,error_message=?,
	resource_id=?,source_url=?,imported_at=?,dataset_max_quote_date=?,unsupported_rows=? WHERE id=? AND status='running'`,
		run.FinishedAt, run.Status, run.RecordsRead, run.RecordsWritten, run.ErrorMessage, run.ResourceID,
		run.SourceURL, run.ImportedAt, run.DatasetMaxQuoteDate, string(excluded), run.ID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("sync run is missing or already finished")
	}
	return nil
}

func (s *Store) SyncRun(ctx context.Context, id string) (SyncRun, error) {
	var run SyncRun
	var excluded string
	err := s.db.QueryRowContext(ctx, `SELECT id,source,started_at,finished_at,status,records_read,records_written,error_message,
	resource_id,source_url,imported_at,dataset_max_quote_date,unsupported_rows FROM sync_runs WHERE id=?`, id).Scan(
		&run.ID, &run.Source, &run.StartedAt, &run.FinishedAt, &run.Status, &run.RecordsRead, &run.RecordsWritten,
		&run.ErrorMessage, &run.ResourceID, &run.SourceURL, &run.ImportedAt, &run.DatasetMaxQuoteDate, &excluded)
	if err != nil {
		return run, err
	}
	err = json.Unmarshal([]byte(excluded), &run.Unsupported)
	return run, err
}
