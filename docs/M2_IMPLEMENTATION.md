# M2: official data, incremental implementation

## M2.1 — persistent storage (2026-09-08)

Running `tesouro-lab` without a subcommand now initializes or reopens a local SQLite database and starts the loopback server. Source equivalent: `go run ./cmd/tesouro-lab`. The empty page explains that no official data has been imported and how to start the separate offline demo.

The default directory is `tesouro-lab` under `os.UserConfigDir`: macOS `~/Library/Application Support`, Linux `$XDG_CONFIG_HOME` or `~/.config`, and Windows `%AppData%`. This is the M2.1 directory convention; the V1 specification did not prescribe an exact path. `--data-dir PATH` overrides it for persistent startup only; relative paths resolve from the working directory. The filename is always `tesouro-lab.db`. `--port` remains available with the default address `127.0.0.1:8080`.

New directories/files request permissions 0700/0600 on Unix; existing permissions are preserved and Windows uses OS access controls. SQLite URI encoding preserves spaces and reserved characters in paths. Both modes share migrations, foreign-key enforcement, and a bounded five-second busy timeout. No schema change or new dependency was needed.

Persistent startup never seeds fixtures. The demo still opens only a private in-memory database and rejects `--data-dir`. Server source is fixed at startup; HTTP source mismatches return `source_mismatch` and cannot switch databases. Closing the server preserves the local database. Invalid paths or database files produce a startup error without replacing the file with demo data.

## Validation

Passed locally on macOS/arm64:

- `go fmt ./...`
- `go vet ./...`
- `go test ./...`
- `go test -race ./...`
- `CGO_ENABLED=0 go build -trimpath -o /private/tmp/tlab-m2-storage ./cmd/tesouro-lab`
- `/private/tmp/tlab-m2-storage analyze prefixado:2015-01-01 --source demo --basis purchase --date 2012-01-03 --yield 8.88`
- `git diff --check`

Some initial checks were blocked by sandbox access to the Go cache; reruns with the required access passed. The binary's demo scenario PU was 774.9918980432597 with 755 business days, preserving the existing fixture expectation.

New deterministic tests cover persisted quotes across reopen, migration idempotency and failed-migration rollback, invalid directories/database preservation, reserved path characters, demo separation, empty HTML without demo controls, source mismatch, unavailable synchronized analysis, and CLI startup/graceful shutdown/reopen with a temporary data directory. HTTP checks use in-process handlers; the CLI lifecycle test binds loopback on an available port.

## Remaining M2 work

This step does not implement CKAN access, synchronization, synchronized scenarios, market history, freshness, or expanded calendars. `sync` and `analyze --source synced` remain explicitly unavailable. A database already containing quotes shows a status page; browsing is still pending. No browser interaction or cross-platform execution was verified in this step. CI and release publication remain unverified.

The subsequent M2.2 checkpoint is recorded below. See [next steps](NEXT_STEPS.md).

## M2.2 — official datasource contract (2026-09-08)

Verified the official CKAN API, metadata PDF, and downloaded CSV. Added unmodified API metadata and a ten-row raw CSV extract with hashes, line numbers, provenance, and source license attribution. See the [contract](OFFICIAL_DATASOURCE.md) and [fixture notes](../internal/datasource/tesouro/testdata/README.md).

`ParseDataset` reads up to 32 MiB, validates source formatting, returns no-coupon Prefixado quotes, and counts excluded instrument names. The strict demo parser retains its 1 MiB limit and embedded data. No dependencies, database migrations, financial formulas, or calendar rules changed.

A one-off local check of the downloaded 14,474,206-byte CSV passed: 28,080 accepted Prefixado rows and 147,788 excluded rows. The temporary check was removed; routine tests use only committed fixtures and synthetic edge cases. The CKAN historical row and the demo methodology example differ, so their provenance and values remain separate. This is ingestion validation, not independent scenario validation.

The next bounded step is M2.3 synchronization. CKAN discovery/download, sync status, and transactional integration remain pending; `sync` and synchronized analysis are still unavailable.


M2.2 final validation on 2026-09-09 (macOS/arm64): `go fmt ./...`, `go vet ./...`, `go test ./...`, `go test -race ./...`, `CGO_ENABLED=0 go build -trimpath -o /private/tmp/tlab-m22-parser ./cmd/tesouro-lab`, and `git diff --check`. Sandbox restrictions initially blocked Go cache access and the existing CLI loopback test; the required checks were rerun with access. Cross-platform execution and remote CI were not verified.
