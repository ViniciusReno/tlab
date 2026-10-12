# Planned free scope — closure record

Date: 2026-10-12. M1–M6 implementation, validation, and publication are complete.
Stable [v1.0.0](https://github.com/ViniciusReno/tlab/releases/tag/v1.0.0) is public. No additional features are required to satisfy the approved plan.
The repository remains available for maintenance and contributions.

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
- [x] Stable `v1.0.0` publication and final repository checkpoint.

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

## Stable publication

[v1.0.0](https://github.com/ViniciusReno/tlab/releases/tag/v1.0.0) was published
on 2026-10-12 at 21:35 UTC, neither a draft nor a prerelease. The tag and embedded
Git revision resolve to `2cbe402108456a4c8e28a2e204d9dca8fea83646`. The
[Go checks](https://github.com/ViniciusReno/tlab/actions/runs/38309935739) and
[exact-version five-target native checks](https://github.com/ViniciusReno/tlab/actions/runs/38309943069)
passed for that release commit after the acceptance documentation merged.

Five platform archives and `SHA256SUMS` were uploaded. All six GitHub asset
SHA-256 digests matched the local files; every archive contains launch/release
notes and project/dependency/source attribution. The final extracted macOS
archives also passed bilingual smoke checks on default port 8080 under OS-level
external-network denial. Go build metadata confirmed the tag revision, CGO
disabled, and `vcs.modified=false`. Subsequent publication-checkpoint changes
are documentation only and do not alter the shipped software.

The original empty, unrelated npm lockfile was preserved outside the repository
and is absent from source/release artifacts. The application has no Node/npm
runtime. Implementation and release documentation were committed and merged
through protected-main pull requests. There is no remaining planned milestone.

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
