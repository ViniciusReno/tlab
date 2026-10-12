# Release notes

## 1.0.0 — 2026-10-12

Browser screens, explanations, forms, and application validation messages now support English and Brazilian Portuguese. English remains the default; an explicit language link saves a local browser preference. Language changes preserve selected scenario inputs, including after asynchronous updates. No-JavaScript navigation and forms remain available. Numeric/date entry instructions are visible in both languages; canonical BRL, decimal-point, and ISO formats remain unchanged.

No dependencies, Go-version changes, migrations, financial formulas, calendar/settlement changes, or CLI/API changes. Existing data and private notes are preserved. Local checks and bilingual Chrome journeys passed; see `docs/M6_IMPLEMENTATION.md`. The [Go checks](https://github.com/ViniciusReno/tlab/actions/runs/38308467948) and [five-target native checks with version 1.0.0](https://github.com/ViniciusReno/tlab/actions/runs/38308470160) passed for M6 merge commit `5752a97`. Both extracted macOS builds additionally passed with empty PATH and OS-level external-network denial. A live official sync to an isolated temporary database succeeded on 2026-10-12: 177,260 source records read, 68,577 supported records stored, maximum official quote date 2026-10-09. The initial free scope is complete; see the project closure record for publication status.

This stable release includes all M1–M6 capabilities: official quote/history synchronization, offline demos, transparent Prefixado/IPCA+ scenarios, Selic quote/base-PU valuation, local portfolios, Prefixado gross comparisons, and Learn. Linux amd64/arm64, Windows amd64, and macOS amd64/arm64 packages ship SHA-256 checksums. Values remain gross; coupon-bearing instruments, Selic yield scenarios, tax/net calculations, inflation forecasts, brokerage integration, and recommendations remain outside scope. Packages are unsigned/unnotarized; screen-reader certification and other browser engines were not evaluated.

## 0.1.0 — 2026-10-09

[Published V1](https://github.com/ViniciusReno/tlab/releases/tag/v0.1.0), tag commit `1d4c64d`. Packaged notes retain the pre-publication candidate record; the GitHub release notes record publication validation.

Implemented behavior:

- Offline historical Prefixado and IPCA+ demos with temporary portfolio edits.
- Explicit official synchronization, transactional/idempotent quote storage,
  history, provenance, and dated freshness/maturity states.
- Shared CLI/browser zero-coupon scenarios with exact source/date/context and
  independently validated financial fixtures.
- Local quantity-based portfolios, optional acquisition cost, official gross
  base-PU valuation, and validated Prefixado hold/early-redemption comparisons.
- Learn, connected navigation, keyboard skip links, numeric chart tables, and
  yield sliders based on the selected official yield for both analytical instruments.
- Local release packaging, SHA-256 checksums, and extracted-binary smoke checks.

Compatibility: no dependency or Go-version upgrade, no new migration, no changes
to financial formulas, settlement rules, or calendar versions. Existing M4
databases are reused. Build versions are stamped by the packager and exposed by
`version` and scenario results. Python is needed only for maintainer packaging
and smoke scripts, never by a release user.

Known limits: English-only V1, gross values without taxes/fees, no future inflation
assumption, no coupon-bearing analytical support, and no Selic yield scenarios.
Prefixado's demo has purchase quotes only. Verified calendars cover 2012–2015 in
demo and 2002–2050 in synced mode. Prefixado early-redemption records that fail
standalone validation remain unavailable. After-13:00 requests, suspended trading,
actual execution, and reconciliation of maturity cash are not modeled.

Validation evidence and remaining release gates are tracked in
`docs/M5_ACCEPTANCE.md` in the source repository. The [Go workflow](https://github.com/ViniciusReno/tlab/actions/runs/37848148816) and
[five-target native release workflow](https://github.com/ViniciusReno/tlab/actions/runs/37848148688)
passed for code commit `84caae5` on 2026-10-08. Local formatting, vet, unit/race
tests, Chrome demo/persistent journeys, and network-isolated extracted macOS
demos also passed. The merge commit also passed both workflows, as recorded in the M5 audit.
All five platform archives and SHA-256 checksums were published. No package
is signed or notarized.
