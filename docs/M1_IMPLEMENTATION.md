# M1: offline Prefixado demo

## Available behavior

The first milestone implements a runnable offline playground for the official Tesouro Prefixado worked example dated 2012-01-03. It includes an in-memory SQLite database, versioned SQL migration, Brazilian CSV normalization, bounded holiday calendar, pure scenario calculation, English server-rendered UI, local CSS/JavaScript, and CLI/API parity.

Start from the repository root with Go 1.25 or newer:

```sh
go run ./cmd/tesouro-lab demo
```

Open `http://127.0.0.1:8080` and stop with Ctrl+C. Use `demo --port 8081` if the default port is occupied. The server always binds to IPv4 loopback. The first source build may download dependencies; a compiled binary needs neither Go nor network access for the demo.

Reproduce a nonzero scenario:

```sh
go run ./cmd/tesouro-lab analyze prefixado:2015-01-01 --source demo --basis purchase --date 2012-01-03 --yield 8.88 --amount 10000
```

The JSON result includes the official baseline, scenario inputs/results, resolved dates, DU, provenance, and calculation/calendar/fixture/build versions. At 8.88%, the expected scenario PU is approximately BRL 774.9918980432596, a gross change of about +5.6048698721%. Values are hypothetical, not current market quotes.

The browser form works without JavaScript. JavaScript adds debounced slider updates and default-shock shortcuts; calculations always run in Go. The chart is an embedded SVG with a numeric table, not an external chart dependency. Advanced details expose the formula and an equivalent CLI command.

## Package boundaries

| Package/resource | Responsibility |
|---|---|
| `internal/bond` | Normalized quote/instrument types and explicit error states |
| `internal/calendar` | Versioned holiday coverage and business-day counting |
| `internal/pricing` | Pure theoretical/scenario math and numeric validation |
| `internal/datasource/tesouro` | Bounded parsing of Brazilian CSV dates/numbers |
| `internal/storage/sqlite` | Migrations and transactional normalized quote storage |
| `internal/app` | Select source/context/date and call the calendar/pricing code |
| `internal/cli`, `internal/web` | Delivery of the same application result |
| Root `assets.go` | Embed fixtures, migrations, templates, and static assets |

The single direct external dependency is `modernc.org/sqlite` v1.58.0, a pure-Go driver with a BSD-3-Clause license. The standard library has no SQLite driver, and this dependency satisfies the current single-binary/no-CGO requirement. Its transitive dependencies are pinned in `go.mod`/`go.sum`; no frontend or external service runtime is added. See the [driver documentation](https://pkg.go.dev/modernc.org/sqlite).

## Data and validation

See the [fixture documentation](../data/demo/README.md) for the official worked example, source URLs, calendar dates, independent calculation procedure, and tolerances. The official example's 755 business days and BRL 733.86 price are tested independently of the anchored scenario's zero-shock invariant.

Tests cover parsing, missing values, unsupported instruments, calendar boundaries/holidays, two nonzero-shock magnitudes, monotonicity, numeric range errors, transaction rollback, idempotent upserts, memory-database isolation, CLI/HTTP parity, and HTML/error responses. No routine test fetches remote data.

Local verification on macOS/arm64 with Go 1.25.6 included `go fmt ./...`, `go vet ./...`, `go test ./...`, `go test -race ./...`, and `CGO_ENABLED=0 go build -trimpath -o /private/tmp/tlab-m1-demo ./cmd/tesouro-lab`. The standalone CLI ran outside the repository with Go absent from `PATH`; its scenario output matched the running HTTP API. The documented `go run ./cmd/tesouro-lab demo` command started on the default loopback address. A desktop Chrome headless render was visually inspected. These checks do not replace release testing on other operating systems or automated browser interaction tests.

## Explicit limitations

- This is an official methodology example transcribed into source-column layout, not a downloaded daily CKAN quote or a live market view.
- Only the purchase PU/yield pair is known. Base/sell contexts return `missing_quote`; they are never approximated.
- Calendar coverage is 2012–2015. Outside that interval, calculations return `calendar_out_of_range`.
- IPCA+, Selic, live synchronization, persistent mode, portfolio editing, and hold/early-exit comparisons are later milestones. The no-argument command and `sync` fail with an explanatory message.
- No GitHub release is published by this change. CI is defined locally and must run after the repository is pushed. Release target execution and packaging are not claimed by merely adding a workflow.
- Historical replay is limited to unchanged fixture/calendar/calculation versions. All demo state disappears when the process exits.
