# Tesouro Lab

A local educational lab for Brazilian government bonds, built around official data, transparent calculations, and hypothetical scenarios.

**Status: M1 offline Prefixado demo implemented.** Run from source or build a local executable. No downloadable release has been published yet; the remaining V1 milestones are planned.

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

## Run from source

The browser and CLI use the same calculation. Reproduce a hypothetical scenario from the official historical example:

```sh
go run ./cmd/tesouro-lab analyze prefixado:2015-01-01 --source demo --basis purchase --date 2012-01-03 --yield 8.88 --amount 10000
```

This changes the example's annual yield from 10.88% to a hypothetical 8.88%, producing a scenario unit price of approximately BRL 774.99 from the official BRL 733.86 baseline. These are not current market values. See [fixture provenance and assumptions](data/demo/README.md).

Docker may become an optional convenience later. It is not the primary installation path or an M1 dependency.

## V1 scope

The table describes the target V1, not today's complete feature set. M1 implements the Prefixado purchase-context scenario demo only; IPCA+, Selic, live history, and portfolios are not yet available.

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
| `tesouro-lab` | Persistent mode is planned for M2; currently returns an explicit unavailable error |
| `tesouro-lab demo` | Start the isolated offline demo |
| `tesouro-lab sync` | Official synchronization is planned for M2; currently returns an explicit unavailable error |
| `tesouro-lab analyze <bond-id> --source demo --yield <percent>` | Run the same scenario calculation used by the browser |

`analyze` accepts `--source demo|synced`, `--basis purchase|mark_to_market|early_exit`, `--date YYYY-MM-DD`, and an optional `--amount` in BRL. For example, `--yield 12.00` means 12% per year. The source defaults to `synced`, which is unavailable in M1, so pass `--source demo`. The default context is purchase and the default date is the latest available quote for the bond. The resolved quote date is reported. The fixture has no base/sell quotes; those contexts return `missing_quote`. Analysis never triggers synchronization automatically.

Reproducing a result requires the same quote contents and calendar/calculation versions. Advanced results provide an equivalent CLI command with explicit inputs. Official source corrections can change historical results; see the [CLI contract](docs/V1_SPEC.md#24-cli).

## Official quote contexts

The source dataset distinguishes three unit prices (PU):

| Source field | Meaning | Settlement |
|---|---|---|
| `PU Compra Manha` | Purchase quote | Next financial-market business day, D+1 |
| `PU Venda Manha` | Early-redemption quote | Next financial-market business day, D+1 |
| `PU Base Manha` | Portfolio mark-to-market valuation | Quote date, D0 |

Portfolio valuation uses the official base unit price. Hypothetical early redemption uses the official sell unit price. Every current-looking value includes its official quote date; the import timestamp is not a market date. Portfolio and early-exit values are gross.

Source metadata, formulas, and limitations are documented in the [V1 specification](docs/V1_SPEC.md).

## Implementation milestones

| Milestone | Demonstrable outcome |
|---|---|
| M1 — Implemented Prefixado demo | One command starts an offline demo; a user can change a yield and reproduce the result through the CLI |
| M2 — Official data and dates | Idempotent synchronization, history, explicit quote contexts, and accurate freshness states |
| M3 — IPCA+ and Selic | Validated real-yield scenarios for IPCA+; explicitly limited Selic display/valuation |
| M4 — Local portfolio | Quantity-only positions, optional acquisition cost, and gross Prefixado hold/early-exit comparison |
| M5 — V1 completion | Market, Playground, Portfolio, and Learn screens; accessibility and release validation; all acceptance criteria met |

Every milestone includes relevant tests and documentation. Financial tests use small official fixtures, independently checked nonzero-shock expectations, explicit tolerances, and maturity/numeric boundary cases. Routine tests must not depend on live internet access.

The [CI workflow](.github/workflows/ci.yml) covers formatting, static analysis, unit tests, race checks, a build without CGO, and an offline CLI smoke check. It must run on GitHub after pushing; creating the workflow does not verify a remote run. Future releases must include checksums and be tested against the documented launch instructions.

## Documentation and contributions

- [V1 specification](docs/V1_SPEC.md): product scope, calculations, data semantics, and acceptance criteria.
- [M1 implementation](docs/M1_IMPLEMENTATION.md): available behavior, package boundaries, validation, and limitations.
- [Next steps](docs/NEXT_STEPS.md): current checkpoint and ordered implementation checklist through V1.
- [Contributing](CONTRIBUTING.md): language policy, development workflow, checks, and review expectations.
- [Agent instructions](AGENTS.md): mandatory rules for automated contributors.

The specification is the source of truth when documents disagree. Changes outside the locked V1 scope require an explicit maintainer decision.

## License

Project code is covered by the [MIT license](LICENSE). Third-party assets and official data retain their own applicable licenses and attribution requirements; fixtures must document their provenance.
