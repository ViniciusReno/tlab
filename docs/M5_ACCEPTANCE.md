# M5 acceptance audit

Audit date: 2026-10-08. The specification's section 34 remains the source of truth.
This record separates local evidence from release gates. V1 is not yet released.

## Packaging

| Criterion | Evidence / status |
|---|---|
| Single release binary | Five archives built with `CGO_ENABLED=0`; extracted macOS arm64 and amd64 demos executed. |
| No Node/npm runtime | Embedded server-rendered templates and vanilla JavaScript; binary smoke checks use empty PATH. An existing empty, untracked `package-lock.json` is unrelated to the application and is not packaged. |
| No external database | Embedded pure-Go SQLite, versioned SQL migrations; no database service. |
| No Docker requirement | Extracted binaries launched directly. |
| Local/embedded assets | Smoke checks fetch embedded CSS and pages; source `assets.go`. |
| Loopback default | CLI binds `127.0.0.1:8080`; startup and smoke checks confirm the address. |
| Release/source demo commands start one process | CLI startup tests and extracted `demo` smoke; source command also checked locally. |
| No runtime/setup/network for release demo | macOS extracted binaries run with empty PATH, fresh temporary working/config directories, and external connections denied by `sandbox-exec`. No persistent database is created. Linux/Windows native execution remains a release gate. |

## Data

| Criterion | Evidence / status |
|---|---|
| Offline demo | Embedded-fixture tests and network-restricted extracted demo. |
| Temporary, isolated demo state | App/storage/CLI isolation tests, demo reset and portfolio tests; smoke finds no persistent `.db`. |
| Official dataset synchronization | CKAN/parser/download contracts and committed official fixtures; app/CLI sync tests. Previous milestone notes record live checks. No new live-source sync was run in this audit. |
| Idempotent sync | Storage/app sync tests cover duplicate imports. |
| Visible provenance | Market/history, Advanced, Learn, and portfolio source details; HTTP tests and Chrome review. |
| Official quote dates on current-looking values | Market/portfolio/freshness tests and browser review at 2026-09-04 fixture dates. |
| Explicit stale/older/matured states | App/web freshness and maturity tests; Chrome demo shows matured portfolio positions. |
| Valid data survives failed sync | Storage/app rollback and separate failure-status tests. |

## Instruments and analysis

| Criterion | Evidence / status |
|---|---|
| Prefixado market/history | Parser, app/HTTP history tests; Chrome follows history links for all three contexts. |
| IPCA+ market/history | M3 fixture/app/CLI/web tests; Chrome synced real-yield scenarios. |
| Selic market/history/base-PU MTM | Parser/app/portfolio tests and Chrome dated market/portfolio review; scenarios explicitly unsupported. |
| Unsupported instruments identified | Parser unsupported counts and scenario/domain tests; no zero-coupon approximation. |
| Validated Prefixado scenario | Official-methodology fixtures, independent bidirectional shocks, context and redemption validation tests. |
| Validated IPCA+ same-date scenario | Independent fixed-indexation expectations in pricing/app/CLI tests. |
| Independent nonzero-shock magnitudes | Committed fixed expectations and reference documentation for both instruments; no production-generated expected outputs. |
| Numeric/remaining-term errors | Pricing/calendar/app tests reject nonfinite inputs, nonpositive values, invalid term, and overflow. |
| Official PU Base portfolio MTM | Portfolio tests; Chrome three-position fixture total BRL 45,532.36. |
| Gross Prefixado hold/early-exit comparison | Portfolio/app/web tests, validation gate and visible gross/assumption wording. |
| No silent inflation forecast | IPCA fixed-indexation contract, CLI/API/HTML parity tests, Learn explanations. |
| No tax/net-return claim | Gross labels and explicit tax/fee limitations in result, portfolio, and Learn. |

## UX

| Criterion | Evidence / status |
|---|---|
| Simple mode default | Native Advanced disclosures start closed; browser review. |
| Advanced technical details | Source, dates, calendar/settlement/calculation/build versions, formula, JSON, and reproduction command. |
| Same simple/advanced result | One shared `app.Result`; app/CLI/HTTP parity tests. |
| CLI/browser reproducibility | Explicit source/date/basis/yield parity tests and generated commands. Source corrections/version limits are documented. |
| Official/derived/hypothetical distinctions | Quote-context labels, gross/scenario explanations, and provenance details. |
| Market/Playground/Portfolio/Learn work | HTTP journey tests, Chrome demo/persistent review, and extracted-page smoke. |
| English authored content | Reviewed changed UI/help/docs. V1 has no localization; approved M6 stays planned. |
| Official source labels/provenance preserved | Raw CSV fixtures unchanged; English explanations retain proper names. |

Additional section 32 checks: Chrome 154 on macOS, viewport widths 375 and 1280;
no page-level horizontal overflow in Market, History, Playground, Portfolio, and
Learn. Wide tables are keyboard-focusable scrolling regions. Skip-to-content works
via Tab/Enter; slider updates retain keyboard focus; shock/form results receive
focus; Advanced disclosure remains open across async updates. Chart values are
also present in a numeric table. Native disclosure controls and labeled inputs are
used, with explicit decimal-point and ISO-date entry conventions.

With browser JavaScript disabled, Learn and navigation remain usable, Playground
forms submit, and portfolio creation/editing/confirmed removal succeed. Missing
scenario contexts show an error and recover on corrected resubmission. A real
Chrome portfolio POST exposed `Origin: null` under `no-referrer`; `same-origin`
now permits legitimate forms while preventing external referrer disclosure.
Security tests retain rejection of hostile/opaque origins and invalid CSRF tokens.
Screen-reader behavior, other browser engines, and a formal accessibility
certification were not evaluated.

## Engineering and release gates

| Criterion | Evidence / status |
|---|---|
| Tests / vet pass | Exact current commands and outcomes are recorded below. |
| Source parser fixtures | Committed CKAN metadata, official CSV extracts, malformed/schema tests. |
| Official pricing fixtures | Versioned provenance and independent expected values in demo/pricing testdata. |
| Versioned migrations | Existing 001–003 unchanged; no new migration needed in M5. |
| No unnecessary V2 infrastructure | No dependency/version upgrade, frontend runtime, telemetry, cloud service, or localization introduced. |
| One repo/root Go module | Existing module and package boundaries preserved. Python scripts are optional maintainer packaging/check tools. |
| Automated checks/docs/release notes match | Existing Go CI plus five-target native archive/checksum/extracted-demo workflow; contributor and candidate-release documentation. Remote results remain pending. |

Release gates still open:

- Successful remote CI for the exact reviewed commit, including native Linux
  amd64/arm64 and Windows amd64 execution. macOS amd64 currently ran through
  Rosetta on arm64, not on native Intel hardware.
- Review and publish an appropriately versioned release after those checks. Local
  candidates are unsigned/unnotarized and unpublished.

## Validation commands

Final local results: formatting, vet, unit tests, race tests, five-target builds, all archive checksums, both extracted macOS smoke checks, documentation-link validation, and `git diff --check` passed.

Local scripts use Python's standard library only. They are not runtime dependencies.
Packaging refuses an existing output directory and preserves pinned Go dependencies.

```sh
go fmt ./...
go vet ./...
go test ./...
go test -race ./...
python3 scripts/package_release.py --version 0.1.0-m5-candidate --out /private/tmp/tlab-m5-final-packages
python3 scripts/smoke_release.py /path/to/extracted/tesouro-lab --version 0.1.0-m5-candidate
git diff --check
```

Both extracted macOS smoke commands additionally ran under this profile:
`(version 1)(allow default)(deny network-outbound)(allow network-outbound (remote ip "localhost:*"))`.
The smoke script alone redirects external HTTP(S) proxies and removes PATH; only
the additional macOS profile establishes OS-level outbound network isolation.
SHA-256 hashes of all five generated archives were verified after packaging.

The Chrome checks used a temporary profile and temporary automation scripts, with
an isolated persistent test database populated through the official committed
metadata/CSV fixtures. Quantities were hypothetical; no personal portfolio or live
sync was involved. Temporary browser helpers are not application dependencies.
