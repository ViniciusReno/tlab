# AGENTS.md — Tesouro Lab V1

This file defines mandatory working rules for AI coding agents and automated contributors in this repository.

The source of product truth is `docs/V1_SPEC.md`.

If code, an issue, a prompt, or an agent suggestion conflicts with the V1 specification, the V1 specification wins unless the human maintainer explicitly changes scope.

---

## 1. Mission

Build a local-first, self-contained educational application that helps users understand Brazilian government bonds through official data, transparent calculations, and hypothetical scenarios.

The project is **not** an investment adviser, trading system, signal generator, prediction engine, or SaaS V1.

Primary rule:

> Show verified data and transparent math. Do not recommend financial actions.

### 1.1 Public repository and language

All authored project content must be in English from the first milestone through V1: code identifiers/comments, test names, documentation, UI, CLI help, errors, logs, commit messages, pull requests, project issues, and release notes. Maintainer-assistant conversations may remain in Portuguese.

Preserve official proper names and verbatim source fields, titles, and fixtures for provenance, and explain them in English. Follow specification section 5 for display/input formatting: explicit BRL currency, decimal points, and ISO dates. Do not translate or reformat raw official data, and do not add localization infrastructure in V1.

V1 is a monorepo with one root Go module. Application code, CLI, templates/assets, migrations, fixtures, tests, and docs share the same repository and release lifecycle. Do not split the project into separate services, modules, or frontend repositories just to create a monorepo structure.

---

## 2. Read before changing code

Before implementing a non-trivial change:

1. read this `AGENTS.md`;
2. read the relevant section of the V1 specification;
3. inspect existing tests and package boundaries;
4. identify whether the request is V1 scope;
5. implement the smallest coherent change;
6. run relevant tests;
7. report limitations instead of hiding them.

Do not start by adding dependencies, infrastructure, abstractions, or new packages.

### 2.1 Implementation sequence

Unless the maintainer explicitly requests otherwise, implement V1 in this order:

1. **M1 Demo Prefixado** — single binary, offline demo, one official Prefixado fixture, anchored scenario, simple/advanced result, CLI parity;
2. **M2 Official sync** — CKAN import, quote dates/freshness, history, and correct Buy/Sell/Base PU contexts;
3. **M3 IPCA+ and Selic** — IPCA+ same-date real-yield scenario; Selic market/history/current MTM only;
4. **M4 Portfolio** — quantity-only positions, optional cost basis, PU Base MTM, Prefixado hold-vs-early-exit;
5. **M5 V1 closure** — remaining screens, accessibility, release targets, acceptance checklist.

Do not implement later-milestone infrastructure preemptively. A milestone may reuse clean domain code from an earlier milestone, but it must not pull V2 scope forward.

---

## 3. Hard V1 scope

### Supported analytical instruments

Full analytical support:

- Tesouro Prefixado / LTN without coupon;
- Tesouro IPCA+ / NTN-B Principal without coupon.

Display/history/current-MTM support:

- Tesouro Selic / LFT, using official `PU Base` for portfolio mark-to-market only.

Do not run Tesouro Selic yield-shock, duration/convexity, or hold-vs-early-exit calculations in V1.

### Explicitly unsupported in V1

Do not implement unless the maintainer explicitly changes scope:

- Prefixado com Juros Semestrais;
- IPCA+ com Juros Semestrais;
- RendA+;
- Educa+;
- CDB/LCI/LCA;
- debentures;
- funds;
- equities;
- ETFs;
- crypto;
- FX;
- derivatives.

Do not approximate unsupported instruments using the supported formulas.

Correct behavior is to return/display `unsupported`.

---

## 4. Forbidden V1 features

Do not add:

- AI/LLM features;
- predictions;
- recommendation engines;
- opportunity scores;
- buy/sell signals;
- brokerage APIs;
- trading automation;
- user accounts;
- authentication;
- cloud database;
- external portfolio sync;
- telemetry;
- analytics trackers;
- ads;
- push/e-mail alerts;
- multi-user features;
- React/Vue/Svelte/Vite;
- Node/npm runtime;
- PostgreSQL;
- Redis;
- Kafka/queues;
- Kubernetes;
- microservices;
- mandatory Docker;
- a generic “future SaaS” abstraction layer.

Do not create V2 infrastructure “for later”.

---

## 5. Architectural constraints

V1 must remain easy to download and run.

The primary trial command is `./tesouro-lab demo` after extracting a release (`.\tesouro-lab.exe demo` in Windows PowerShell). The source equivalent is `go run ./cmd/tesouro-lab demo` from the repository root with Go installed. One process initializes the demo database, loads fixtures, serves assets, and prints the local URL and shutdown instructions. No manual database setup, sync, configuration file, or extra runtime is required for the release demo. Keep Docker optional and outside the primary quick start; do not add container tooling during M1.

### Required direction

- Go application;
- single release binary;
- `net/http`;
- `html/template`;
- SQLite;
- pure-Go SQLite driver;
- server-rendered HTML;
- minimal vanilla JavaScript;
- local embedded static assets;
- explicit SQL migrations;
- loopback-only bind by default.

### Default network address

```text
127.0.0.1:8080
```

Never change the default to `0.0.0.0` without explicit maintainer instruction.

### SQLite

Prefer the project-selected pure-Go SQLite driver. Do not switch to a CGO-dependent driver without a demonstrated requirement and maintainer approval.

### Frontend dependencies

Do not add a frontend build toolchain.

If a small JS/chart library is required:

- prefer one existing approved library;
- vendor/ship the production asset locally;
- preserve license/attribution;
- do not require CDN access at runtime.

---

## 6. Dependency policy

Prefer the Go standard library.

A new dependency is acceptable only when it materially reduces correctness or portability risk.

Before adding one, verify:

1. the standard library is insufficient;
2. the dependency solves a current V1 need;
3. it does not add a required external runtime/service;
4. it is maintained and appropriately licensed;
5. it does not force architecture outside this spec.

Do not add frameworks for convenience alone.

Do not upgrade existing major dependencies or the Go version as a side effect of unrelated work.

---

## 7. Domain/package boundaries

Financial logic must be deterministic and isolated.

Packages containing pricing/scenario math must not depend on:

- HTTP handlers;
- HTML templates;
- SQLite;
- CLI argument parsing;
- remote APIs.

Datasource code may depend on external source formats but must normalize them before returning domain objects.

Storage code must not contain financial formulas.

HTTP handlers must not duplicate financial formulas.

CLI commands must call the same domain services used by the UI.

Simple and advanced UI modes must consume the same calculation result.

---

## 8. Financial correctness rules

### 8.1 Never invent market data

Current/historical market values must come from the configured official datasource or clearly labeled demo fixtures.

Never invent:

- current yield;
- current PU;
- current Selic;
- current IPCA;
- maturity;
- historical quote.

When unavailable, return an explicit unavailable state.

### 8.2 Official quote contexts are distinct

Never treat the three official PUs as interchangeable:

```text
PU Compra = purchase quote, D+1
PU Venda  = early-redemption quote, D+1
PU Base   = mark-to-market valuation, D0
```

For an existing non-matured portfolio position, if official `PU Base` is available:

```text
current gross MTM = quantity × official PU Base
```

Use `PU Venda` only for a hypothetical early-exit/redemption calculation.
Use `PU Compra` only for a hypothetical purchase context.

Do not replace an official PU with a theoretical repricing unless the feature explicitly asks for a scenario calculation.

### 8.3 Scenario is not forecast

Every yield shock is hypothetical.

Preferred language:

```text
If the yield were X, holding the other scenario conditions constant...
```

Never write:

```text
When the yield falls to X...
```

unless describing a known historical event.

### 8.4 No silent inflation assumption

If an IPCA+ result requires future inflation:

- request/use an explicit user assumption; or
- do not calculate the result.

Never hard-code an assumed future IPCA as if it were fact.

### 8.5 Gross vs net

V1 has no complete tax/cost engine.

Values calculated by portfolio/exit analysis are gross unless explicitly proven otherwise.

Always label them as gross where a user could confuse them with cash received.

### 8.6 Float behavior

Do not test financial `float64` values using exact equality.

Use documented tolerances.

Do not round intermediate values merely for display convenience.

Apply official truncation/rounding only where required by the documented methodology.

### 8.7 Quote date is mandatory context

Never present `ImportedAt` as a market date.

Every current-looking market/portfolio value must carry the official `quote_date` (`Data Base`).

Do not write `today` unless the official quote date equals the user's current local date, and still show the explicit date.

If a bond quote is older than the dataset's latest quote date, show that state. Never forward-fill a missing quote.

A matured bond must be shown as matured, not valued using its last pre-maturity quote.

### 8.8 Scenario basis contract

Supported zero-coupon scenarios are anchored to the official PU/yield pair for the context:

```text
purchase:       BuyPU  + BuyYield  + D+1
mark_to_market: BasePU + SellYield + D0
early_exit:     SellPU + SellYield + D+1
```

Use:

```text
PU_scenario = PU_base × ((1 + base_yield) / (1 + scenario_yield))^(DU/252)
```

`DU` is counted from settlement inclusive to maturity exclusive.

A zero shock must reproduce the official base PU within `1e-9` before presentation.

For standalone Prefixado validation against official PU, truncate the theoretical PU to two decimals and require absolute difference `<= BRL 0.01`.

Do not loosen tolerances to make a fixture pass.

Apply `docs/V1_SPEC.md` sections 13 and 15: reject non-finite inputs and nonpositive PUs/quantities/amounts; require settlement before maturity and verified `DU > 0` for scenarios. Return `no_remaining_term` or `calculation_out_of_range` as specified instead of producing invalid numeric results. Historical scenarios use their selected quote/settlement dates even if the instrument has matured today.

### 8.9 Portfolio input completeness

Minimum position input is title + positive quantity.

If acquisition cost is unknown, current official MTM may still be shown, but acquisition-dependent results must be unavailable rather than inferred.

Never infer the user's actual purchase cost from a historical daily market quote.

### 8.10 Hold vs early exit scope

V1 requires gross hold-vs-early-exit comparison for Tesouro Prefixado only.

It is unsupported for Tesouro IPCA+ and Tesouro Selic in V1.

The comparison requires positive finite amounts and `DU > 0`; never annualize over a zero or negative remaining term.

---

## 9. Source-of-truth policy

For financial behavior, use this hierarchy:

1. Tesouro Nacional/Tesouro Direto official data;
2. Tesouro Direto official methodology and dataset metadata;
3. ANBIMA official financial-market holiday calendar for business-day fixtures;
4. Banco Central official data for allowed macro context;
5. repository fixtures derived from official sources;
6. internal derived calculations.

Avoid using blogs, YouTube, news articles, or third-party finance sites as normative implementation sources when an official source exists.

Third-party sources may be used only to investigate an issue, never as silent authority for a financial formula.

---

## 10. Datasource rules

### Tesouro Transparente

The official CKAN dataset is the primary source for daily prices/rates.

Datasource implementation must:

- isolate source field names;
- parse Brazilian dates explicitly;
- parse comma-decimal values explicitly;
- handle semicolon CSV where applicable;
- use timeouts;
- limit download size;
- validate required columns;
- reject malformed required values;
- preserve provenance;
- upsert idempotently.

Do not depend on the machine locale.

### Sync failure

A failed sync must not delete or corrupt previous valid data.

Use a transaction for the normalized write phase.

Record sync status/error separately.

---

## 11. Database rules

Use versioned migrations.

Do not mutate an already released migration. Add a new migration.

Use natural uniqueness for quotes:

```text
bond_id + quote_date
```

Do not use SQLite-specific behavior inside financial/domain packages.

Do not add an ORM unless explicitly approved. Prefer clear SQL and `database/sql`.

---

## 12. UI rules

### Simple mode is default

Assume the user is curious but not financially trained.

Every primary screen should answer:

1. what am I seeing?
2. why does it matter?
3. what happens if the relevant variable changes?

### Advanced mode

Advanced mode exposes technical details without changing the result.

Do not maintain parallel simplified math.

### Explain technical terms in context

Preferred:

```text
How sensitive is this bond to yield changes?
High sensitivity

Technical detail: modified duration = 4.38
```

Avoid presenting unexplained finance acronyms as the default UI.

### No color-only meaning

Charts/statuses must include text/numbers. Do not communicate positive/negative results by color alone.

---

## 13. Recommendation-language policy

Do not produce UI copy, logs, docs, examples, or API fields that imply investment advice.

These restrictions concern advice, imperatives, and rankings. Descriptive terms such as `purchase quote`, `sell unit price`, and technical identifiers such as `BuyPU` or `SellYield` are allowed under specification section 33. Do not enforce a blanket word blacklist.

Avoid:

```text
buy
sell
best
worst
opportunity
recommended
should buy
should sell
entry
exit signal
target yield
strong buy
worth buying
worth selling
```

Use:

```text
current quote
historical quote
scenario
hypothesis
sensitivity
gross market value
mathematical comparison
simulated yield
simulated price
```

A button may say `Simulate early redemption` if it calculates a hypothetical early-exit value. It must not say `Sell` because the app does not execute or recommend a transaction.

---

## 14. Error policy

Unknown is not zero.

Unsupported is not approximate.

Missing is not inferred unless the inference is mathematically defined, documented, and visible as derived data.

Prefer explicit typed errors/states for:

- unsupported instrument;
- missing quote;
- invalid scenario;
- source unavailable;
- source schema changed;
- insufficient inputs;
- calculation not validated.

User-facing errors should explain the state without stack traces.

---

## 15. Testing requirements

A financial calculation is not done until it has tests.

At minimum, changes to financial/data code require relevant tests for:

- parser behavior;
- normalization;
- business-day logic;
- pricing;
- scenario math;
- storage idempotency;
- edge cases.

### Required invariants for supported zero-coupon scenario pricing

These invariants apply to valid finite inputs with a positive remaining business-day count.

```text
if yield decreases, price increases
if yield increases, price decreases
if yield is unchanged, anchored scenario PU equals official base PU within 1e-9
```

Business-day count must use the repository's versioned ANBIMA-based financial-market holiday fixture. Dates outside verified calendar coverage return `calendar_out_of_range`; do not fall back to weekdays only.

### Official-data fixtures

Prefer small, committed fixtures based on official historical records.

Record fixture provenance in the fixture or adjacent documentation.

For each analytical instrument, test nonzero shocks in both directions against independent fixed expected PUs and variations, with independently verified DU and documented tolerances. Follow specification section 28.3; production pricing/calendar functions must not generate their own expected outputs. Include maturity and numeric-range boundaries.

Do not make normal unit tests depend on live internet access.

Network integration tests, if added, must be explicitly separated from deterministic unit tests.

---

## 16. Commands to run before considering work complete

Run the relevant subset during development and, before finalizing a normal code change, run:

```bash
go fmt ./...
go vet ./...
go test ./...
```

When supported and relevant:

```bash
go test -race ./...
```

Do not claim tests passed if they were not run.

If a test cannot run in the current environment, state that explicitly.

---

## 17. Change discipline

Keep changes small and focused.

Do not combine:

- feature work;
- broad refactor;
- dependency upgrades;
- formatting entire unrelated directories;
- architecture redesign;

in the same change unless required.

Do not rename packages/files casually if it makes review harder.

Do not “clean up” unrelated code while implementing a focused task.

---

## 18. No premature abstraction

Do not create interfaces merely because there is one implementation.

Create an interface when at least one of these is true:

- there are multiple real implementations;
- testing requires a boundary around external I/O;
- the domain boundary is materially clearer because of it.

Do not create repositories/services/managers/providers/factories in layers without a concrete reason.

This project should remain understandable to a developer reading it on GitHub.

---

## 19. No premature optimization

Do not add caches, worker pools, goroutine pipelines, background schedulers, or batching frameworks without measurements showing a problem.

Correctness and clarity outrank theoretical throughput.

The dataset and local workload do not justify distributed-system patterns.

---

## 20. Security/privacy rules

Even as a local tool:

- bind to loopback;
- set HTTP timeouts;
- escape rendered data;
- validate imported files;
- cap request/file sizes;
- avoid arbitrary filesystem access from HTTP inputs;
- avoid arbitrary remote URL fetches;
- do not log private portfolio notes unnecessarily;
- do not add telemetry.

No secrets should be required for normal V1 operation.

---

## 21. Documentation rules

When adding a financial concept:

1. document the plain-language meaning;
2. document the technical term;
3. document the formula/assumption if calculated;
4. identify the source;
5. state limitations;
6. add tests.

Do not write marketing claims.

README examples must use scenario language and avoid investment advice.

Keep README quick starts accurate and mark planned commands/features until implemented. Maintain `CONTRIBUTING.md` with prerequisites, checks, and contribution expectations. Add documentation and tests alongside each milestone; M5 is final verification, not the first documentation/testing pass. Follow specification section 29 for automated checks and release validation once code exists.

---

## 22. Demo-data rules

Demo mode must remain usable offline.

Demo fixtures must:

- be small;
- be deterministic;
- have provenance;
- be clearly labeled in the UI;
- cover both supported analytical bond types;
- not masquerade as current live data.

Do not silently fall back from failed live sync to demo data while labeling it as current.

Follow specification section 7: demo uses a private in-memory SQLite database per process, with temporary edits and no access to the user's persistent database. CLI demo analysis uses the same embedded fixture version. Demo requests must not sync official data or switch to persistent storage.

---

## 23. CLI/UI parity

When a calculation exists in both CLI and browser UI, both must use the same domain function/service.

Never reimplement pricing inside a handler or CLI command.

A reproducible scenario should produce the same numeric result across delivery mechanisms, subject only to presentation formatting.

Follow specification sections 24 and 25.3 for source, basis, exact quote date, units, defaults, and generated reproduction commands. CLI/URL yields are percentages (`12.00` means `0.12` in the domain). Report resolved inputs and calendar/build versions; never silently substitute a missing date, context, or source. Reproducibility across source corrections or changed versions is limited as documented; do not build an archival subsystem.

---

## 24. Current-value labeling

Distinguish these concepts in names and UI:

```text
Official buy quote
Official sell quote
Official base PU
Derived theoretical PU
Hypothetical scenario PU
Portfolio gross current value
```

Do not call all of them `price` without context.

Prefer explicit domain field names.

---

## 25. Business-day/calendar changes

Calendar logic is financial logic.

Any change to holidays/business-day calculation requires:

- tests;
- explanation of the calendar rule/source;
- revalidation against at least one official pricing fixture affected by the change.

Do not insert one-off date exceptions inside pricing formulas.

---

## 26. Handling ambiguous requirements

If a request can be implemented inside V1 without changing product meaning, choose the smallest compatible interpretation.

If a request clearly crosses V1 boundaries, do not silently implement it.

For an AI agent response, state concisely:

```text
This is outside the locked V1 scope because <reason>.
```

Then, if useful, describe where it would belong in V2 without creating the V2 implementation.

Human maintainer instruction can explicitly override scope.

---

## 27. Definition of done for a feature

A V1 feature is complete only when applicable items are satisfied:

- implementation respects package boundaries;
- official/derived/scenario provenance is preserved;
- simple-mode copy is understandable;
- advanced-mode inputs/results are inspectable;
- unsupported cases fail explicitly;
- tests exist and pass;
- no new runtime installation requirement was introduced;
- no recommendation language was added;
- docs are updated when behavior or assumptions changed.

---

## 28. Agent final-response format

After making a code change, report concisely:

1. **Changed** — what was implemented;
2. **Why** — relevant V1 requirement;
3. **Validation** — exact tests/checks run;
4. **Limitations** — anything still approximate, unsupported, or not verified.

Do not claim future work has been completed.

Do not hide failing tests.

Do not propose multiple unrelated follow-up features unless asked.

---

## 29. Core anti-drift checklist

Before adding anything, ask:

```text
Does this help explain Tesouro Direto V1?
Does it use official data or transparent assumptions?
Does it preserve local-first/single-binary operation?
Does it avoid recommendations/predictions?
Does it avoid adding infrastructure the product does not need?
Can a beginner understand the simple view?
Can an advanced user inspect the math?
Can the result be tested deterministically?
```

If multiple answers are `no`, stop and re-check the V1 specification.

---

## 30. Final principle

When choosing between:

```text
more features
```

and:

```text
fewer features with correct data, transparent math, strong tests, and clear explanations
```

choose the second option for V1.
