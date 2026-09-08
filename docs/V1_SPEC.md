# Tesouro Lab — Technical Specification V1

**Status:** scope locked for V1  
**Specification version:** 1.3  
**Working name:** Tesouro Lab  
**Audience:** contributors, reviewers, AI coding agents, and technical readers  
**Primary language:** English throughout the product and public repository  
**Distribution:** open-source, local-first, single-binary application

---

## 1. Product definition

Tesouro Lab V1 is a local, self-contained educational laboratory for Brazilian government bonds available through Tesouro Direto.

The application exists to help a user understand:

- what a government bond quote means;
- why bond prices change before maturity;
- the relationship between yield and price;
- how mark-to-market affects an existing position;
- how a hypothetical change in yield would affect a bond price;
- the difference between nominal and inflation-linked bonds;
- the difference between current market value and holding a bond to maturity.

The product **does not recommend investments** and **does not predict market direction**.

Core product rule:

> Do not tell the user what to think or what to buy. Show enough verified data and transparent math for the user to understand what is happening.

---

## 2. V1 goals

V1 must satisfy four goals simultaneously.

### 2.1 Educational

A beginner must be able to open the application and understand the main concepts without knowing terms such as PU, yield, duration, basis points, or mark-to-market in advance.

### 2.2 Technically inspectable

A developer or advanced user must be able to inspect the same calculation in technical detail, including inputs, formulas, assumptions, source data, and intermediate values.

### 2.3 Local-first and self-contained

A non-developer should be able to download a release binary and run the project without installing a database, Node.js, Docker, Python, Java, or any external runtime.

Target user flow:

```text
download and extract the release for your OS/architecture
      ↓
run tesouro-lab demo
      ↓
open http://127.0.0.1:8080
      ↓
explore the embedded demo without network access
```

After download/extraction, the primary trial path is one launch command: `./tesouro-lab demo` on Linux/macOS or `.\tesouro-lab.exe demo` in Windows PowerShell. The application initializes its demo database, loads fixtures, serves embedded assets, and prints `http://127.0.0.1:8080` and instructions to stop with Ctrl+C. Opening that URL is the only browser step; no separate database setup, migrations command, configuration file, account, or synchronization is required.

From a cloned repository with the supported Go toolchain installed, the equivalent development command is `go run ./cmd/tesouro-lab demo`. The first source build may require internet access to download Go dependencies; the packaged demo itself must work offline. These commands become runnable at M1, and documentation must clearly label them as planned until then.

The no-argument command continues to use persistent local data under section 24. Trying the demo never implicitly switches to synced mode. Docker is an optional future convenience, not a prerequisite or the primary quick start; V1 does not introduce a container stack merely to launch one process.

### 2.4 Good public GitHub case

The repository should demonstrate practical backend engineering rather than artificial architectural complexity:

- data ingestion;
- data normalization;
- financial-domain modeling;
- deterministic calculations;
- time-series storage;
- local persistence;
- HTTP/UI delivery;
- CLI support;
- testing against official data;
- explicit product constraints.

---

## 3. Non-goals

The following items are **out of scope for V1**.

- investment recommendations;
- buy/sell signals;
- opportunity scores;
- predictions of Selic, inflation, or bond yields;
- AI/LLM integration;
- brokerage integration;
- automated trading;
- portfolio synchronization with financial institutions;
- user accounts;
- authentication;
- cloud persistence;
- multi-user operation;
- mobile application;
- push notifications;
- e-mail alerts;
- paid features;
- CDB, LCI, LCA, debentures, funds, equities, ETFs, crypto, FX, or derivatives;
- portfolio optimization;
- full tax engine;
- accounting engine;
- real-time/intraday trading terminal;
- ANBIMA ETTJ/curve-relative opportunity ranking;
- ML models;
- generic investment-adviser language.

If a feature requires one of the items above, it is not a V1 feature.

---

## 4. Supported instruments

V1 intentionally limits analytical scope.

| Instrument | Market display | History | Portfolio current MTM | Yield-shock playground | Technical calculation view |
|---|---:|---:|---:|---:|---:|
| Tesouro Prefixado (LTN, no coupon) | Yes | Yes | Yes | Yes | Yes |
| Tesouro IPCA+ (NTN-B Principal, no coupon) | Yes | Yes | Yes | Yes | Yes |
| Tesouro Selic (LFT) | Yes | Yes | Yes, official PU Base only | No | Limited |
| Prefixado com Juros Semestrais | No | No | No | No | No |
| IPCA+ com Juros Semestrais | No | No | No | No | No |
| RendA+ | No | No | No | No | No |
| Educa+ | No | No | No | No | No |

### 4.1 Why Tesouro Selic is limited in V1

Tesouro Selic is useful to beginners and belongs in the market overview, but its pricing behavior is not equivalent to applying the same yield-shock model used for zero-coupon Prefixado and IPCA+ bonds.

V1 therefore shows its official market data and educational explanation but does not pretend that the same playground model is valid for it.

A Tesouro Selic position may be stored in the local portfolio and valued with the official `PU Base` when available. V1 does not run yield-shock, theoretical hold/exit, duration, or convexity calculations for Tesouro Selic.

---

## 5. Product language rules

English is required from the first implementation milestone through the V1 release for repository documentation, code identifiers/comments, test names, UI copy, CLI help, errors, logs, commit messages, pull requests, issues authored for the project, and release notes. Maintainer-assistant conversations may remain in Portuguese; that conversational language must not leak into authored repository content.

Preserve official instrument/institution names, exact source field names, quoted source titles, and raw official fixtures in their original form for identity and provenance. Explain them in English: for example, `Tesouro Prefixado — fixed-rate bond` and `Base unit price (source field: PU Base Manha)`. Do not translate raw CSV headers or alter source values to satisfy the language policy. V1 has a single English interface and does not require a localization framework or language switcher.

Display monetary values with an explicit `BRL` currency code, decimal point, and English grouping (for example, `BRL 11,420.31`). Rates use a decimal point (`7.5%`), and displayed dates use unambiguous ISO `YYYY-MM-DD`. User inputs use decimal points without grouping separators, matching CLI/URL conventions; use ISO date inputs. Source adapters still parse Brazilian dates and comma-decimal numbers exactly as documented by the source. This changes presentation, not currency, calendar, source data, or financial conventions.

### 5.1 Simple mode

Simple mode is the default.

It uses questions and consequences rather than jargon.

Preferred labels:

- `What is the value at the latest official quote?`
- `What happens if the yield changes?`
- `Why did the price change?`
- `How much would this bond respond to a yield change?`
- `What does IPCA + 7.5% mean?`

Avoid jargon as the primary label.

Example:

```text
How sensitive is this bond to yield changes?

High sensitivity

Technical detail: modified duration = 4.38
```

### 5.2 Advanced mode

Advanced mode exposes, where applicable:

- bond identifier/type;
- maturity;
- quote date;
- buy yield;
- sell yield;
- buy PU;
- sell PU;
- base PU;
- business days to maturity;
- annualized yield convention;
- scenario yield delta in basis points;
- theoretical scenario PU;
- gross scenario return;
- duration / modified duration when implemented and validated;
- assumptions;
- formula and intermediate values.

### 5.3 Terminology rule

Never replace the official technical term. Introduce it after the plain-language explanation.

Good:

```text
Value at the official quote dated 2026-09-04
BRL 11,420.31

This is the position's gross mark-to-market value using the official base unit price (PU Base, D0).
```

Bad:

```text
MTM: 11,420.31
```

with no explanation.

---

## 6. Data sources

### 6.1 Primary source: Tesouro Transparente

The primary market dataset is the official **Taxas dos Títulos Ofertados pelo Tesouro Direto** dataset from Tesouro Transparente.

At the time this specification was written, the official dataset:

- contains daily prices and rates for Tesouro Direto bonds;
- covers data from January 2002 onward;
- is updated daily;
- is exposed through CKAN and CSV resources;
- is licensed under Open Data Commons ODbL.

The importer should discover the current CSV resource using the CKAN package API instead of depending exclusively on a hard-coded resource URL.

Expected source fields currently include concepts equivalent to:

```text
Tipo Titulo
Data Vencimento
Data Base
Taxa Compra Manha
Taxa Venda Manha
PU Compra Manha
PU Venda Manha
PU Base Manha
```

The official metadata assigns different settlement semantics to the three PU fields:

- `PU Compra Manha`: investor purchase price, D+1 settlement;
- `PU Venda Manha`: investor early-redemption price, D+1 settlement;
- `PU Base Manha`: mark-to-market valuation price, D0 settlement, using the sell rate.

These fields are not interchangeable.

The parser must map source column names into internal domain names. Source-column spelling must remain isolated in the datasource package.

### 6.2 Context source: Banco Central do Brasil SGS

V1 may show a small macro context panel using official Banco Central do Brasil SGS data.

Allowed V1 context:

- Selic;
- IPCA historical series when needed for explanation/context.

These values are contextual data only. They must not be transformed into automatic predictions or investment signals.

### 6.3 Source priority

When data conflicts, use the following priority:

1. official Tesouro/BCB source data;
2. official methodological documentation;
3. local cached copy with provenance;
4. bundled demo fixture;
5. derived calculation.

Never silently replace missing official data with a guessed value.

---

## 7. Data modes

V1 has two operating modes.

### 7.1 Demo mode

Must work without internet access.

The binary embeds a compact, curated dataset sufficient to exercise:

- current-market screen;
- historical chart;
- Prefixado scenario;
- IPCA+ scenario;
- sample portfolio position;
- simple/advanced explanations.

Demo values must be clearly labeled:

```text
Demo data
```

The demo dataset must contain source/provenance metadata and respect source licensing requirements.

Each demo process initializes a private in-memory SQLite database from the embedded fixtures, using the same migrations and normalized schema as synced mode. It must not open, migrate, seed, or modify the user's persistent database. Demo portfolio edits are temporary and disappear when that process ends; the UI must explain this before accepting edits.

The active source is fixed for the lifetime of the server: `demo` or `synced`. Demo requests cannot trigger official synchronization or switch to the persistent database. `analyze --source demo` loads the same versioned embedded fixtures into its own private in-memory database, without network access. Restarting a demo restores its original fixtures.

### 7.2 Synced mode

The application downloads the official source, parses it, validates it, and upserts normalized records into the local SQLite database.

Sync must be:

- explicit;
- idempotent;
- retryable;
- non-destructive;
- provenance-aware.

A failed sync must not corrupt or delete the previously valid dataset.

Synced mode never seeds demo quotes or sample positions into the persistent database. An empty database shows an explicit empty state with instructions to synchronize or start demo mode; it does not silently fall back to demo data.

---

## 8. Local-first requirements

### 8.1 Runtime

A release user installs only the Tesouro Lab binary.

Required runtime dependencies:

```text
none
```

Internet access is required only for live-data synchronization.

### 8.2 Persistence

Use SQLite through a pure-Go driver so the project does not require CGO or a separately installed SQLite library.

Recommended driver:

```text
modernc.org/sqlite
```

### 8.3 Web assets

All templates, CSS, JavaScript, migrations, and small demo fixtures must be embedded in the executable using `go:embed` when practical.

No CDN may be required at runtime.

If a chart library is used, its production asset must be vendored/embedded and its license preserved.

### 8.4 Network binding

Default HTTP bind:

```text
127.0.0.1:8080
```

Do not bind to `0.0.0.0` by default.

### 8.5 Privacy

V1 has:

- no telemetry;
- no analytics;
- no account;
- no cloud sync;
- no remote portfolio storage.

Portfolio data stays on the user's machine.

---

## 9. Technology choices

### 9.1 Backend

- Go;
- standard `net/http`;
- `html/template`;
- `database/sql`;
- pure-Go SQLite driver;
- explicit SQL migrations.

### 9.2 Frontend

- server-rendered HTML;
- CSS;
- minimal vanilla JavaScript;
- embedded charting library if needed.

V1 does **not** require:

- React;
- Vue;
- Svelte;
- Vite;
- Node.js;
- npm/pnpm/yarn;
- a separate frontend build pipeline.

### 9.3 Infrastructure

V1 does **not** require:

- Docker;
- Kubernetes;
- Redis;
- PostgreSQL;
- message queues;
- microservices;
- external cache;
- background worker infrastructure.

A Dockerfile may only be added later as an optional convenience if it does not become the primary installation path.

---

## 10. Proposed repository layout

V1 is a monorepo: application code, financial calculations, CLI, browser templates/assets, migrations, fixtures, tests, and documentation live together and share one release lifecycle. Use one root Go module (`go.mod`). Package boundaries provide separation inside the application; do not create separate frontend/backend repositories, multiple Go modules, a `go.work` workspace, or monorepo orchestration tooling without a concrete V1 requirement.

```text
tesouro-lab/
├── AGENTS.md
├── README.md
├── CONTRIBUTING.md
├── LICENSE
├── go.mod
├── go.sum
├── cmd/
│   └── tesouro-lab/
│       └── main.go
├── internal/
│   ├── app/
│   ├── bond/
│   │   ├── bond.go
│   │   ├── prefixado.go
│   │   ├── ipca.go
│   │   └── selic.go
│   ├── calendar/
│   │   ├── business_days.go
│   │   └── holidays.go
│   ├── datasource/
│   │   ├── tesouro/
│   │   └── bcb/
│   ├── pricing/
│   │   ├── prefixado.go
│   │   ├── ipca.go
│   │   ├── duration.go
│   │   └── scenario.go
│   ├── portfolio/
│   ├── storage/
│   │   └── sqlite/
│   ├── sync/
│   └── web/
├── web/
│   ├── templates/
│   └── static/
├── data/
│   └── demo/
├── migrations/
├── docs/
│   └── V1_SPEC.md
└── tests/
    └── fixtures/
```

Package boundaries are more important than this exact directory tree. Do not split packages merely to match the diagram.

---

## 11. Domain model

### 11.1 Bond

```go
type BondKind string

const (
    BondPrefixado BondKind = "prefixado"
    BondIPCA      BondKind = "ipca"
    BondSelic     BondKind = "selic"
)

type Bond struct {
    ID       string
    Kind     BondKind
    Name     string
    Maturity time.Time
}
```

Bond identity must be stable and independent of display text.

Recommended logical identity:

```text
<kind>:<maturity-date>
```

Example:

```text
prefixado:2032-01-01
```

### 11.2 Market quote

```go
type MarketQuote struct {
    BondID      string
    QuoteDate   time.Time // source column "Data Base"
    BuyYield    *float64
    SellYield   *float64
    BuyPU       *float64  // D+1
    SellPU      *float64  // D+1
    BasePU      *float64  // D0 mark-to-market PU
    Source      string
    ImportedAt  time.Time
}
```

`QuoteDate` is the market date to which the official quote refers. `ImportedAt` is only the local synchronization timestamp and must never be presented as the market date.

Nullable numeric values are required because the upstream dataset may contain missing values for some instrument/date combinations.

### 11.3 Portfolio position

```go
type Position struct {
    ID            string
    BondID        string
    PurchaseDate  *time.Time
    Quantity      float64
    PurchasePU    *float64
    PurchaseYield *float64
    InvestedBRL   *float64
    Note          string
}
```

Minimum position input is `BondID + Quantity`.

Quantity is the canonical basis for current market valuation. Acquisition fields enrich the position but are not required merely to calculate current official MTM.

### 11.4 Scenario

```go
type ScenarioBasis string

const (
    ScenarioPurchase     ScenarioBasis = "purchase"
    ScenarioMarkToMarket ScenarioBasis = "mark_to_market"
    ScenarioEarlyExit    ScenarioBasis = "early_exit"
)

type YieldScenario struct {
    BondID         string
    Basis          ScenarioBasis
    QuoteDate      time.Time
    SettlementDate time.Time
    BaseYield      float64
    BasePU         float64
    ScenarioYield  float64
    BusinessDays   int
}
```

Scenario results must preserve all input values used in the calculation so they can be shown to the user.

---

## 12. SQLite schema

Suggested logical tables:

### `bonds`

```text
id TEXT PRIMARY KEY
kind TEXT NOT NULL
name TEXT NOT NULL
maturity DATE NOT NULL
first_seen DATE
last_seen DATE
```

### `market_quotes`

```text
bond_id TEXT NOT NULL
quote_date DATE NOT NULL
buy_yield REAL
sell_yield REAL
buy_pu REAL
sell_pu REAL
base_pu REAL
source TEXT NOT NULL
imported_at DATETIME NOT NULL
source_hash TEXT
PRIMARY KEY (bond_id, quote_date)
```

### `portfolio_positions`

```text
id TEXT PRIMARY KEY
bond_id TEXT NOT NULL
purchase_date DATE
quantity REAL NOT NULL
purchase_pu REAL
purchase_yield REAL
invested_brl REAL
note TEXT
created_at DATETIME NOT NULL
updated_at DATETIME NOT NULL
```

### `macro_observations`

```text
series TEXT NOT NULL
date DATE NOT NULL
value REAL NOT NULL
source TEXT NOT NULL
PRIMARY KEY (series, date)
```

### `sync_runs`

```text
id TEXT PRIMARY KEY
source TEXT NOT NULL
started_at DATETIME NOT NULL
finished_at DATETIME
status TEXT NOT NULL
records_read INTEGER NOT NULL DEFAULT 0
records_written INTEGER NOT NULL DEFAULT 0
error_message TEXT
```

### `schema_migrations`

Use monotonic versioned migrations.

---

## 12.1 Temporal and freshness semantics

V1 must distinguish four dates/timestamps that are easy to confuse:

- **today**: the user's current local calendar date; never a market quote by itself;
- **quote date**: the official source `Data Base` to which rates/PUs refer;
- **settlement date**: the date used by the pricing convention for the selected quote context;
- **imported at**: when the local application downloaded/imported the source.

### Latest available quote

`latest available quote` means the record with the greatest `quote_date` available for that bond in the active local dataset. It does **not** mean a quote generated today.

Simple mode must display the quote date beside every current-looking market or portfolio value. Preferred copy:

```text
Value at the official quote dated 2026-09-04
```

The word `today` may appear only when the displayed official `quote_date` is the user's current local date. Even then, the exact date remains visible.

### Dataset freshness

For each successful sync store:

```text
imported_at
dataset_max_quote_date
```

For each bond distinguish:

- `latest`: bond quote date equals `dataset_max_quote_date`;
- `older_quote`: bond has a quote, but its latest quote date is older than `dataset_max_quote_date`;
- `no_quote`: no usable quote exists;
- `matured`: the user's current local date is on or after the bond maturity date.

If the whole dataset is older than expected, the UI must show the actual `dataset_max_quote_date` and `imported_at`; it must never silently relabel old data as current.

V1 may warn that the dataset appears stale when it is at least **two Brazilian financial-market business days** behind the most recent base date reasonably expected from the source publication cadence. A one-business-day discrepancy is not enough by itself to declare the source stale because publication timing can vary.

### Missing recent quote

If the active dataset contains newer market dates but a specific bond does not, do not forward-fill, interpolate, or copy another bond's quote. Show the bond's last quote date and mark it as older/unavailable for current valuation.

### Matured bonds

Matured bonds are excluded from the default current-market list but remain available in historical views.

For a matured portfolio position:

- do not report a stale pre-maturity quote as current market value;
- show status `Matured on YYYY-MM-DD`;
- V1 does not reconcile the actual cash settlement into a cash account;
- for Prefixado, the contractual nominal maturity amount may be shown separately as an educational/historical value (`quantity × BRL 1,000.00`), never as a current live quote.

Historical and demo scenarios evaluate remaining term at their selected quote/settlement dates, not at the machine's current date. A bond that has matured today may still have a valid historical scenario, explicitly labeled with its historical quote date.

---

## 13. Numeric policy

V1 is an educational analytical tool, not a settlement or accounting system.

Use `float64` for pricing/yield calculations, with the following constraints:

- never compare floats with exact equality in financial tests;
- define explicit tolerance;
- preserve original source values before transformation when useful;
- round only for presentation unless the official methodology explicitly requires truncation;
- when validating against official PU, reproduce the official truncation/rounding convention where documented;
- show no more precision than the source or calculation justifies.

Do not claim cent-perfect settlement values unless the implementation and all relevant costs have been validated for that purpose.

Reject non-finite numeric inputs (`NaN`, positive/negative infinity). Required PUs, quantities, and monetary input amounts must be strictly positive; optional missing inputs remain absent. Both baseline and simulated yields must be finite and greater than -100%. Derived quantities, unit prices, and gross valuation/maturity amounts must also remain finite and positive; gains/losses and variations may be zero or negative but must be finite. If arithmetic overflows, underflows to an invalid result, or otherwise produces a non-finite result, return `calculation_out_of_range` instead of displaying it or clamping it to zero.

---

## 14. Market valuation and quote-context rules

The official dataset exposes three different price contexts. V1 must preserve the distinction.

### 14.1 Current portfolio mark-to-market

For a non-matured existing position, current gross mark-to-market uses the latest available official **PU Base**:

```text
current_gross_mtm = quantity × base_pu
```

`PU Base` is a D0 valuation price and is the official field intended to mark Tesouro Direto holdings to market.

Do not use `PU Venda` as the portfolio MTM field when `PU Base` is available.

### 14.2 Hypothetical purchase

For a user exploring a purchase quote:

```text
base_yield = official buy yield
base_pu    = official buy PU
settlement = next Brazilian financial-market business day (D+1)
```

### 14.3 Hypothetical early exit

For a user exploring an early redemption:

```text
base_yield = official sell yield
base_pu    = official sell PU
settlement = next Brazilian financial-market business day (D+1)
```

This is not the same as portfolio MTM.

### 14.4 Portfolio yield-shock scenario

For a same-date scenario applied to an existing position:

```text
base_yield = official sell yield
base_pu    = official base PU
settlement = quote date (D0)
```

This keeps the scenario anchored to the official mark-to-market value rather than to a hypothetical redemption value.

### 14.5 No quote substitution

If the required PU/yield pair for a context is missing, that context is unavailable. Do not silently substitute buy, sell, or base PU for another context.

---

## 15. Zero-coupon scenario pricing contract

Prefixado and IPCA+ without coupons share a useful same-date repricing property: when the valuation/indexation base is held constant, changing only yield changes the discount factor.

V1 scenarios are **anchored to an official PU/yield pair** instead of rebuilding the current quote from scratch.

For a supported zero-coupon bond:

```text
PU_scenario = PU_base × ((1 + base_yield) / (1 + scenario_yield))^(DU/252)
```

where:

- `PU_base` = official PU for the selected quote context;
- `base_yield` = official yield paired with that context;
- `scenario_yield` = hypothetical annual yield as decimal;
- `DU` = Brazilian financial-market business days from settlement date **inclusive** to maturity date **exclusive**;
- annual convention = 252 business days.

This anchored form is intentional. It guarantees that a zero-yield shock returns the official baseline PU and prevents small differences in source-rate precision, VNA reconstruction, or truncation rules from being misrepresented as market movement.

Required invariant:

```text
scenario_yield == base_yield  =>  PU_scenario == PU_base within numerical tolerance
```

The scenario engine must reject yields `<= -100%`.

Scenario eligibility requires settlement strictly before maturity and a verified business-day count `DU > 0`. Otherwise return `no_remaining_term`, including when D+1 reaches maturity even though the quote date precedes it. Never substitute one business day or continue with a zero/negative exponent. The zero-shock and strict price/yield invariants apply only to eligible inputs. The numeric validation rules in section 13 apply before and after calculation.

---

## 16. Prefixado calculation

For Tesouro Prefixado without coupon, the standalone theoretical relation used for validation and explanation is:

```text
PU_theoretical = 1000 / (1 + y)^(DU/252)
```

where `DU` follows the settlement-inclusive, maturity-exclusive convention above.

Use the standalone relation for:

- educational formula display;
- deterministic unit tests;
- validation against official historical fixtures.

Use the anchored equation from section 15 for interactive scenarios.

Official Tesouro methodology states that negotiated PUs are truncated to two decimal places. Validation against an official historical Prefixado fixture therefore uses this contract:

```text
abs(truncate_2_decimals(PU_theoretical) - official_PU) <= BRL 0.01
```

Do not widen this tolerance globally to make a failing fixture pass. A failure requires investigation of calendar, settlement context, source-rate precision, or fixture provenance.

---

## 17. IPCA+ same-date calculation

Tesouro IPCA+ without coupon uses an inflation-linked VNA plus a real-yield discount factor.

### 17.1 Current portfolio MTM

Use official `PU Base` directly.

No VNA reconstruction is required merely to answer the current mark-to-market value of a position.

### 17.2 Same-date real-yield scenario

For V1, preserve the official PU as the inflation/indexation base and alter only the real-yield discount factor using the anchored equation from section 15.

This is valid for the V1 question:

```text
If the real yield were X on this same date, holding the indexation base constant, how would the price respond?
```

It is **not** a future inflation forecast and does not reconstruct future VNA.

### 17.3 Future inflation

V1 does not silently project IPCA.

Any feature that would require an unknown future inflation path is unsupported unless the specification explicitly defines an input assumption for that feature.

---

## 17.4 Business-day and settlement convention

V1 uses a deterministic Brazilian financial-market calendar.

Rules:

- weekends are not business days;
- holidays come from a versioned local calendar fixture based on the ANBIMA national financial-market holiday calendar;
- pricing day count is from settlement date **inclusive** to maturity date **exclusive**;
- D+1 means the next date considered a business day by that calendar;
- `PU Compra` and `PU Venda` scenarios use D+1 settlement;
- `PU Base` mark-to-market scenarios use D0 settlement (`settlement_date = quote_date`);
- no live calendar dependency is required at scenario runtime.

The calendar fixture must expose provenance and covered years. If a requested date falls outside the verified calendar coverage, the calculation returns `calendar_out_of_range`; it must not guess holidays.

---

## 18. Scenario engine

### 18.1 Default shocks

For supported bonds:

```text
-400 bps
-300 bps
-200 bps
-100 bps
   0 bps
+100 bps
+200 bps
+300 bps
+400 bps
```

The user may also enter a custom scenario yield.

### 18.2 Rule

A scenario means:

> hold the valuation date and other modeled variables constant; change the specified yield assumption; calculate the corresponding theoretical price.

It does not mean:

> this is where the market will trade.

### 18.3 Required scenario output

Simple mode:

```text
Official baseline yield: 14.23%
Simulated yield:         12.23%

Official baseline price: BRL X
Simulated price:         BRL Y
Scenario change:         +Z%
Official quote date:     YYYY-MM-DD
```

Advanced mode additionally shows:

```text
yield delta
basis points
business days
pricing convention
formula
input PU
calculated PU
calculation timestamp/source date
```

---

## 19. Portfolio V1

Portfolio is local and manual.

### 19.1 Minimum input and completeness states

Minimum required fields:

```text
Bond
Quantity
```

Optional acquisition fields:

```text
Purchase date
Gross acquisition amount
Purchase unit price
Purchase yield
Note
```

The application supports two explicit completeness states.

#### `valuation_only`

Available when title + quantity are known.

Can show:

- latest official quote date;
- current gross MTM using `quantity × PU Base`;
- current sell quote when available;
- hypothetical same-date yield scenarios for supported instruments.

Cannot show:

- amount originally invested;
- gross gain/loss since purchase;
- purchase-to-current performance.

Simple-mode copy:

```text
Change since purchase is unavailable.
Enter the gross acquisition amount or purchase unit price to calculate this comparison.
```

#### `cost_basis_known`

Available when a gross acquisition basis can be determined.

Accepted ways to determine it:

```text
quantity + invested_brl
quantity + purchase_pu
invested_brl + purchase_pu  -> derive quantity
```

If `quantity + purchase_pu` are supplied and `invested_brl` is absent:

```text
invested_brl = quantity × purchase_pu
```

If `invested_brl + purchase_pu` are supplied and quantity is absent at form input:

```text
quantity = invested_brl / purchase_pu
```

The normalized stored position always has a positive quantity.

If all three values are supplied, they must satisfy:

```text
abs(invested_brl - quantity × purchase_pu)
<= max(BRL 0.02, invested_brl × 0.0001)
```

Otherwise reject the input and ask the user which value is authoritative. V1 does not silently absorb brokerage fees/taxes into this discrepancy.

`invested_brl` means gross acquisition value of the bonds for V1 purposes; it is not a tax-adjusted cost basis.

### 19.2 Historical quote is not acquisition data

Do not infer a user's actual purchase PU or invested amount from the daily historical dataset merely because a purchase date is known. Daily official quotes may not equal the user's exact execution terms.

Historical data may be shown for context, clearly labeled as a market quote, but never silently promoted to the user's acquisition cost.

### 19.3 Position screen

Simple mode always shows when available:

- latest official quote date;
- current gross market value using official `PU Base`;
- current sell quote separately when available;
- plain-language explanation;
- scenario shortcut for supported instruments.

Only when cost basis is known, also show:

- amount invested;
- gross change since acquisition basis.

Advanced mode additionally shows when available:

- quantity;
- purchase date;
- purchase PU;
- purchase yield;
- current base PU (D0);
- current sell PU (D+1);
- current sell yield;
- quote date;
- scenario settlement date;
- technical scenario details.

### 19.4 Tesouro Selic in portfolio

Tesouro Selic positions support official current MTM using `PU Base` and quote/history display.

V1 does not calculate Tesouro Selic yield-shock scenarios, theoretical duration/convexity, or hold-vs-early-exit break-even.

### 19.5 No tax engine in V1

V1 explicitly labels portfolio/early-exit values as **gross** unless a value comes directly from an official source already net of a specific item.

Do not imply that gross market value equals cash received after all taxes/costs.

---

## 20. Hold vs early-exit view

This feature is **required in V1 for Tesouro Prefixado positions only**.

It is unsupported in V1 for Tesouro IPCA+ and Tesouro Selic because a robust future-horizon comparison would require additional indexation/model assumptions or instrument-specific logic outside the locked scope.

The view is a gross mathematical comparison, never a recommendation.

### Path A — hold Prefixado to maturity

For a quantity `q`:

```text
gross_nominal_maturity_amount = q × BRL 1,000.00
```

### Path B — hypothetical early exit on the quote date

Use the official sell PU (D+1), not PU Base:

```text
gross_early_exit_value = quantity × official_sell_pu
```

Always show the official quote date.

### Break-even reinvestment

Calculate the annualized gross reinvestment rate that would make the gross early-exit value reach the Prefixado nominal maturity amount over the remaining D+1-to-maturity business-day period:

```text
break_even_rate = (maturity_amount / early_exit_value)^(252 / DU) - 1
```

where `DU` is counted from D+1 settlement inclusive to maturity exclusive.

The comparison requires a positive finite quantity and official sell PU, positive finite maturity and early-exit amounts, settlement before maturity, and verified `DU > 0`. Apply the same `no_remaining_term` and `calculation_out_of_range` rules as the scenario engine. A contractual maturity amount may still be explained separately when the comparison is unavailable; do not display an annualized break-even rate for a zero or negative remaining term.

Required disclaimer:

```text
Gross hypothetical comparison. Taxes, fees, and actual reinvestment conditions may change the outcome.
```

Do not emit:

```text
You should sell
Sell
Hold
Best option
```

---

## 21. Dashboard information architecture

V1 has four primary areas.

### 21.1 Market

Purpose:

> show what government bonds are available in the latest synchronized dataset and explain the quote.

Default columns/cards use the latest available **buy-side** quote because this screen answers what is being offered to an investor exploring a purchase:

- name;
- maturity;
- official quote date;
- buy rate;
- buy PU;
- plain-language bond type.

Advanced toggle adds the raw buy/sell/base quote fields and their D0/D+1 semantics.

### 21.2 Playground

Purpose:

> make yield/price sensitivity visible interactively.

Controls:

- bond selector;
- investment amount;
- yield slider/custom yield;
- default basis-point shock buttons;
- simple/advanced toggle.

Outputs:

- official base PU vs simulated PU;
- official base amount vs simulated amount;
- percent change;
- yield/price chart;
- explanation.

### 21.3 Portfolio

Purpose:

> let users enter local positions and see official current gross mark-to-market.

No remote sync.

### 21.4 Learn

Purpose:

> provide contextual, interactive explanations tied to actual app data.

Topics limited to V1 concepts:

- Prefixado;
- IPCA+;
- Tesouro Selic;
- price vs yield;
- mark-to-market;
- maturity;
- buy quote vs sell quote;
- basis points;
- inflation-linked return;
- gross vs net;
- scenario vs prediction.

---

## 22. Simple/advanced mode contract

Simple mode and advanced mode must render the **same underlying calculation**.

They are presentation layers, not different engines.

Never allow a simplified explanation to use a different formula merely because it is easier to describe.

Every advanced result should be traceable to the same domain result used by simple mode.

---

## 23. Explainability contract

Every calculated result must be able to answer:

1. What data was used?
2. What source/date did it come from?
3. What assumption did the user provide?
4. What formula/model was used?
5. Is the value official, derived, estimated, or hypothetical?

Recommended result metadata:

```go
type Provenance struct {
    Kind       string // official | derived | scenario | demo
    Source     string
    SourceDate time.Time
    Notes      []string
}
```

---

## 24. CLI

The browser UI is the primary user interface. The CLI exists for developers and reproducibility.

### Required commands

```text
tesouro-lab
```

Starts the local server using the existing database, initializing it if needed.

```text
tesouro-lab sync
```

Synchronizes official data.

```text
tesouro-lab demo
```

Starts the app with embedded demo data and no required network access.

```text
tesouro-lab analyze <bond-id> --yield <percent> [--amount <brl>] [--source demo|synced] [--basis purchase|mark_to_market|early_exit] [--date YYYY-MM-DD]
```

Runs the same scenario engine used by the UI and prints a deterministic result.

### Scenario selection and reproducibility

- `--source` defaults to `synced`, reading existing local data without synchronizing. `demo` selects the embedded fixtures under section 7's isolation rules.
- `--basis` defaults to `purchase`. The selected basis determines the PU/yield pair and D0/D+1 settlement under section 14; never infer it from whichever fields happen to be present.
- `--date` is an exact official quote date in ISO format. If omitted, resolve the latest record for that bond in the selected source and print the resolved date. If no record exists for the explicitly requested date, or the selected record lacks the required context fields, return `missing_quote`; never search backward for a usable pair or switch sources.
- `--yield 12.00` means 12% per year, normalized to `0.12` once at the delivery boundary. CLI numbers, URL parameters, and browser numeric inputs use a decimal point, without currency symbols, percent signs, or grouping separators. Browser labels identify the units; all delivery mechanisms normalize inputs before calling the same application function.
- `--amount` is an optional positive gross baseline value in BRL. Derive the hypothetical quantity as `amount / selected_official_pu`, without intermediate rounding. If omitted, report unit-price results only. A portfolio scenario uses the stored quantity directly; it does not reconstruct it from a displayed, rounded amount.

The standalone playground uses the same default basis; portfolio shortcuts explicitly select `mark_to_market`, and early-exit scenario shortcuts explicitly select `early_exit`. Unsupported instruments fail according to the support matrix regardless of delivery mechanism.

Every result must expose the resolved source, bond, basis, quote date, settlement date, baseline PU/yield, simulated yield, DU, optional amount/quantity, calendar version, and calculation/build version. Demo results also identify the fixture version. UI technical details must provide an equivalent CLI invocation with explicit source, basis, date, yield, and amount when applicable, using unrounded round-trip numeric representations.

Reproducibility means equal numeric inputs, quote contents, calendar version, and calculation version produce equal results within the documented tolerances. A command that selects the latest quote is intentionally relative to local data. Even a fixed date can change after an official source correction; the result must preserve the actual baseline values used so this is inspectable. V1 does not promise archival replay across overwritten source revisions or changed calendar/build versions, and does not add a versioned quote archive for that purpose.

### Optional V1 command

```text
tesouro-lab version
```

Prints build/version information.

Do not build a large CLI framework unless the standard library becomes materially inadequate.

---

## 25. HTTP/UI behavior

### 25.1 Server

- local bind only by default;
- sensible read/write/header timeouts;
- graceful shutdown;
- no external session service;
- no auth in V1;
- local state only.

### 25.2 Browser state

Simple/advanced preference may be persisted locally in the browser.

Do not create a server-side user profile.

### 25.3 URLs

Standalone scenario selection must be representable in query parameters using the same names and units as the CLI (`bond`, `yield`, `amount`, `source`, `basis`, `date`). Basis/date defaults match the CLI; an omitted browser `source` uses the running server's active source. Generated scenario links must include the resolved source, basis, and quote date instead of leaving them implicit.

Example:

```text
/playground?bond=prefixado:2032-01-01&yield=12.00&amount=10000&source=demo&basis=purchase&date=2026-09-04
```

This is a syntax example, not a claim that the named bond/date exists in the fixtures. Missing records fail explicitly. The source parameter must match the running server's active source; a mismatch returns `source_mismatch` with instructions to start the matching mode. A URL cannot switch databases, trigger sync, or load an arbitrary file. Portfolio notes and acquisition details must not be included in generated scenario URLs.

This supports reproducibility within the data/calendar/version limits in section 24 without requiring cloud storage.

---

## 26. Sync behavior

### 26.1 Tesouro sync sequence

```text
resolve CKAN package
      ↓
locate active CSV resource
      ↓
download with timeout
      ↓
validate content type/size
      ↓
parse source locale
      ↓
normalize instrument names/dates/numbers
      ↓
validate rows
      ↓
transactional upsert
      ↓
record sync run
```

### 26.2 Parser requirements

The source has historically used Brazilian formatting conventions such as semicolon separators and comma decimals. The importer must handle source formatting explicitly and test it using fixtures.

Do not depend on the OS locale.

### 26.3 Idempotency

Running `sync` multiple times against the same source must not duplicate quotes.

Natural uniqueness:

```text
bond_id + quote_date
```

### 26.4 Source changes

Unknown columns may be ignored only when safe.

Missing required columns must fail the sync with a clear error while preserving the previous database state.

---

## 27. Error model

User-facing errors must separate:

### Data unavailable

```text
No official quote is available for this date.
```

### Unsupported calculation

```text
This calculation is not supported for this bond type in V1.
```

### Invalid assumption/input

```text
The simulated yield must be greater than -100%.
```

### Sync failure

```text
The update failed. Previously stored local data is still available.
```

### No remaining term (`no_remaining_term`)

```text
There are no remaining business days between settlement and maturity for this scenario.
```

### Calculation outside numeric range (`calculation_out_of_range`)

```text
The result is outside the supported numeric range. Review the scenario inputs.
```

Invalid numeric inputs return `invalid_input`; missing required quote fields return `missing_quote`. Neither case may be converted into a numeric result. Calendar coverage errors remain `calendar_out_of_range`.

Never convert an unknown into zero.

---

## 28. Validation and testing strategy

Financial calculations must be test-driven against fixed official historical examples/fixtures.

### 28.1 Unit tests

Required areas:

- source number/date parser;
- bond normalization;
- business-day counter;
- Prefixado pricing;
- IPCA+ same-date yield-shock calculation;
- scenario return;
- portfolio MTM;
- break-even calculation;
- input validation.

### 28.2 Golden/fixture tests

Store small source fixtures derived from official historical data with attribution.

Tests must confirm that parsed records match expected normalized values.

### 28.3 Pricing validation

For historical Prefixado quote fixtures where the official PU and yield are known:

- use the quote context's settlement convention (D0 or D+1);
- count settlement inclusive and maturity exclusive;
- apply the official two-decimal PU truncation;
- require absolute difference from official PU `<= BRL 0.01`;
- do not loosen the global tolerance to hide a mismatch.

For anchored scenario tests:

```text
zero shock: abs(PU_scenario - PU_base) <= 1e-9
```

For each analytically supported instrument, commit at least one official baseline fixture with two hypothetical nonzero shocks (one higher yield and one lower yield). Check expected PU magnitude and gross percentage variation, not only direction. Each fixture must record:

- official source URL, retrieval date, original quote values, quote context, and source date;
- independently checked settlement date and expected DU, with calendar provenance/version;
- explicit hypothetical yields, expected unrounded scenario PUs and variations, and documented absolute/relative tolerances;
- an adjacent worked calculation or independent high-precision reference calculation used to obtain those expectations.

Expected outputs must be fixed, reviewable values, not generated at test runtime by the production pricing or calendar functions. For IPCA+, the reference calculation must preserve a single indexation factor while independently discounting at each real yield. Distinguish official baseline data from derived hypothetical expectations; these are not official future quotes. Include D0 and D+1 context coverage across the fixtures and a weekend/holiday settlement boundary. Changing DU by one day must be detectable by the chosen nonzero-shock assertions.

For internal non-money floating calculations use a documented relative/absolute tolerance appropriate to the calculation; do not use exact float equality.

A fixture outside verified holiday-calendar coverage must fail with `calendar_out_of_range`, not fall back to weekday-only counting.

No financial formula is considered complete merely because the code compiles.

### 28.4 Invariants

For supported zero-coupon scenario calculations with finite valid inputs and `DU > 0`:

```text
higher yield  => lower price
lower yield   => higher price
same yield    => same theoretical price within tolerance
```

For a fixed bond/date:

```text
PU(y1) > PU(y2) when y1 < y2
```

Boundary tests must cover settlement at/after maturity, D+1 reaching maturity, zero/negative DU, out-of-coverage calendars, and historical scenarios for bonds that have matured today. Validate rejection of nonpositive PU/quantity/amount, yields at/below -100%, non-finite inputs, and non-representable outputs. Include `DU = 0` in break-even tests and verify an explicit error rather than division by zero.

### 28.5 Storage tests

- migrations from empty DB;
- idempotent upsert;
- failed transaction rollback;
- no duplicate quote keys;
- demo startup/edits/restart leave a pre-existing synced database and its portfolio unchanged;
- demo also works when the persistent data location is inaccessible, proving it is not required;
- restarting demo restores its fixtures and discards temporary edits.

### 28.6 HTTP smoke tests

- app starts;
- dashboard renders;
- demo mode works without network;
- scenario endpoint returns valid result;
- unsupported title does not invoke unsupported math.

### 28.7 CLI/UI reproducibility

- explicit source, basis, date, yield, and amount produce matching normalized results through CLI and browser;
- `12.00` is normalized to `0.12` in canonical inputs;
- omitted date resolves and reports the latest record; an explicit missing date or missing context pair fails without fallback;
- a newer quote changes a latest-date request but not a fixed-date request while that record remains unchanged;
- demo CLI and demo UI use the same fixtures without network access or persistent-database access;
- a source mismatch or sync request in demo cannot mutate persistent state;
- generated CLI instructions and scenario URLs retain the resolved context and unrounded numeric inputs.

---

## 29. Build and release requirements

Release users should not compile source code.

GitHub Releases should eventually publish binaries for at least:

```text
linux/amd64
linux/arm64
windows/amd64
darwin/amd64
darwin/arm64
```

Each release should include checksums.

From M1 onward, the README must distinguish implemented behavior from planned milestones and document the release and source demo commands, prerequisites, URL, and shutdown behavior. Keep contributor setup and checks in `CONTRIBUTING.md`, and financial methodology/assumptions beside the relevant specification sections and fixtures. Add usage documentation and tests with each implemented feature rather than deferring them to M5. Release notes must state what is implemented, validation performed, known limitations, and any migration or compatibility impact.

Once the Go module exists, automated checks on pull requests must verify formatting, run `go vet ./...` and `go test ./...`, and check a build with `CGO_ENABLED=0`. Run race tests where supported. Release validation must exercise the documented demo command from an extracted package with no Go installation or network requirement, and verify the declared OS/architecture targets before publishing them as supported. The source demo command must also be checked from the repository root. Do not claim these checks exist or pass before they are implemented and run.

The build must not require CGO for official release targets.

Primary local developer checks:

```bash
go fmt ./...
go vet ./...
go test ./...
```

Run race tests where supported/useful:

```bash
go test -race ./...
```

---

## 30. Security constraints

Although this is a local app, basic security still applies.

- bind only to loopback by default;
- use HTTP timeouts;
- cap remote download size;
- validate CSV/imported files;
- do not execute user-supplied content;
- escape HTML through templates;
- limit portfolio import size;
- never store secrets because V1 should not need secrets;
- no arbitrary URL fetch endpoint;
- no hidden telemetry.

---

## 31. Performance targets

V1 does not need premature optimization.

Reasonable targets on a normal desktop:

- cold start with existing DB: < 2 s preferred;
- dashboard query: < 200 ms preferred after startup;
- scenario calculation: effectively instant (< 50 ms expected);
- full official CSV sync: bounded by download/parse speed, not user interaction;
- memory use appropriate for a local desktop utility.

Do not add caches until measurements justify them.

---

## 32. Accessibility and UX baseline

- responsive desktop-first UI;
- keyboard-accessible controls;
- labels for form inputs;
- charts must have numeric/text equivalents;
- do not communicate gains/losses only by color;
- English UI copy, explicit BRL currency, decimal-point numbers, and ISO dates as defined in section 5;
- technical/raw values may expose canonical representations in advanced mode.

---

## 33. Wording policy: no recommendation language

Forbidden product wording in V1:

The prohibition applies to advice, imperatives, and rankings, not to descriptive terminology. Labels such as `purchase quote`, `sell unit price`, and `Simulate early redemption`, and technical identifiers such as `BuyPU` and `SellYield`, are allowed. Do not implement a blanket ban on individual words that would prevent explaining official quote contexts.

```text
buy now
sell now
best bond
worst bond
investment opportunity
worth buying
recommended investment
entry signal
exit signal
target yield
guaranteed profit potential
```

Allowed analytical wording:

```text
current yield
current price
scenario
hypothesis
estimated change in this scenario
gross market value
mathematical comparison
sensitivity
if the yield were X, holding the other scenario conditions constant...
```

The application may describe a historical fact but must not convert it into advice.

---

## 34. V1 acceptance criteria

V1 is complete when all items below are true.

### Packaging

- [ ] release runs as a single binary;
- [ ] no Node/npm runtime;
- [ ] no external DB;
- [ ] no Docker requirement;
- [ ] assets are local/embedded;
- [ ] default bind is loopback;
- [ ] documented release and source demo commands start the complete app in one process;
- [ ] the release demo needs no runtime installation, manual setup, or network access.

### Data

- [ ] demo mode works offline;
- [ ] demo state is temporary and isolated from the persistent database and portfolio;
- [ ] official Tesouro dataset sync works;
- [ ] sync is idempotent;
- [ ] source provenance is visible;
- [ ] every current-looking market/portfolio value shows its official quote date;
- [ ] stale/older/matured quote states never masquerade as current data;
- [ ] previous valid data survives failed sync.

### Instruments

- [ ] Prefixado market/history supported;
- [ ] IPCA+ market/history supported;
- [ ] Tesouro Selic market/history and official-PU-Base portfolio MTM supported;
- [ ] unsupported instrument types are explicitly identified rather than miscalculated.

### Analysis

- [ ] Prefixado scenario engine validated;
- [ ] IPCA+ same-date real-yield scenario validated;
- [ ] nonzero-shock magnitude is checked against independent expected values for both analytical instruments;
- [ ] invalid numeric inputs and scenarios without remaining term fail explicitly;
- [ ] portfolio uses official PU Base for current gross MTM;
- [ ] Prefixado gross hold/early-exit comparison is implemented and does not imply advice;
- [ ] no silent inflation forecast;
- [ ] no tax/net-return claim.

### UX

- [ ] simple mode default;
- [ ] advanced mode exposes technical details;
- [ ] both modes use the same calculation result;
- [ ] CLI and browser reproduce explicit scenario source/date/basis within documented data/version limits;
- [ ] explanations identify official vs derived vs hypothetical values;
- [ ] basic market, playground, portfolio, and learning areas work;
- [ ] all authored UI, CLI, help, errors, and documentation are in English;
- [ ] original source labels and fixtures retain their provenance, with English explanations.

### Engineering

- [ ] `go test ./...` passes;
- [ ] `go vet ./...` passes;
- [ ] source parsing has fixtures;
- [ ] pricing has official-data validation fixtures;
- [ ] SQLite migrations are versioned;
- [ ] no unnecessary V2 infrastructure exists;
- [ ] application, assets, fixtures, tests, and documentation share one repository and root Go module;
- [ ] automated checks, contributor instructions, and release notes reflect the implemented behavior.

---

## 35. Explicit V2 boundary

The V1 repository must not be designed around a hypothetical SaaS backend.

If the project later becomes an online product, V2 may add separate services around the V1 domain engine.

Potential V2 areas, intentionally excluded now:

- hosted accounts;
- synced portfolios;
- watchlists;
- saved cloud scenarios;
- alerts;
- richer historical curve analytics;
- curve-relative comparisons;
- ETTJ integration;
- premium reports;
- multi-device state.

V1 should remain useful even if V2 never exists.

---

## 36. Architectural principle

The financial/domain engine must not depend on HTTP, HTML, SQLite, or CLI packages.

Preferred dependency direction:

```text
                   ┌─────────────┐
                   │   pricing   │
                   │    bond     │
                   │  scenario   │
                   └──────▲──────┘
                          │
             ┌────────────┼────────────┐
             │            │            │
         portfolio     web/app        cli
             ▲            ▲            ▲
             │            │            │
          storage     datasource     commands
```

The diagram is conceptual. Avoid dependency cycles and keep deterministic financial logic isolated from delivery/storage concerns.

---

## 37. Definition of a trustworthy result

A Tesouro Lab result is trustworthy only when:

1. the input source is known;
2. the source date is known;
3. missing data is not invented;
4. the implemented instrument is supported by the engine;
5. assumptions are explicit;
6. the formula is tested;
7. the result type is labeled as official, derived, or hypothetical;
8. uncertainty/limitations are visible.

Correct behavior when the application does not know is:

```text
A reliable result cannot be calculated with the available data.
```

not a guessed number.

---

## 38. Official references used to define V1

Primary references to validate implementation and update documentation:

- Tesouro Transparente — dataset “Taxas dos Títulos Ofertados pelo Tesouro Direto”  
  https://www.tesourotransparente.gov.br/ckan/dataset/taxas-dos-titulos-ofertados-pelo-tesouro-direto

- Tesouro Transparente — metadata for the daily prices/rates dataset (`Taxa.pdf`), including D0/D+1 PU semantics  
  https://tesourotransparente.gov.br/ckan/dataset/df56aa42-484a-4a59-8184-7676580c81e3/resource/1a8eb2e3-4902-4a38-a1eb-6410f23d90de/download/Taxa.pdf

- Tesouro Direto — methodology for Tesouro Prefixado (LTN)  
  https://www.tesourodireto.com.br/documents/d/guest/tesouro_prefixado

- Tesouro Direto — methodology for Tesouro IPCA+ / NTN-B Principal  
  https://www.tesourodireto.com.br/documents/d/guest/tesouro_ipca_juros_semestrais

- Banco Central do Brasil — SGS public time-series system  
  https://www3.bcb.gov.br/sgspub/

- BCB Open Data — Selic annualized base 252 (SGS 1178)  
  https://dadosabertos.bcb.gov.br/dataset/1178-taxa-de-juros---selic-anualizada-base-252

- ANBIMA — Brazilian financial-market holiday calendar  
  https://www.anbima.com.br/feriados/

If a source changes, update the datasource adapter and references without leaking source-specific changes into the pricing/domain packages.
