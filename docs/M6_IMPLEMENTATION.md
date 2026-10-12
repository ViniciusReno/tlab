# M6 — browser language and usability

Implemented locally on 2026-10-09, after publication of V1 `v0.1.0`.
This is the maintainer-approved post-V1 language/usability scope recorded in
`NEXT_STEPS.md`. The published V1 archives remain English-only.

## Language and formatting contract

English is the default. Every primary page offers keyboard-accessible English and
Português links. An explicit `lang=en` or `lang=pt-BR` query selects the language;
a local, HttpOnly, SameSite=Lax cookie remembers it for up to one year. This
preference does not require an account, database migration, JavaScript, or network
access. Invalid/repeated language selections fail explicitly. No automatic
browser-language detection is introduced.

The selected language changes browser copy, accessible labels, unavailable
states, and application validation messages. Both languages retain explicit BRL,
decimal points, English grouping in displayed money, and ISO `YYYY-MM-DD` dates.
Inputs have no grouping separators: `10000.50` is an amount and `8.88` is a yield
percentage. This deliberate shared format keeps form entries, URLs, and CLI
reproduction commands directly comparable. Translated hints explain it; comma
decimals and day/month/year inputs are not accepted.

Official names, raw source records, formulas, technical codes, CLI/API fields,
calendar/settlement/build versions, and reproduction commands retain their
canonical representations. The API remains independent of browser preference.
Private notes are displayed as entered and escaped, never translated.

## Implementation and task review

One set of embedded templates and a Go translation catalog serve both languages.
Templates are parsed at startup for each language with immutable translation
functions. Both handlers use the same application service and CSRF token; a
language switch does not invalidate an open portfolio form. Locale selection is
removed before financial query parsing and never changes the source or database.
Existing request limits, security headers, and origin/token validation remain.

The review used concrete beginner tasks:

| Task | Implemented behavior and evidence |
|---|---|
| Start an offline example and understand its date | Demo notices, maturity limits, assumptions, and Learn explanations are available in both languages. |
| Change language while exploring a scenario | Links preserve the query, including the most recently calculated asynchronous shock; keyboard switching keeps the same financial result. |
| Enter an amount or yield | A visible, associated Playground hint explains canonical number/date conventions before entry; portfolio entry retains its translated instructions. |
| Inspect an official quote and open its history | Portuguese Market/History keep official dates, PU distinctions, stale/matured states, provenance, and scenario links. |
| Recover from an unavailable quote context | Translated errors retain selected inputs; correcting the context completes the same scenario. |
| Create/edit/remove a quantity-only position | Labeled forms and explicit removal confirmation work without JavaScript; missing acquisition cost remains unavailable. |
| Understand technical limits | Gross values, fixed IPCA+ indexation, Selic restrictions, quote contexts, and historical-calendar validation remain visible in both languages. |

The language picker identifies the selected language through text emphasis and
`aria-current`. HTML `lang` and response `Content-Language` match the selection.
Shared structure, skip links, disclosures, numeric chart tables, and focus behavior
are retained. The asynchronous scenario script updates language links alongside
the displayed result and URL.

## Validation

Executed locally:

```sh
go fmt ./...
go vet ./...
go test ./internal/web
go test ./...
go test -race ./...
CGO_ENABLED=0 go build -trimpath -o /private/tmp/tlab-m6-final ./cmd/tesouro-lab
git diff --check
```

One initial full-test attempt was blocked by the sandbox when the CLI test bound
a temporary loopback port. The full suite passed when rerun with the required
permission. Translation tests check every literal template lookup and domain
error, exact scenario monetary/CLI/JSON parity, language preference/link behavior,
duplicate financial inputs, escaping, shared CSRF tokens, and empty persistent
source isolation. Committed official records also exercise Portuguese Market and
History. Financial-domain tests remain unchanged.

Headless Chrome on macOS checked both languages at 375px and 1280px with no
page-level overflow in Playground, Portfolio, Learn, Market, and History. The
checks used private in-memory demo state and an isolated persistent database
populated from the committed official CSV fixture; no personal data or live sync
was used. Keyboard skip/language links, asynchronous shocks, unchanged-result
language switches, unavailable-context recovery, and no-JavaScript scenario and
portfolio creation/editing/confirmed removal passed. Populated persistent checks
confirmed official-date/stale/Selic states, history-to-scenario navigation, and
quantity-only valuation. Test timestamps and quantities were explicit test inputs.

Temporary Chrome/Node/Python helpers were used only for maintainer verification;
they are not committed or application runtime dependencies. Local tests used
ports 18082–18084 to preserve the existing server on 8080.

## Limitations and release status

M6 has no new dependency, migration, financial formula, inflation assumption,
instrument support, or runtime requirement. Locale-specific numeric/date input
formats are intentionally not introduced. CLI/API remain English/canonical.
Screen-reader behavior, other browser engines, and formal accessibility
certification were not evaluated. The original 2026-10-09 review did not perform a live-source sync; the closure check below did.

Stable release validation completed on 2026-10-12 under explicit maintainer authorization. [PR #1](https://github.com/ViniciusReno/tlab/pull/1) merged M6 and the release smoke checks into `main` at `5752a97`. The [Go checks](https://github.com/ViniciusReno/tlab/actions/runs/38308467948) and [five-target native run stamped 1.0.0](https://github.com/ViniciusReno/tlab/actions/runs/38308470160) passed. Each native runner built, hashed, extracted, and executed the binary with empty PATH, including English/pt-BR preference and canonical API checks.

Locally, all five 1.0.0 archives cross-built without CGO and their checksums passed. Both extracted macOS demos passed under OS-level external-network denial at port 18085; amd64 used Rosetta. The source `go run ./cmd/tesouro-lab demo --port 18086` also started offline and served Portuguese. One initial repeated smoke was blocked by Unix TIME_WAIT after shutdown; the committed probe fix permits repeated checks while rejecting an active listener. That rejection was tested explicitly. No running user server was stopped.

A live official sync on 2026-10-12 used only an isolated temporary database: 177,260 source rows read, 68,577 supported rows stored, maximum official quote date 2026-10-09. Unsupported types were counted and excluded. No personal portfolio was accessed.

The implementation and release validation are complete. See [project closure](PROJECT_CLOSURE.md) for publication and repository status.
