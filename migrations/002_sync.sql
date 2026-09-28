ALTER TABLE market_quotes ADD COLUMN imported_at TEXT;

CREATE TABLE sync_runs (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    started_at TEXT NOT NULL,
    finished_at TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('running', 'success', 'failed')),
    records_read INTEGER NOT NULL DEFAULT 0,
    records_written INTEGER NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    resource_id TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    imported_at TEXT NOT NULL DEFAULT '',
    dataset_max_quote_date TEXT NOT NULL DEFAULT '',
    unsupported_rows TEXT NOT NULL DEFAULT '{}'
);
