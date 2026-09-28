# Next steps toward V1

Last updated: 2026-09-28.

This is a continuation checklist, not a replacement for the [V1 specification](V1_SPEC.md). Follow [AGENTS.md](../AGENTS.md) when implementing changes. Unchecked items are pending; this document does not authorize publishing releases or changing V1 scope.

## Current checkpoint

M1 implements the offline Prefixado demo: one Go module, embedded assets and fixtures, in-memory SQLite, versioned migration, isolated pricing/calendar logic, server-rendered English UI, and shared CLI/API calculations.

M2.1 adds persistent database startup, migration reuse, a local empty state, and fixed server-source isolation. See [M2 implementation notes](M2_IMPLEMENTATION.md).

M2.2 verifies the official datasource contract and adds offline parser fixtures. M2.3 implements explicit `sync`, bounded official downloads, transactional upserts, provenance, and separately persisted failure status. M2.4 adds the verified 2002–2032 calendar and independent quote-context validation; recent early-redemption validation remains unresolved. M2.5a adds the persistent market/history UI, explicit quote/import dates, and maturity/freshness states. M2.5b adds synchronized purchase/base scenarios through the shared CLI/API/browser service. M2.5c adds conservative dataset-age warnings using verified business days.

See [M1 implementation notes](M1_IMPLEMENTATION.md) for behavior and validation already performed, and [fixture provenance](../data/demo/README.md) for sources and assumptions.

Important limitations to preserve until explicitly addressed:

- The fixture is an official historical methodology example, not a live CKAN quote.
- The demo has only the purchase-context PU/yield pair; its other contexts return `missing_quote`. Synchronized data preserves all three official PU fields, with purchase/base scenarios now available. All synchronized early-redemption scenarios remain blocked pending settlement validation.
- The original demo retains 2012–2015 calendar coverage. The new embedded calendar covers 2002–2032. Do not infer business days outside the selected version’s range.
- Recent SellPU rows fail standalone validation under the locked D+1 rule. Preserve the specification until an explicit maintainer decision; see [the evidence and limitations](../internal/pricing/testdata/README.md#known-sourcespecification-discrepancy).
- Synchronized early redemption, IPCA+, Selic, and portfolios are not implemented. Persistent mode exposes market/history browsing; official import is available through the CLI. Dataset-age warnings use the previous local business day and a two-business-day threshold; the policy is an explicit assumption, not an official publication guarantee.
- CI is configured but a successful GitHub run has not been verified. No release has been published.
- Local checks covered macOS/arm64. Cross-platform release execution and automated browser interactions remain unverified.

## Next implementation milestone: M2 — official data

Proceed in small, independently tested changes:

1. **Persistent local storage — completed (M2.1).** Normal startup opens `tesouro-lab.db` under the OS user configuration directory or `--data-dir`. Migrations are reused, demo remains isolated, and an empty database shows an explicit empty state. Tests cover reopens, failed migrations, invalid paths/databases, CLI shutdown, and HTTP source isolation.
2. **Official datasource contract — completed (M2.2).** Verified CKAN API, CSV, and metadata; added raw attributed fixtures and a bounded mixed-dataset parser. M2 accepts only Prefixado and reports excluded names/counts. See [datasource contract](OFFICIAL_DATASOURCE.md).
3. **Explicit synchronization — completed (M2.3).** `sync [--data-dir PATH]` discovers and downloads the official CSV, validates all rows, imports Prefixado transactionally, and prints/persists a report with excluded instruments and source dates. Failure preserves existing quotes. Tests cover URL/redirect restrictions, limits, cancellation, migration, rollback, idempotency, and demo isolation. No background synchronization or arbitrary URL inputs.
4. **Verified calendars and quote contexts — implemented with a validation limitation (M2.4).** Added the 381-date ANBIMA 2002–2032 fixture, preserving the original demo calendar/version. The shared pure pricing resolver selects the exact PU/yield pair and D0/D+1 term. Independent 60-digit reference fixtures validate all contexts for 2012/2016 and purchase/base for 2024/2026, with fixed DU, nonzero shocks, missing-input and boundary tests. Two recent early-redemption cases are explicitly `calculation_not_validated`; the locked D+1 rule differs from observed source values. Resolve this contract before enabling recent early-redemption analysis.
5. **History, dates, and delivery — implemented (M2.5a/b), with early redemption blocked.** M2.5a exposes synchronized purchase quotes and paginated history, with advanced buy/sell/base fields, source provenance, and separate quote/import dates. Empty, older, missing, unknown-freshness, and matured states are explicit. M2.5b exposes synchronized purchase/base scenarios through the shared CLI/API/browser service with exact-date selection, explicit defaults, and reproducible inputs. Market/history links select the exact context and date; commands preserve the local data directory. All synchronized early-redemption scenarios remain unavailable until the historical settlement contract is resolved.

6. **Dataset-age warnings — implemented (M2.5c).** The market page warns at a lag of at least two verified business days relative to the previous local business day, retaining actual dataset/import dates. Missing metadata, future dates, and calendar coverage limits remain explicit unavailable states. See [the policy](OFFICIAL_DATASOURCE.md#dataset-age-warning-policy).

M2 completion checks:

- [x] Repeating a sync does not duplicate quotes (`bond_id` + `quote_date`).
- [x] Failed sync preserves existing valid data and records an explanatory status.
- [x] Demo never accesses persistent storage or starts a live sync.
- [x] Exact-date and missing-context requests never silently select another quote.
- [x] D0/D+1, holidays, maturity, and calendar bounds have deterministic tests with independent expectations.
- [x] Dataset-age warnings preserve actual dates and document the cadence assumption.
- [ ] Resolve the recent SellPU settlement discrepancy before enabling those early-redemption scenarios.
- [x] Routine tests remain offline; any live integration checks are separate and explicit.
- [x] README, CLI help, and implementation notes match available behavior.

## M2 closure review

Official transition records and independent D0/D+1 expectations are now preserved. The [settlement decision proposal](M2_SETTLEMENT_DECISION.md) identifies both the 2021 settlement change and additional historical-calendar discrepancies. Maintainer approval of the amended contract (or an explicit milestone exception) is pending. M2 remains open; runtime behavior and the locked specification are unchanged.

## Following milestones

### M3 — IPCA+ and Selic

- [ ] Add independently validated IPCA+ same-date real-yield scenarios and an official demo fixture.
- [ ] Preserve a single indexation factor within each scenario; never silently assume future inflation.
- [ ] Add Selic market/history support and official-PU-Base valuation support for later portfolio use, without yield-shock or hold/early-exit analysis.
- [ ] Keep coupon-bearing and other unsupported instruments explicitly unsupported.

### M4 — local portfolio

- [ ] Implement positions with title and positive quantity; acquisition cost is optional.
- [ ] Use official `PU Base` for gross current mark-to-market value, with explicit quote dates and matured/missing states.
- [ ] Keep acquisition-dependent results unavailable when cost is unknown.
- [ ] Implement the gross Prefixado hold-versus-hypothetical-early-exit comparison using official `PU Venda` and the documented remaining-term rules.
- [ ] Preserve temporary demo edits and persistent/demo separation.

### M5 — V1 closure

- [ ] Complete and review Market, Playground, Portfolio, and Learn against the specification.
- [ ] Verify keyboard access, narrow layouts, no-JavaScript operation, slider/shock interactions, error recovery, and numeric chart equivalents.
- [ ] Verify CI on GitHub after an authorized push; resolve failures before release.
- [ ] Build and execute supported release targets, including an extracted offline demo without Go installed.
- [ ] Prepare checksums, attribution, release notes, launch instructions, and migration/compatibility notes.
- [ ] Audit every V1 acceptance criterion; do not mark V1 complete based only on successful compilation.

## Resume checklist

1. Read this checkpoint, the relevant specification sections, and existing code/tests; do not assume this snapshot is still current.
2. Review the documented recent SellPU/D+1 discrepancy before enabling synchronized early-redemption analysis. Any change to the locked settlement contract requires an explicit maintainer decision supported by official transition fixtures. M2.5b synchronized purchase/base scenario delivery is implemented. The remaining M2 work is settlement-contract investigation; all synchronized redemption scenarios remain unavailable. Do not pull later-milestone infrastructure forward.
3. Keep all repository content and application copy in English; maintainer conversations may remain in Portuguese.
4. Preserve the single-binary Go architecture, loopback default, transparent math, and no-recommendation policy. Do not introduce a frontend runtime, mandatory Docker, cloud services, or V2 features.
5. For code changes, run `go fmt ./...`, `go vet ./...`, `go test ./...`, and `go test -race ./...` where supported. Verify builds without CGO when relevant.
6. Update this checklist and implementation notes with completed work, exact validation, and remaining limitations. Keep historical verification separate from newly performed checks.
