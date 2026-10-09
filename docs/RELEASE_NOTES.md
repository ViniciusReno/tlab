# V1 release candidate — unpublished

This is a candidate validation record, not an announcement of a published release.

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
demos also passed. Review and publication remain maintainer actions. No package
is signed or notarized, and no publication has been performed.
