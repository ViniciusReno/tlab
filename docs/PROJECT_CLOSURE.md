# Planned free scope — closure record

Date: 2026-10-12. M1–M6 implementation and validation are complete. The final
administrative step is publishing stable release `v1.0.0`; no additional features
are required to satisfy the approved plan. The repository remains available for
maintenance and contributions.

## Completed scope

- [x] One local Go binary, embedded assets, pure-Go SQLite, offline isolated demo.
- [x] Official synchronization, history, quote dates, provenance, freshness, and explicit missing/matured states.
- [x] Validated Prefixado and IPCA+ scenarios shared by browser and CLI.
- [x] Selic quotes/history and official base-PU valuation within the specified limits.
- [x] Local quantity-based portfolios, optional acquisition cost, gross valuation, and Prefixado hold/early-redemption comparison.
- [x] Market, Playground, Portfolio, and Learn with transparent assumptions.
- [x] English/pt-BR browser interface, canonical number/date formats, keyboard and narrow-screen review, and no-JavaScript workflows.
- [x] Formatting, vet, unit/race tests, CGO-free builds, and five-target extracted-binary checks.
- [x] License/source/dependency attribution, launch instructions, migration notes, and documented limitations.
- [ ] Stable `v1.0.0` publication and final repository checkpoint.

## Evidence

M6 and its release checks merged through [PR #1](https://github.com/ViniciusReno/tlab/pull/1)
at `5752a97e83fc94141738ef4b7469d5e033b0c62b`. The
[Go checks](https://github.com/ViniciusReno/tlab/actions/runs/38308467948) and
[version 1.0.0 native checks](https://github.com/ViniciusReno/tlab/actions/runs/38308470160)
passed. Native runners verified Linux amd64/arm64, Windows amd64, and macOS
amd64/arm64 archive checksums, extraction, startup, embedded English/Portuguese
pages, both analytical instruments, canonical API build/date fields, and shutdown
without Go or another runtime on PATH.

Local `go fmt ./...`, `go vet ./...`, `go test ./...`, and `go test -race ./...`
passed (cached Go tests). All five local archives cross-built with CGO disabled;
their hashes passed. Extracted macOS arm64/amd64 demos ran under OS-level external
network denial, at isolated port 18085; amd64 used Rosetta. The source demo
started offline on port 18086. Browser tasks, translation coverage, source
isolation, exact result parity, and form safety are recorded in
[M6 implementation](M6_IMPLEMENTATION.md); original financial/acceptance evidence
is in [M5 acceptance](M5_ACCEPTANCE.md).

A live sync against the official Tesouro Transparente datasource succeeded in a
temporary database on 2026-10-12: 177,260 records read, 68,577 supported records
stored, dataset maximum official quote date 2026-10-09. Unsupported rows were
explicitly counted and excluded. No private portfolio or persistent user data
was used. Market dates are not inferred from the date of this check.

## Deliberate limits

The completed scope is an educational application, with no recommendations,
predictions, brokerage integration, accounts, paid features, or cloud service.
Coupon-bearing instruments and other asset classes are unsupported. Selic has
no yield-shock/duration/hold-exit analysis. Values are gross without a complete
tax/cost engine. IPCA+ has no implicit inflation forecast. Calendar coverage and
Prefixado redemption validation gates remain explicit.

Packages are unsigned/unnotarized. Chrome keyboard/layout checks are not formal
accessibility certification; screen readers and other browser engines were not
evaluated. These are disclosed limits, not unimplemented features in the plan.
No new dependency, migration, financial formula, or frontend runtime was added.

Further feature development requires a new maintainer decision. Maintenance can
address defects, official-source changes, verified calendar updates, and portability
issues while preserving the scope and single-binary architecture.
