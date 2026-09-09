# Next steps toward V1

Last updated: 2026-09-09.

This is a continuation checklist, not a replacement for the [V1 specification](V1_SPEC.md). Follow [AGENTS.md](../AGENTS.md) when implementing changes. Unchecked items are pending; this document does not authorize publishing releases or changing V1 scope.

## Current checkpoint

M1 implements the offline Prefixado demo: one Go module, embedded assets and fixtures, in-memory SQLite, versioned migration, isolated pricing/calendar logic, server-rendered English UI, and shared CLI/API calculations.

M2.1 adds persistent database startup, migration reuse, a local empty state, and fixed server-source isolation. See [M2 implementation notes](M2_IMPLEMENTATION.md).

M2.2 verifies the official datasource contract and adds offline parser fixtures; no synchronization is enabled yet.

See [M1 implementation notes](M1_IMPLEMENTATION.md) for behavior and validation already performed, and [fixture provenance](../data/demo/README.md) for sources and assumptions.

Important limitations to preserve until explicitly addressed:

- The fixture is an official historical methodology example, not a live CKAN quote.
- Only the purchase-context PU/yield pair is available; other contexts return `missing_quote`.
- Verified calendar coverage is 2012–2015. Do not infer business days outside that range.
- Synchronization, synchronized analysis, IPCA+, Selic, and portfolios are not implemented. Persistent mode currently exposes a database status/empty page only.
- CI is configured but a successful GitHub run has not been verified. No release has been published.
- Local checks covered macOS/arm64. Cross-platform release execution and automated browser interactions remain unverified.

## Next implementation milestone: M2 — official data

Proceed in small, independently tested changes:

1. **Persistent local storage — completed (M2.1).** Normal startup opens `tesouro-lab.db` under the OS user configuration directory or `--data-dir`. Migrations are reused, demo remains isolated, and an empty database shows an explicit empty state. Tests cover reopens, failed migrations, invalid paths/databases, CLI shutdown, and HTTP source isolation.
2. **Official datasource contract — completed (M2.2).** Verified CKAN API, CSV, and metadata; added raw attributed fixtures and a bounded mixed-dataset parser. M2 accepts only Prefixado and reports excluded names/counts. See [datasource contract](OFFICIAL_DATASOURCE.md).
3. **Explicit synchronization.** Implement `sync` with timeouts, bounded downloads, required-field validation, provenance, transactional upserts, and separately recorded sync status/errors. A failed download, parse, or write must preserve previous valid quotes. Do not add background synchronization or arbitrary URL inputs.
4. **Verified calendars and quote contexts.** Expand calendar coverage using official ANBIMA sources and independent pricing fixtures. Validate purchase (`BuyPU` + `BuyYield`, D+1), mark-to-market (`BasePU` + `SellYield`, D0), and early redemption (`SellPU` + `SellYield`, D+1). Missing pairs remain unavailable.
5. **History, dates, and delivery.** Expose synchronized quotes/history and the same scenario service through CLI and browser. Show official quote dates, distinguish import timestamps, and handle empty, older, unavailable, and matured states without forward-filling or substituting demo data.

M2 completion checks:

- [ ] Repeating a sync does not duplicate quotes (`bond_id` + `quote_date`).
- [ ] Failed sync preserves existing valid data and records an explanatory status.
- [ ] Demo never accesses persistent storage or starts a live sync.
- [ ] Exact-date and missing-context requests never silently select another quote.
- [ ] D0/D+1, holidays, maturity, and calendar bounds have deterministic tests with independent expectations.
- [ ] Routine tests remain offline; any live integration checks are separate and explicit.
- [ ] README, CLI help, and implementation notes match available behavior.

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
2. Start with M2.3: explicit synchronization with bounded official downloads, transactional upserts, and separately recorded status/errors unless the maintainer changes priorities. Do not pull later-milestone infrastructure forward.
3. Keep all repository content and application copy in English; maintainer conversations may remain in Portuguese.
4. Preserve the single-binary Go architecture, loopback default, transparent math, and no-recommendation policy. Do not introduce a frontend runtime, mandatory Docker, cloud services, or V2 features.
5. For code changes, run `go fmt ./...`, `go vet ./...`, `go test ./...`, and `go test -race ./...` where supported. Verify builds without CGO when relevant.
6. Update this checklist and implementation notes with completed work, exact validation, and remaining limitations. Keep historical verification separate from newly performed checks.
