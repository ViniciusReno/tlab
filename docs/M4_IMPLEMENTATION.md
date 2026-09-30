# M4 — Local portfolio

Implemented on 2026-09-30. The browser route `/portfolio` adds manual local positions without new dependencies, external services, or a frontend build step. M5 acceptance and release checks remain separate.

## Position input and persistence

Migration `003_portfolio.sql` adds `portfolio_positions` with a foreign key to normalized bonds and positive-value constraints. Existing migrations and official quotes are unchanged. The storage package performs SQL only; `internal/portfolio` owns deterministic financial normalization and calculations. Application services share these calculations with HTTP presentation.

Required input is bond plus positive quantity. Optional fields are purchase date, purchase PU, gross acquisition amount, purchase yield, and note. Missing quantity can be derived as acquisition amount / purchase PU; missing amount can be derived as quantity × purchase PU. Supplying all three requires absolute difference within `max(BRL 0.02, invested amount × 0.0001)`. Conflicts require the user to clear or correct the conflicting field. No historical quote becomes actual acquisition data. Unknown acquisition cost produces `valuation_only`; a known gross acquisition amount produces `cost_basis_known`. Notes and acquisition details never enter scenario links.

Create, edit, and explicitly confirmed removal use POST forms and redirect after success. Forms cap bodies at 16 KiB, reject non-finite numbers and malformed dates, limit notes to 2,000 bytes, escape HTML, and require a cryptographically random process-local form token. Cross-origin submissions are rejected. A stale token requires reloading the form after a server restart. GET requests do not mutate positions. User notes are not logged. No portfolio import/export format is introduced.

Normal-mode edits persist in the selected local database. Each demo process instead seeds one hypothetical two-unit IPCA+ position in its private in-memory database; acquisition data remains unknown. The UI explains temporary edits before submission. Restarting the demo restores the sample and never changes the persistent portfolio.

## Valuation and explanation

For each supported, non-matured position with official PU Base:

```text
gross market value = stored quantity × official PU Base
```

A known acquisition amount also allows:

```text
gross change = gross market value − gross acquisition amount
gross change fraction = gross market value / gross acquisition amount − 1
```

This is a comparison of gross amounts, not a total-return series or annualized purchase performance. Taxes and fees are excluded. Input and intermediate calculations are not rounded for display.

The overview shows gross value, valuation coverage, composition by supported instrument type, and a position table. Composition percentages describe only positions with available valuation; they are not targets. Partial totals explicitly exclude unavailable positions. The date range beside the total identifies mixed quote dates. A numeric overflow makes the total unavailable instead of presenting infinity. Per-position quote dates, dataset date, and import timestamp remain distinct. The existing conservative dataset-age warning is reused.

Missing BasePU never falls back to BuyPU, SellPU, or a calculated price. Older quotes remain labeled as last available valuations. Future-dated quotes are unavailable for current valuation. Matured positions show their maturity date, not a stale current value; actual maturity cash is not reconciled. The historical demo therefore has no current value after its maturity, but its explicitly historical scenarios remain usable.

## Same-date scenarios

Position scenarios select the exact quote date and `mark_to_market` basis, call the existing shared `Analyze` service, and multiply the resulting scenario PU by the stored quantity directly. They never reconstruct quantity from a displayed or rounded amount. The unit-price inspection link leads to the existing playground, including CLI reproduction and advanced calculation metadata. No acquisition data is included. IPCA+ holds the same-date indexation anchor fixed and uses real yields. Selic scenarios remain `unsupported`.

## Prefixado maturity and early-redemption comparison

This gross comparison is limited to Prefixado. It resolves official SellPU/SellYield using the existing morning settlement convention: D+1 before 2021-09-13, D0 thereafter. The standalone theoretical PU, truncated to cents, must match SellPU within BRL 0.01. A mismatch yields `calculation_not_validated`; the official price and calendar are never fitted or replaced.

```text
nominal maturity amount = quantity × BRL 1,000.00
hypothetical early-redemption amount = quantity × official SellPU
annualized gross break-even rate = (maturity amount / early-redemption amount)^(252 / DU) − 1
```

DU is the verified business-day count from settlement inclusive to maturity exclusive. The implementation uses logarithms for the ratio to reduce intermediate overflow. Positive finite inputs/results and DU > 0 are mandatory. Underflow/overflow yields `calculation_out_of_range`; zero or negative remaining term yields `no_remaining_term`. The rate describes a hypothetical reinvestment condition, not an available offer or recommendation. IPCA+ and Selic return `unsupported` for this comparison. Morning normal-market limitations and the absence of a tax engine remain visible.

Sources and methodology: [V1 sections 19–20](V1_SPEC.md#19-portfolio-v1), [official Prefixado methodology](https://www.tesourodireto.com.br/documents/d/guest/tesouro_prefixado), the [approved settlement decision](M2_SETTLEMENT_DECISION.md), and [official transition fixtures](../internal/pricing/testdata/README.md). Independent financial tests reuse those unchanged official rows and independently verified DU values. At quantity 2.5 and DU 701:

| Official quote date | Settlement | SellPU | Gross early redemption | Annual break-even fraction |
| --- | --- | --- | --- | --- |
| 2021-09-10 | 2021-09-13 (D+1) | BRL 765.36 | BRL 1,913.40 | 0.10090206178697514 |
| 2021-09-13 | 2021-09-13 (D0) | BRL 762.66 | BRL 1,906.65 | 0.10230155977571715 |

Fixed expected rates were evaluated independently with Python standard-library Decimal at precision 60 as `(Decimal(1000) / Decimal(sell_pu)) ** (Decimal(252) / Decimal(701)) - 1`. Tests do not generate expected values using production math or calendar functions. Tolerances are BRL 1e-9 and annual fractional rate 1e-12.

## Validation and limitations

Checks passed on 2026-09-30:

- `go fmt ./...`
- `go vet ./...`
- `go test ./...`
- `go test -race ./...` (followed by `go test -race ./internal/web` after the final currency-formatting adjustment)
- `CGO_ENABLED=0 go build -trimpath -o /private/tmp/tlab-m4 ./cmd/tesouro-lab`
- `/private/tmp/tlab-m4 analyze ipca:2015-05-15 --source demo --basis mark_to_market --date 2012-02-17 --yield 3.51`
- `git diff --check`

The compiled demo scenario reproduced PU BRL 1,897.2254119803258 with DU 813. Currency presentation tests include negative gross changes, preserving the sign independently of thousands grouping.

Deterministic tests cover acquisition completeness/derivation/conflicts, non-finite and range boundaries, official BasePU selection, partial totals, mature/older/missing/future quote states, local-date maturity boundaries, Selic limits, both settlement conventions, validation mismatches, stored-quantity scenarios, and persistent/demo isolation. Storage tests upgrade an actual M3 schema, preserve quotes, reopen edited positions, reject foreign-key failures atomically, and retain positions during quote upserts. HTTP tests exercise forms without JavaScript, rendering completion, escaped notes, edit recovery, scenario results, removal confirmation, body limits, and cross-origin/token rejection.

Interactive browser layout/keyboard testing, cross-platform execution, remote CI, and release packaging remain unverified M5 work. No historical portfolio performance series, tax engine, cash account, or external portfolio synchronization is implemented. Broader usability work and Brazilian Portuguese support remain deferred until the planned V1 is complete.
