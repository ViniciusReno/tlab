# Next steps toward V1 and M6

Last updated: 2026-10-08.

This is a continuation checklist, not a replacement for the [V1 specification](V1_SPEC.md). Follow [AGENTS.md](../AGENTS.md) when implementing changes. Unchecked items are pending; this document does not authorize publishing releases or changing V1 scope.

## Current checkpoint

M1 implements the offline Prefixado demo: one Go module, embedded assets and fixtures, in-memory SQLite, versioned migration, isolated pricing/calendar logic, server-rendered English UI, and shared CLI/API calculations.

M2.1 adds persistent database startup, migration reuse, a local empty state, and fixed server-source isolation. See [M2 implementation notes](M2_IMPLEMENTATION.md).

M2.2 verifies the official datasource contract and adds offline parser fixtures. M2.3 implements explicit `sync`, bounded official downloads, transactional upserts, provenance, and separately persisted failure status. M2.4 adds the verified 2002–2032 calendar and independent quote-context validation; its original early-redemption discrepancy is resolved by the approved M2 closure amendment below. M2.5a adds the persistent market/history UI, explicit quote/import dates, and maturity/freshness states. M2.5b adds synchronized purchase/base scenarios through the shared CLI/API/browser service. M2.5c adds conservative dataset-age warnings using verified business days.

See [M1 implementation notes](M1_IMPLEMENTATION.md) for behavior and validation already performed, and [fixture provenance](../data/demo/README.md) for sources and assumptions.

## M2 completion

M2 is implemented under the maintainer-approved [settlement decision](M2_SETTLEMENT_DECISION.md). Official morning early-redemption scenarios use SellPU/SellYield with D+1 before 2021-09-13 and D0 thereafter. Each selected Prefixado redemption record must pass standalone PU validation before anchored repricing. Historical-calendar mismatches remain explicitly unavailable; no calendar or price substitution is performed.

- [x] Idempotent official synchronization, transaction rollback, and separate failure status.
- [x] Persistent market/history with exact quote dates, provenance, and maturity/freshness states.
- [x] Dataset-age warnings using verified business days and documented publication assumptions.
- [x] Purchase, mark-to-market, and validated morning early-redemption scenarios through shared CLI/API/browser calculations.
- [x] Exact-date selection and missing-context errors without fallback.
- [x] Independent transition and recent quote fixtures, both shock directions, historical mismatch gates, and calendar/maturity boundaries.
- [x] Source isolation, offline demo, and reproducible commands with explicit settlement version.
- [x] Documentation, formatting, vet, tests, race checks, and local build validation.

Remaining limitations are explicit: the Prefixado demo has purchase quotes only; the IPCA+ demo has all three contexts. Both retain the 2012–2015 calendar; M3 extends synced coverage to 2002–2050. Historical calendar reconstruction, after-13:00 requests, suspended trading, and actual execution are not modeled. Mismatching redemption records return `calculation_not_validated`. IPCA+/Selic support is described under M3 below; portfolios remain M4. Remote CI, automated browser interaction, cross-platform execution, and release publication remain unverified.

M3 now provides IPCA+ scenarios within verified calendar coverage and Selic quotes. See [M2 implementation notes](M2_IMPLEMENTATION.md) for the chronological work and checks.

## Following milestones

### M3 — IPCA+ and Selic

- [x] Add independently validated IPCA+ same-date real-yield scenarios and an official demo fixture.
- [x] Preserve a single indexation factor within each scenario; never silently assume future inflation.
- [x] Add Selic market/history support and official-PU-Base valuation support for later portfolio use, without yield-shock or hold/early-exit analysis.
- [x] Keep coupon-bearing and other unsupported instruments explicitly unsupported.
- [x] Expand verified calendar coverage through 2050 with official annual ANBIMA tables, preserving prior fixture versions and adding a long-maturity official IPCA+ scenario fixture.

See [M3 implementation notes](M3_IMPLEMENTATION.md). Quantity-based portfolio valuation is implemented in M4 below.

### M4 — local portfolio

- [x] Implement positions with title and positive quantity; acquisition cost is optional.
- [x] Use official `PU Base` for gross current mark-to-market value, with explicit quote dates and matured/missing states.
- [x] Keep acquisition-dependent results unavailable when cost is unknown.
- [x] Implement the gross Prefixado hold-versus-hypothetical-early-exit comparison using official `PU Venda` and the documented remaining-term rules.
- [x] Preserve temporary demo edits and persistent/demo separation.

See [M4 implementation notes](M4_IMPLEMENTATION.md) for migration, acquisition normalization, gross valuation, scenario quantities, and validation limits.

### M5 — V1 closure

- [x] Add Learn with the V1 concepts, expandable explanations, official-source attribution, and historical exercises using existing Playground calculations.
- [x] Connect the primary areas through navigation and add skip-to-content links; correct obsolete portfolio availability copy.
- [x] Complete and review Market, Playground, Portfolio, and Learn against the specification.
- [x] Verify keyboard access, narrow layouts, no-JavaScript operation, slider/shock interactions, error recovery, and numeric chart equivalents in Chrome on macOS.
- [ ] Verify CI on GitHub after an authorized push; resolve failures before release.
- [ ] Build and execute supported release targets, including an extracted offline demo without Go installed.
- [x] Prepare local checksums, dependency/source attribution, candidate release notes, launch instructions, and migration/compatibility notes.
- [x] Audit every V1 acceptance criterion with evidence and explicit remote/native release gates; V1 is not yet declared complete.

M5 learning increment (2026-09-30): `/learn` is embedded and available offline in both modes, without JavaScript or a new dependency. The demo offers six explicit historical scenario links covering both analytical instruments, zero/down/up Prefixado shocks, and IPCA+ base-PU context. Persistent mode points to local quote selection without loading demo data or silently choosing a date. Native disclosure controls expose formulas and limits; the existing Playground service owns all numeric results. Explanations cover every topic in specification section 21.4, including Selic restrictions, inflation assumptions, quote dates, maturity, and gross versus net values.

Validation for this increment: `go fmt ./...`, `go vet ./...`, `go test ./internal/web`, `go test ./...`, `go test -race ./...`, `CGO_ENABLED=0 go build -trimpath -o /private/tmp/tlab-m5-learn ./cmd/tesouro-lab`, and `git diff --check`. The build succeeded with a sandbox warning about writing Go's module metadata cache. HTTP tests follow all local learning anchors and exercise links, verify completed scenario responses and navigation in both modes, reject POST to Learn, and confirm the persistent database remains empty. Interactive browser/keyboard checks, narrow-screen review, remote CI, cross-platform execution, and release packaging are still pending; these HTTP checks do not close those acceptance items.

M5 closure increment (2026-10-08): reviewed all four areas in Chrome with demo and isolated persistent fixtures, at 375px and 1280px. Added baseline-specific Prefixado/IPCA+ sliders and a persistent bond selector, preserved results-region announcements/focus, and made portfolio input conventions explicit. Fixed real-browser portfolio POST rejection caused by `no-referrer` turning Origin into `null`; same-origin referrer policy keeps external disclosure blocked. CLI help now reflects M4.

Added standard-library-only maintainer packaging and smoke scripts, third-party license notices, five-target native CI, and candidate release documentation. All five targets cross-built without CGO and their archive hashes were checked. Extracted macOS arm64 and amd64 (Rosetta) demos ran with empty PATH and external network connections denied. See [M5 acceptance audit](M5_ACCEPTANCE.md) for evidence and open gates. Remote CI and native Linux/Windows execution remain pending; compilation alone does not close them.

### M6 — Brazilian Portuguese and usability (planned after V1)

Maintainer direction (2026-09-30): make Brazilian Portuguese support and broader usability improvements an explicit M6 milestone, after M5 closes the existing V1 plan. This supersedes the unnumbered deferral recorded on 2026-09-28. M6 is planned, not implemented; no localization infrastructure or redesign is introduced now. The accessibility and basic UX acceptance checks already specified for M5 remain required.

- [ ] Add Brazilian Portuguese (pt-BR) support to user-facing screens, explanations, forms, and validation messages, retaining English support.
- [ ] Define language selection and localized display/input conventions for dates, numbers, and BRL amounts before implementation; preserve canonical CLI/API inputs, raw official records, and reproducible calculations.
- [ ] Review the first-use experience, navigation, and clarity of Market, Playground, Portfolio, and Learn with concrete user tasks.
- [ ] Improve portfolio entry/editing, scenario controls, and empty/error/unavailable states based on that review, while preserving transparent quote dates and gross-value labels.
- [ ] Validate both languages, keyboard and narrow-screen use, and the main user journeys; verify that presentation changes leave financial results unchanged.
- [ ] Update user documentation for the implemented language and usability changes.

M6 may introduce the localization needed for this approved post-V1 scope. Code identifiers, developer documentation, commits, and raw official data retain their existing language/provenance rules. The local-first single-binary architecture, instrument support boundaries, and no-recommendation policy remain unchanged.

## Resume checklist

1. Read this checkpoint, the relevant specification sections, and existing code/tests; do not assume this snapshot is still current.
2. M2 is complete under the approved settlement contract. M3 IPCA+ scenarios, Selic quote support, and verified 2002–2050 calendar coverage are implemented. M4 local portfolio is implemented. M5 Learn and integrated navigation are implemented; continue the remaining screen, accessibility, acceptance, and release checks. Preserve the Prefixado redemption validation gate and the IPCA+ fixed-indexation contract.
3. Through M5, keep repository content and application copy in English; maintainer conversations may remain in Portuguese. After V1 closure, M6 adds the approved user-facing pt-BR support without translating raw source records or code identifiers.
4. Preserve the single-binary Go architecture, loopback default, transparent math, and no-recommendation policy. Do not introduce a frontend runtime, mandatory Docker, cloud services, or V2 features.
5. For code changes, run `go fmt ./...`, `go vet ./...`, `go test ./...`, and `go test -race ./...` where supported. Verify builds without CGO when relevant.
6. Update this checklist and implementation notes with completed work, exact validation, and remaining limitations. Keep historical verification separate from newly performed checks.
