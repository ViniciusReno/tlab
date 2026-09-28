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

## M2.3 — explicit official synchronization (2026-09-10)

`tesouro-lab sync [--data-dir PATH]` now discovers the active official CSV through CKAN, validates and normalizes the download, and imports no-coupon Prefixado quotes into persistent storage. It uses the same data-directory convention as the server. The command prints a JSON report and returns a nonzero exit code on failure. Persistent startup remains free of automatic network requests, and the browser status page explains how to import data with the CLI.

The downloader accepts only the fixed official HTTPS host, including redirects. Metadata and CSV limits are 1 MiB and 32 MiB; each HTTP request has a 60-second timeout, including redirects and body reading. Package identity/state/privacy, resource ownership and uniqueness, HTTP status/content type, and CSV contents are validated before writing quotes. Unsupported instrument rows are counted by their exact source names. A dataset with no supported rows fails explicitly. Duplicate supported bond/date rows within a download are rejected; source corrections across imports update existing records. See the [datasource contract](OFFICIAL_DATASOURCE.md).

Migration `002_sync.sql` adds nullable quote import timestamps and `sync_runs`. Existing quotes survive migration with an unknown import timestamp; no timestamp is invented for legacy or demo data. Each successful run stores the CKAN endpoint, selected resource ID, final download URL, local import timestamp, maximum official quote date across the full validated CSV, counts, and excluded names. `records_read` counts all fully validated source rows; `records_written` counts supported rows upserted, including unchanged records, rather than newly inserted rows. Failed parsing returns no partial batch, so its validated row count is zero. `dataset_max_quote_date` is distinct from import time and includes excluded instruments.

The application records a running attempt before network access. The complete quote batch and success status commit in one transaction. Download, parse, or write failure preserves prior quotes and successful-run metadata; failure status is recorded separately after rollback. Cancellation permits a bounded five-second failure-status write. If that write also fails, the printed report says its status could not be saved. A forced process termination may leave a running record; no background recovery or scheduler is introduced.

The datasource contains no financial formulas, storage accepts normalized data, and the application coordinates the workflow. No dependencies, calendar rules, pricing formulas, or embedded market fixtures changed. Demo services reject sync before accessing the network and continue using their private in-memory databases.

### Validation

Passed locally on macOS/arm64, with final checks on 2026-09-10:

- `go fmt ./...`
- `go vet ./...`
- `go test ./...`
- `go test -race ./...`
- `CGO_ENABLED=0 go build -trimpath -o /private/tmp/tlab-m23-sync ./cmd/tesouro-lab`
- `/private/tmp/tlab-m23-sync analyze prefixado:2015-01-01 --source demo --basis purchase --date 2012-01-03 --yield 8.88`
- `/private/tmp/tlab-m23-sync sync --data-dir /private/tmp/tlab-m23-live.UJqJLJ`
- `git diff --check`

Routine tests use committed official fixtures and in-memory HTTP transports, without live network access. They cover successful discovery/provenance, malformed or ambiguous metadata, unsafe/looping redirects, HTTP status/content type, download size, truncated transfers, cancellation, unsupported-only datasets, duplicate rows, full-dataset date selection, migration from the original schema, NULL preservation, corrections, rollback, persisted status, retries, CLI output/exit codes, and demo isolation. A storage test forces an SQL failure while updating success status, proving that preceding quote updates roll back as part of the same transaction.

Separate manual live checks used only a temporary database. On 2026-09-09, the command read 175,926 source rows and wrote 28,085 Prefixado quotes, with maximum source quote date 2026-09-08. On 2026-09-10, the final binary read 175,984 rows and wrote 28,090 Prefixado quotes, with maximum date 2026-09-09. The source had advanced between checks. A read-only SQLite inspection confirmed 28,090 rows and 28,090 distinct bond/date keys, preserved failure/success reports, and no foreign-key violations. Fixed-fixture tests separately verify identical-input idempotency.

The initial sandboxed live attempt failed before downloading and recorded a failed run; retry with network access succeeded. Go checks also needed cache access outside the sandbox. The compiled offline demo retained scenario PU 774.9918980432597 with 755 business days.

### Remaining scope

M2.4 is next: verified ANBIMA calendar expansion and independent validation of purchase, mark-to-market, and early-redemption contexts. Synchronized analysis, history/freshness UI, IPCA+, Selic, and portfolios remain unavailable. The browser currently shows database presence, while detailed sync reports are printed by the CLI and stored in SQLite. If local storage itself cannot be opened or cannot create the initial run, no durable sync report can be guaranteed. No automated browser interaction, cross-platform execution, remote CI run, or release publication was verified.

## M2.4 — verified calendars and quote contexts (2026-09-10)

Added an embedded, versioned ANBIMA calendar covering 2002-01-01 through 2032-12-31, with 381 extracted holiday dates and per-year source URLs, hashes, and counts. The original demo calendar/version remains unchanged, and its entire overlapping business-day classification matches the expanded fixture. No runtime calendar download or inferred future holiday algorithm was added. See [calendar provenance and rules](../data/calendar/README.md).

The application now uses `pricing.ResolveBasis` to select the exact PU/yield pair and verified settlement/DU. This pure domain function preserves the locked V1 purchase/early-redemption D+1 and mark-to-market D0 rules. Missing context inputs, invalid numeric values, unsupported instruments, out-of-coverage dates, and nonpositive remaining terms return explicit errors. Historical evaluation uses the selected quote date, never the current date. The resolver does not replace official prices with theoretical ones or claim standalone validation for arbitrary imported rows.

Four unmodified official CSV records establish 12 context cases. The 2012 and 2016 quotes validate purchase, base, and early-redemption standalone PUs after truncation within BRL 0.01. The 2024 and 2026 quotes validate purchase/base; recent early-redemption cases fail the locked D+1 validation and remain explicitly recorded as `calculation_not_validated`. The ten validated cases include zero shocks and hypothetical ±100-basis-point shocks with fixed PU and variation expectations. Source, retrieval, raw-line provenance, independent DU, annual contributions, and a 60-digit Python Decimal reference are committed in [pricing testdata](../internal/pricing/testdata/README.md). Python is optional development tooling, not a runtime or Go test requirement.

The recent discrepancy is material: SellPU BRL 765.76 on 2024-11-19 gives BRL 766.14 under standalone D+1; SellPU BRL 490.42 on 2026-09-04 gives BRL 490.68 under D+1. Both match D0 after truncation. The metadata PDF still specifies D+1, while a separate official 2021 announcement introduced D0 redemption under stated conditions. This does not justify silently changing the locked specification or inferring a universal historical switch. The fixture notes preserve the evidence; no tolerance was relaxed, and no accepted scenario expectations were generated for these two failed cases.

### Validation

Passed on macOS/arm64:

- `go fmt ./...`
- `go vet ./...`
- `go test ./...`
- `go test -race ./...`
- `CGO_ENABLED=0 go build -trimpath -o /private/tmp/tlab-m24-contexts ./cmd/tesouro-lab`
- `/private/tmp/tlab-m24-contexts analyze prefixado:2015-01-01 --source demo --basis purchase --date 2012-01-03 --yield 8.88`
- `python3 internal/pricing/testdata/reference.py > /private/tmp/tlab-m24-reference-check.json`
- `cmp internal/pricing/testdata/scenarios.json /private/tmp/tlab-m24-reference-check.json`
- `git diff --check`

Tests cover annual/partial-year totals across the expanded range, original-calendar overlap, Carnival, the November 20 transition, the September 7 weekend boundary, year-end semantics, exact coverage bounds, D+1 reaching/passing maturity, historical matured bonds, missing PU/yield pairs without substitution, non-finite/nonpositive inputs, and unsupported instruments. Financial fixture tests validate magnitude and variation in both shock directions and detect either +1 or -1 DU. The original demo still produces PU 774.9918980432597 with 755 business days and the same calendar/fixture/calculation versions. Go checks required access to the external build cache. ANBIMA and CSV retrievals were separate manual source-verification steps; routine tests remain offline.

### Remaining scope and decision

The expanded calendar and context validation are implemented; recent early-redemption validation is unresolved. Before enabling those synchronized scenarios, reconcile the settlement contract with the maintainer and validate official records around any proposed transition. The V1 specification is unchanged. M2.5 purchase/base delivery and market/history work can proceed independently, keeping affected early-redemption analysis unavailable.

This step adds no market screen or synchronized CLI/API scenario delivery. The demo retains its original data and calendar; the new calendar is embedded and tested for the subsequent synchronized workflow. No cross-platform execution, automated browser interaction, remote CI, or published release was verified.

## M2.5a — local market and quote history (2026-09-15)

Persistent startup now opens a market screen at `/` and `/market`. The default view lists non-matured Prefixado bonds with their latest stored purchase yield/PU, official quote date, and a plain-language fixed-rate explanation. `/market?matured=1` includes matured titles with a maturity status and history links; their old quotes are not presented as current market values.

The application compares each bond's latest quote date with the maximum official date from the most recently completed successful sync, including excluded instruments. States distinguish latest, older, missing purchase inputs, matured, and unknown freshness. A failed or running sync does not replace successful-import metadata. Legacy databases without that metadata show unknown freshness instead of inferring a dataset date from stored Prefixado rows. Quote dates and UTC import timestamps remain separate. The quote list and sync metadata are read within one SQLite snapshot.

Selecting a bond opens `/history?bond=<bond-id>`. History shows up to 100 descending records per page; an exclusive ISO `before` date cursor retrieves older records without duplicating a boundary row. Missing dates and context fields are never filled, substituted, or searched backward for a usable pair. Advanced HTML details expose the distinct buy/sell/base fields, maturity, quote date, provenance, and per-record import timestamp. These views require no JavaScript or network access after import. They cannot switch sources, synchronize, or select a filesystem path.

No financial formulas, calendar rules, dependencies, migrations, or embedded official fixtures changed in this step. Existing SQL quote decoding is shared between exact-date selection, the market list, and history. The original demo and synchronized-scenario unavailable state remain intact.

### Validation

Passed locally on macOS/arm64:

- `go fmt ./...`
- `go vet ./...`
- `go test ./...`
- `go test -race ./...`
- `CGO_ENABLED=0 go build -trimpath -o /private/tmp/tlab-m25-market ./cmd/tesouro-lab`
- `/private/tmp/tlab-m25-market analyze prefixado:2015-01-01 --source demo --basis purchase --date 2012-01-03 --yield 8.88`
- `git diff --check`

New deterministic checks cover empty/legacy databases, bonds without quote rows, missing purchase fields, the full dataset date, successful metadata after a failed sync, local-date maturity boundaries, historical pagination without overlap, malformed inputs, official field rendering, matured-market suppression, and demo source isolation. HTTP tests use in-process handlers and committed official CSV records. Application edge cases use explicitly synthetic test data. The offline CLI result remains PU 774.9918980432597 with 755 business days.

Go cache restrictions and the race suite's loopback bind initially required sandbox escalation; the checks passed after rerunning with the necessary access. The build was also repeated with cache access after a blocked cache-write diagnostic.

### Remaining scope

M2.5b synchronized purchase/base scenario delivery through the shared CLI/browser service is next. Recent early-redemption scenarios remain blocked by the documented settlement discrepancy. Market/history display does not certify standalone pricing validation of imported rows. No automatic dataset-staleness warning, history chart, automated browser interaction, cross-platform execution, remote CI, or release publication was verified or added in this step.

## M2.5b — synchronized purchase/base scenarios (2026-09-27)

The existing in-progress implementation now has passing local validation and updated documentation. Persistent analysis reads stored Prefixado quotes through the same application service used by the demo, CLI, JSON endpoint, and browser. Purchase selects BuyPU/BuyYield/D+1; mark-to-market selects BasePU/SellYield/D0. The pure pricing resolver and anchored formula remain shared, with no new dependency, migration, calendar rule, or financial formula.

An omitted date selects the bond's latest stored record. An explicit date selects only that record; missing dates or required context fields return `missing_quote` without fallback. Historical scenarios retain their selected settlement dates even after maturity. The synchronized calendar is `anbima-2002-2032-v1`; the demo retains its original calendar and private in-memory database.

Market/history links include source, basis, exact quote date, and unrounded yield. Persistent playground pages use ordinary forms and links without JavaScript. Advanced details expose resolved inputs, provenance, import timestamp, calendar/calculation/build versions, JSON, and an equivalent CLI command. `analyze --data-dir PATH` selects a local synchronized database and appears shell-quoted in reproduction commands; HTTP inputs cannot select a filesystem path. Demo analysis rejects this flag. If the optional default shock grid is numerically unavailable, the valid selected result remains visible.

All synchronized early-redemption scenarios remain blocked with `calculation_not_validated` after required quote/term resolution. No transition date is inferred from PU equality or announcements, including for older rows whose individual fixtures pass. Resolving the historical settlement contract remains separate work; the locked specification is unchanged.

### Validation

Passed locally on macOS/arm64:

- `go fmt ./...`
- `go vet ./...`
- `go test ./...`
- `go test -race ./...`
- `CGO_ENABLED=0 go build -trimpath -o /private/tmp/tlab-m25b ./cmd/tesouro-lab`
- `/private/tmp/tlab-m25b analyze prefixado:2015-01-01 --source demo --basis purchase --date 2012-01-03 --yield 8.88`
- `git diff --check`

The initial test run exposed an incorrect overflow expectation: a yield near -100% still produces a finite unit price over the selected short term. The corrected test accepts that finite result and uses a finite maximum gross amount to exercise actual derived-value overflow. No tolerance or production numeric check was relaxed. A separate regression confirms that a newer record advances latest-date selection while leaving fixed-date reproduction unchanged.

Existing in-progress tests now pass for independent purchase/base expectations, CLI/API/HTML parity, source isolation, exact-date and incomplete-pair failures, calendar/maturity limits, generated links, custom-directory quoting, and the optional-grid error. Formatting, race checks, and build were rerun with cache access after sandbox restrictions. The compiled offline demo retains PU 774.9918980432597 and 755 business days.

### Remaining scope

Early-redemption settlement validation and automatic publication-cadence staleness warnings remain unresolved M2 work. No live sync, browser automation, cross-platform execution, remote CI run, or release publication was verified in this step. IPCA+, Selic, and portfolios remain later milestones.

## M2.5c — conservative dataset-age warning (2026-09-27)

The market page now warns when the stored dataset is at least two verified financial-market business days behind the previous business day relative to the user's local calendar date. One business day alone does not trigger a warning. The official resource describes daily publication; using the previous business day without an intraday cutoff is an explicit conservative application assumption. See the [source and policy](OFFICIAL_DATASOURCE.md#dataset-age-warning-policy).

The warning retains the actual dataset maximum quote date and successful-import timestamp, and exposes the expected date and calendar version. Missing or invalid metadata, a future dataset date, or insufficient calendar coverage makes the assessment unavailable. A recent import does not imply fresh quotes; failed syncs do not replace the successful metadata. Per-bond freshness and maturity states remain separate, and no synchronization is triggered by HTTP requests.

`calendar.Previous` reuses the existing verified holiday classifications and fails at coverage boundaries. Neither fixture contents nor scenario counting rules changed. Application logic calculates the lag; HTTP handlers only render the result. A private handler constructor accepts a clock function so rendering tests remain deterministic without a new dependency or public configuration option.

### Validation

Passed locally:

- `go fmt ./...`
- `go test ./internal/web ./internal/app ./internal/calendar ./internal/pricing`
- `go vet ./...`
- `go test ./...`
- `go test -race ./...`
- `git diff --check`

Fixed-date tests cover the warning threshold, same-date quotes, weekends, Carnival, September 7, November 20, year boundaries, local-versus-UTC date differences, future/missing metadata, and calendar bounds. HTTP tests use synthetic quotes and a fixed clock to verify warning text, retained quote/import dates, unavailable states, and failed-sync preservation. The pricing suite revalidated the official context fixtures, including the 2016 Carnival and 2026 September holiday boundaries. Routine tests make no live source requests.

### Remaining scope

The early-redemption settlement discrepancy remains unresolved, and those synchronized calculations stay blocked. The warning is an estimate of local data age, not proof of a source outage or a guarantee that a displayed quote is live. No live sync, automated browser interaction, cross-platform execution, remote CI, or release publication was verified in this step.

## M2 closure — approved morning early-redemption contract (2026-09-28)

The maintainer approved the settlement amendment after reviewing official transition records and requested a checkpoint commit first (`75ea157`). M2 is now implemented under that approved contract. `AGENTS.md` and V1 specification sections 14.3/17.4, together with dependent quote/portfolio documentation, describe D+1 before 2021-09-13 and D0 on/after that date for morning early-redemption scenarios under normal market conditions.

The pure basis resolver selects the convention by quote date. Synchronized Prefixado early-redemption requests then validate the standalone theoretical PU, truncated to cents, against the original official SellPU with a BRL 0.01 tolerance. A mismatch returns `calculation_not_validated`; no alternative settlement, BasePU substitution, holiday exception, or tolerance relaxation is attempted. Accepted requests use the original official SellPU in the unchanged anchored scenario formula.

CLI, JSON, and browser results expose `settlement_convention` and `settlement_version` (`morning-redemption-2021-v1`), alongside the unchanged formula and calendar versions. Market/history provide explicit early-redemption links; the playground labels D0/D+1 and explains the morning/normal-market limitation. Generated commands retain exact source, context, date, and numeric inputs. The original offline demo output and fixture remain unchanged numerically.

### Evidence and validation

Ten official transition records cover five maturities on each of 2021-09-10 and 2021-09-13. Six pass the selected date-dependent convention; four longer-maturity records remain blocked because the current verified calendar does not reproduce their historical PUs. All rows are preserved, including failures. Independent Decimal expectations cover zero shock and both nonzero directions. The existing 2024/2026 official SellPU fixtures now validate using D0 with fixed independent shock expectations. No source CSV or calendar holiday was edited.

Passed locally on macOS/arm64:

- `go fmt ./...`
- `go vet ./...`
- `go test ./...`
- `go test -race ./...`
- `CGO_ENABLED=0 go build -trimpath -o /private/tmp/tlab-m2-closure ./cmd/tesouro-lab`
- `/private/tmp/tlab-m2-closure analyze prefixado:2015-01-01 --source demo --basis purchase --date 2012-01-03 --yield 8.88`
- `git diff --check`

Tests cover date-dependent settlement, missing SellPU without fallback, both shock directions, unchanged-yield anchors, per-record mismatch rejection through service/CLI/API/HTML, context labels and reproduction metadata, D0 maturity/calendar boundaries, validation input/range errors, and the unchanged purchase/base/demo paths. Formatting and build needed cache access after sandbox restrictions.

### Explicit remaining limitations

M2 completion accepts unavailable historical redemption records when validation fails. Reconstructing historical holiday knowledge is not implemented. After-13:00 requests, suspended trading, actual execution, taxes, and fees are not modeled. IPCA+ requires independent instrument-specific evidence in M3; it does not inherit the Prefixado standalone validator. No live sync, automated browser interaction, cross-platform execution, remote CI run, or release publication was performed as part of this closure.
