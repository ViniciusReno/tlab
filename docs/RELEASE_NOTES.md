# Release notes

## Unreleased — M6

Browser screens, explanations, forms, and application validation messages now support English and Brazilian Portuguese. English remains the default; an explicit language link saves a local browser preference. Language changes preserve selected scenario inputs, including after asynchronous updates. No-JavaScript navigation and forms remain available. Numeric/date entry instructions are visible in both languages; canonical BRL, decimal-point, and ISO formats remain unchanged.

No dependencies, Go-version changes, migrations, financial formulas, calendar/settlement changes, or CLI/API changes. Existing data and private notes are preserved. Local checks and bilingual Chrome journeys passed; see `docs/M6_IMPLEMENTATION.md`. Remote CI and native release validation remain pending. This increment is not published.

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
