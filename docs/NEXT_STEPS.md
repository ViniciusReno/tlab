# Next steps toward V1

Last updated: 2026-09-28.

This is a continuation checklist, not a replacement for the [V1 specification](V1_SPEC.md). Follow [AGENTS.md](../AGENTS.md) when implementing changes. Unchecked items are pending; this document does not authorize publishing releases or changing V1 scope.

## Current checkpoint

M1 implements the offline Prefixado demo: one Go module, embedded assets and fixtures, in-memory SQLite, versioned migration, isolated pricing/calendar logic, server-rendered English UI, and shared CLI/API calculations.

M2.1 adds persistent database startup, migration reuse, a local empty state, and fixed server-source isolation. See [M2 implementation notes](M2_IMPLEMENTATION.md).

M2.2 verifies the official datasource contract and adds offline parser fixtures. M2.3 implements explicit `sync`, bounded official downloads, transactional upserts, provenance, and separately persisted failure status. M2.4 adds the verified 2002–2032 calendar and independent quote-context validation; its original early-redemption discrepancy is resolved by the approved M2 closure amendment below. M2.5a adds the persistent market/history UI, explicit quote/import dates, and maturity/freshness states. M2.5b adds synchronized purchase/base scenarios through the shared CLI/API/browser service. M2.5c adds conservative dataset-age warnings using verified business days.

See [M1 implementation notes](M1_IMPLEMENTATION.md) for behavior and validation already performed, and [fixture provenance](../data/demo/README.md) for sources and assumptions.

## M2 completion

M2 is implemented under the maintainer-approved [settlement decision](M2_SETTLEMENT_DECISION.md). Official morning early-redemption scenarios use SellPU/SellYield with D+1 before 2021-09-13 and D0 thereafter. Each selected record must pass standalone PU validation before anchored repricing. Historical-calendar mismatches remain explicitly unavailable; no calendar or price substitution is performed.

- [x] Idempotent official synchronization, transaction rollback, and separate failure status.
- [x] Persistent market/history with exact quote dates, provenance, and maturity/freshness states.
- [x] Dataset-age warnings using verified business days and documented publication assumptions.
- [x] Purchase, mark-to-market, and validated morning early-redemption scenarios through shared CLI/API/browser calculations.
- [x] Exact-date selection and missing-context errors without fallback.
- [x] Independent transition and recent quote fixtures, both shock directions, historical mismatch gates, and calendar/maturity boundaries.
- [x] Source isolation, offline demo, and reproducible commands with explicit settlement version.
- [x] Documentation, formatting, vet, tests, race checks, and local build validation.

Remaining limitations are explicit: the demo has purchase quotes only and retains its 2012–2015 calendar; synced coverage is 2002–2032. Historical calendar reconstruction, after-13:00 requests, suspended trading, and actual execution are not modeled. Mismatching redemption records return `calculation_not_validated`. IPCA+, Selic, and portfolios remain later milestones. Remote CI, automated browser interaction, cross-platform execution, and release publication remain unverified.

The next implementation milestone is M3. See [M2 implementation notes](M2_IMPLEMENTATION.md) for the chronological work and checks.

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
2. M2 is complete under the approved settlement contract. Continue with M3 IPCA+ fixtures and same-date real-yield scenarios; retain the per-record Prefixado redemption gate and do not extend that validation to IPCA+ without instrument-specific evidence.
3. Keep all repository content and application copy in English; maintainer conversations may remain in Portuguese.
4. Preserve the single-binary Go architecture, loopback default, transparent math, and no-recommendation policy. Do not introduce a frontend runtime, mandatory Docker, cloud services, or V2 features.
5. For code changes, run `go fmt ./...`, `go vet ./...`, `go test ./...`, and `go test -race ./...` where supported. Verify builds without CGO when relevant.
6. Update this checklist and implementation notes with completed work, exact validation, and remaining limitations. Keep historical verification separate from newly performed checks.
