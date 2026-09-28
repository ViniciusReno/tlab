# Tesouro Lab

A local educational lab for Brazilian government bonds, built around official data, transparent calculations, and hypothetical scenarios.

**Status: M1 and M2 implemented.** Official synchronization, market/history, dataset-age warnings, and Prefixado scenarios are available. Morning early-redemption scenarios use the approved historical settlement rule and per-record PU validation; mismatching rows remain explicitly unavailable. Run from source or build a local executable. No downloadable release has been published; M3–M5 remain planned.

Tesouro Lab helps beginners understand bond prices and yield changes while letting technical readers inspect the same inputs, formulas, calendars, and results. It does not recommend investments, predict yields, or execute transactions.

## Try the application

With Go 1.25 or newer installed, run from the repository root:

```sh
go run ./cmd/tesouro-lab demo
```

The first build may download Go dependencies. Open `http://127.0.0.1:8080` and stop with Ctrl+C. No Docker, Node/npm, configuration file, or separate database process is required.

To produce a standalone executable with embedded web assets, migrations, and fixtures:

```sh
go build ./cmd/tesouro-lab
```

Then start the compiled application:

Linux/macOS:

```sh
./tesouro-lab demo
```

Windows PowerShell:

```powershell
.\tesouro-lab.exe demo
```

Open `http://127.0.0.1:8080`. Stop the application with Ctrl+C.

The compiled demo works offline without Go. Demo changes are temporary: each process uses a private in-memory SQLite database and never opens your persistent portfolio or synchronized data. Use `demo --port 8081` if port 8080 is occupied.

## Persistent local storage

Run `go run ./cmd/tesouro-lab` (or `./tesouro-lab` after building) to create/open the local database and start the server. Persistent mode opens the market screen with locally stored quotes, or an empty state before import. Purchase, official base PU, and validated morning early-redemption scenarios use the stored quotes. No demo quotes are copied into persistent storage.

The database filename is `tesouro-lab.db`. The default directory follows Go's `os.UserConfigDir` convention:

- macOS: `~/Library/Application Support/tesouro-lab`.
- Linux: `$XDG_CONFIG_HOME/tesouro-lab`, or `~/.config/tesouro-lab` when unset.
- Windows: `%AppData%/tesouro-lab`.

Override it locally with `go run ./cmd/tesouro-lab --data-dir ./local-data --port 8081`. Relative paths resolve from the working directory. Stop with Ctrl+C; the database stays on disk. The demo rejects `--data-dir` and always uses private memory. HTTP parameters cannot select a data directory or switch sources.

New directories/files request owner-only permissions on Unix; Windows access follows OS permissions. Existing directories and files retain their permissions. See [M2 implementation notes](docs/M2_IMPLEMENTATION.md).

## Synchronize official data

Run an explicit import from Tesouro Transparente:

```sh
go run ./cmd/tesouro-lab sync
```

For a custom database, use `go run ./cmd/tesouro-lab sync --data-dir ./local-data`, and start the server with the same `--data-dir`. The compiled equivalent is `./tesouro-lab sync` (`.\tesouro-lab.exe sync` in Windows PowerShell).

The command discovers the official CSV through CKAN, imports no-coupon Prefixado quotes, and prints a JSON report with row counts, excluded instrument names, resource provenance, the local import timestamp, and `dataset_max_quote_date` from official `Data Base` values. This date covers the full validated CSV, including excluded rows; it is not the download date. Other instruments remain excluded until their milestone.

Repeated imports update the same bond/date records without duplicates. Missing fields remain unavailable. Failed downloads, parsing, or database writes preserve prior quotes and produce a failed report and nonzero exit code. Run status is stored separately in `sync_runs`; no background sync or HTTP sync endpoint exists. Demo mode remains isolated and offline.

Downloads require internet access and are limited to the official HTTPS host, 1 MiB of metadata, 32 MiB of CSV, and 60 seconds per HTTP request including redirects and body reading. Reload the market screen after importing. M2.4 adds a verified 2002–2032 calendar and independent context fixtures; the demo retains its original 2012–2015 calendar. Morning early-redemption scenarios use D+1 before 2021-09-13 and D0 thereafter, with per-record validation. Historical calendar discrepancies can leave individual scenarios unavailable. See the [validation evidence](internal/pricing/testdata/README.md).

### Browse market and history

Open `/market` (also the persistent server's home page). The default list shows non-matured Prefixado bonds with the latest stored purchase quote and its official date. States distinguish the dataset's latest date, older quotes, missing purchase fields, and unknown freshness when successful-import metadata is absent. Import time is displayed separately and never treated as a market date.

Choose **Include matured bonds** to find historical titles. Matured rows show their maturity date and direct you to history instead of displaying a current price. Select a bond to view `/history?bond=<bond-id>`, with 100 records per page and an **Older records** link. Missing dates and fields are never filled. Advanced details expose the distinct buy/sell/base fields and each record's source/import timestamp. Both views work offline after synchronization and without JavaScript. Purchase, official base PU, and early-redemption scenario links preserve the exact quote date and context, including historical scenarios for matured bonds.

### Dataset age

The market page warns when the stored dataset is at least two financial-market business days behind the last business day before the user's local date. A lag of one business day does not trigger a warning. This conservative daily-publication policy assumes no intraday cutoff or guaranteed publication time. Weekends and verified ANBIMA holidays do not increase the lag.

The warning shows the actual dataset quote date, import timestamp, and expected base date separately. Missing metadata, future source dates, or dates outside the verified calendar produce an unavailable assessment. A recent import alone cannot clear an old dataset warning, and no automatic sync runs. See [policy and source](docs/OFFICIAL_DATASOURCE.md#dataset-age-warning-policy).

## Run from source

The browser and CLI use the same calculation. Reproduce a hypothetical scenario from the official historical example:

```sh
go run ./cmd/tesouro-lab analyze prefixado:2015-01-01 --source demo --basis purchase --date 2012-01-03 --yield 8.88 --amount 10000
```

This changes the example's annual yield from 10.88% to a hypothetical 8.88%, producing a scenario unit price of approximately BRL 774.99 from the official BRL 733.86 baseline. These are not current market values. See [fixture provenance and assumptions](data/demo/README.md).

Docker may become an optional convenience later. It is not the primary installation path or an M1 dependency.

## V1 scope

The table describes the target V1, not today's complete feature set. Prefixado demo scenarios and synchronized market/history browsing are implemented; IPCA+, Selic, and portfolios are not yet available. Early-redemption rows that fail validation remain unavailable.

| Instrument | Supported behavior |
|---|---|
| Tesouro Prefixado, no coupon | Market/history, portfolio valuation, yield scenarios, and gross hold/early-exit comparison |
| Tesouro IPCA+, no coupon | Market/history, portfolio valuation, and same-date real-yield scenarios |
| Tesouro Selic | Official market/history and portfolio valuation only |

Coupon-paying bonds, RendA+, Educa+, other asset classes, investment recommendations, forecasts, brokerage integrations, accounts, cloud sync, and AI features are outside V1.

Simple mode explains results in plain language. Advanced mode exposes the inputs, quote/settlement dates, assumptions, formulas, and provenance behind the same calculation. Missing information is shown as unavailable, never invented or treated as zero.

## Repository and technology

This is a monorepo with one root Go module and one application release lifecycle. Backend, financial calculations, CLI, templates, CSS/JavaScript, migrations, fixtures, tests, and documentation live together.

The V1 stack is Go, `net/http`, `html/template`, `database/sql`, a pure-Go SQLite driver, server-rendered HTML, CSS, and minimal vanilla JavaScript. Assets and demo data are embedded in the binary. The default bind address is `127.0.0.1:8080`.

All authored repository content and the application interface are in English. Official instrument names and raw source fields remain unchanged for provenance, with English explanations. Displayed values use explicit BRL currency, decimal points, and ISO dates.

## Commands and availability

| Command | Purpose |
|---|---|
| `tesouro-lab [--data-dir PATH] [--port 8080]` | Start local persistent storage; show an empty state before data is imported |
| `tesouro-lab demo` | Start the isolated offline demo |
| `tesouro-lab sync [--data-dir PATH]` | Import official Prefixado quotes and print a JSON success/failure report |
| `tesouro-lab analyze <bond-id> --source demo|synced --yield <percent>` | Run the same scenario calculation used by the browser |

`analyze` accepts `--source demo|synced`, `--basis purchase|mark_to_market|early_exit`, `--date YYYY-MM-DD`, and an optional `--amount` in BRL. For example, `--yield 12.00` means 12% per year. The source defaults to `synced`, reading existing local quotes without syncing. Use `--data-dir PATH` for a custom synchronized database; demo analysis rejects this flag. The default context is purchase and the default date is the latest available quote for the bond. The resolved quote date is reported. An explicit missing date or incomplete context returns `missing_quote`, without searching backward. The demo fixture has no base/sell quotes; those contexts return `missing_quote`. Synchronized early redemption validates the selected SellPU/yield against the standalone theoretical PU (truncated to two decimals, tolerance BRL 0.01), then uses the official SellPU for the anchored scenario. A mismatch returns `calculation_not_validated` without substituting another quote or calendar. Analysis never triggers synchronization automatically.

After synchronization, reproduce a hypothetical scenario using an exact historical quote (the record must exist locally):

```sh
go run ./cmd/tesouro-lab analyze prefixado:2032-01-01 --source synced --basis mark_to_market --date 2026-09-04 --yield 13.43 --amount 10000
```

The browser's advanced details generate the equivalent command, including the local data directory. Generated URLs contain the resolved scenario inputs and cannot select a database. Synchronized pages use ordinary forms and links without requiring JavaScript.

Reproducing a result requires the same quote contents and calendar/calculation versions. Advanced results provide an equivalent CLI command with explicit inputs. Official source corrections can change historical results; see the [CLI contract](docs/V1_SPEC.md#24-cli).

## Official quote contexts

The source dataset distinguishes three unit prices (PU):

| Source field | Meaning | Settlement |
|---|---|---|
| `PU Compra Manha` | Purchase quote | Next financial-market business day, D+1 |
| `PU Venda Manha` | Morning early-redemption quote | D+1 before 2021-09-13; D0 on/after that date |
| `PU Base Manha` | Portfolio mark-to-market valuation | Quote date, D0 |

Portfolio valuation uses the official base unit price. Hypothetical early redemption uses the official sell unit price. Every current-looking value includes its official quote date; the import timestamp is not a market date. Portfolio and early-exit values are gross.

The early-redemption model covers morning quotes under normal market conditions. It does not model after-13:00 requests, suspended trading, or actual execution. The [approved settlement decision](docs/M2_SETTLEMENT_DECISION.md) documents the evidence and historical-calendar limitations. Results expose the settlement convention and version separately from the unchanged pricing formula.

Source metadata, formulas, and limitations are documented in the [V1 specification](docs/V1_SPEC.md).

## Implementation milestones

| Milestone | Demonstrable outcome |
|---|---|
| M1 — Implemented Prefixado demo | One command starts an offline demo; a user can change a yield and reproduce the result through the CLI |
| M2 — Implemented official data and dates | Idempotent synchronization, history, explicit quote contexts, and accurate freshness states |
| M3 — IPCA+ and Selic | Validated real-yield scenarios for IPCA+; explicitly limited Selic display/valuation |
| M4 — Local portfolio | Quantity-only positions, optional acquisition cost, and gross Prefixado hold/early-exit comparison |
| M5 — V1 completion | Market, Playground, Portfolio, and Learn screens; accessibility and release validation; all acceptance criteria met |

Every milestone includes relevant tests and documentation. Financial tests use small official fixtures, independently checked nonzero-shock expectations, explicit tolerances, and maturity/numeric boundary cases. Routine tests must not depend on live internet access.

The [CI workflow](.github/workflows/ci.yml) covers formatting, static analysis, unit tests, race checks, a build without CGO, and an offline CLI smoke check. It must run on GitHub after pushing; creating the workflow does not verify a remote run. Future releases must include checksums and be tested against the documented launch instructions.

## Documentation and contributions

- [V1 specification](docs/V1_SPEC.md): product scope, calculations, data semantics, and acceptance criteria.
- [M1 implementation](docs/M1_IMPLEMENTATION.md): available behavior, package boundaries, validation, and limitations.
- [Official datasource contract](docs/OFFICIAL_DATASOURCE.md): verified CKAN/CSV contract and parser limitations.
- [Next steps](docs/NEXT_STEPS.md): current checkpoint and ordered implementation checklist through V1.
- [Contributing](CONTRIBUTING.md): language policy, development workflow, checks, and review expectations.
- [Agent instructions](AGENTS.md): mandatory rules for automated contributors.

The specification is the source of truth when documents disagree. Changes outside the locked V1 scope require an explicit maintainer decision.

## License

Project code is covered by the [MIT license](LICENSE). Third-party assets and official data retain their own applicable licenses and attribution requirements; fixtures must document their provenance.
