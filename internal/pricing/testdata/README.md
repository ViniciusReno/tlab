# Independent Prefixado quote-context validation

These are official historical validation fixtures, not current market values and not additional demo quotes. `quote-contexts.csv` preserves four complete raw lines and the original header from the [official CSV](https://www.tesourotransparente.gov.br/ckan/dataset/df56aa42-484a-4a59-8184-7676580c81e3/resource/796d2059-14e9-44e3-80c9-2d9e30b405c1/download/precotaxatesourodireto.csv), retrieved on 2026-09-10. `provenance.json` records the URL, source byte count/SHA-256, original 1-based line numbers, and extract SHA-256. Raw names, dates, decimal commas, and line endings are preserved.

Attribution: Tesouro Nacional / Tesouro Transparente. These source extracts retain the dataset's [ODbL terms](https://opendatacommons.org/licenses/odbl/1-0/); the project MIT license does not relicense them. Derived hypothetical outputs are identified separately in `scenarios.json`. Calendar attribution remains with ANBIMA.

## Official inputs and verified terms

`scenarios.json` fixes quote context, settlement, calendar version, independent DU, annual DU contributions, original normalized PU/yield, standalone theoretical PU, and expected nonzero-shock outputs. Yields are annual decimal fractions and PU is BRL per bond. Purchase uses BuyPU/BuyYield/D+1, mark-to-market uses BasePU/SellYield/D0, and morning early redemption uses SellPU/SellYield/D+1 before 2021-09-13 and D0 thereafter under the approved settlement amendment.

| Quote date | Maturity | D+1 settlement | D+1 DU | D0 DU | Standalone validation |
|---|---|---|---:|---:|---|
| 2012-01-03 | 2015-01-01 | 2012-01-04 | 755 | 756 | All three contexts |
| 2016-02-05 | 2018-01-01 | 2016-02-10 | 475 | 476 | All three contexts; Carnival boundary |
| 2024-11-19 | 2027-01-01 | 2024-11-21 | 529 | 530 | All three contexts; November 20 holiday |
| 2026-09-04 | 2032-01-01 | 2026-09-08 | 1331 | 1332 | All three contexts; weekend and September 7 holiday |

The raw 2012 CKAN quote is distinct from the original methodology demo. All expectations here use the CKAN values, without replacing the demo baseline.

## Independent calculation

`reference.py` uses only Python's standard library and 60-digit `Decimal` arithmetic. It does not import or execute Go production functions. Python is an optional developer verification tool, not an application runtime or Go test dependency.

The calendar source is [anbima-2002-2032-v1](../../../data/calendar/README.md). For each half-open interval `[settlement, maturity)`, count full weeks and remainder weekdays, then subtract listed holidays on weekdays. The script independently checks this count by enumerating each date. Per-year counts are saved for review; for 2026 purchase, `79 + 251 + 248 + 249 + 252 + 252 = 1331`.

Let `t = DU / 252`, baseline PU be `P`, and annual baseline yield be `y`. Standalone validation computes `1000 / (1 + y)^t`. Independently construct the fixed anchor `A = P × (1 + y)^t`, then discount `A / (1 + hypothetical_yield)^t` at the baseline yield and at ±0.01 (±100 basis points). Fractional variation is `scenario_PU / P - 1`. These are hypothetical same-date changes, not forecasts or official future quotes.

For example, purchase on 2016-02-05 uses `P = 766.64`, `y = 0.1514`, and `t = 475 / 252`. The unrounded standalone baseline is approximately `766.643441784018699342736540228035`. At hypothetical yields 0.1414 and 0.1614, anchored PUs are approximately `779.3494392188653` and `754.2450586980491`; fractional changes are `0.016578106045686835` and `-0.01616787710261779`.

To print independent expectations without modifying the committed file:

```sh
python3 internal/pricing/testdata/reference.py
```

Go tests load the committed values directly. Standalone validation truncates to two decimals and requires absolute difference <= BRL 0.01. Scenario PU and zero-shock tolerances are absolute BRL 1e-9; yield/fractional-change tolerances are 1e-12. No intermediate display rounding is used. Nonzero shocks must detect both +1 and -1 DU errors in PU and variation.

## Resolved settlement discrepancy and remaining historical limits

The previous unconditional D+1 rule missed the recent 2024 and 2026 SellPUs by BRL 0.38 and BRL 0.26 respectively. The approved morning-quote D0 rule now validates both records, with independent zero and ±100 bps shock expectations in `scenarios.json`. The original official CSV is unchanged. The [approved decision](../../../docs/M2_SETTLEMENT_DECISION.md) explains why the metadata's unconditional D+1 wording is no longer the runtime contract.

`redemption-transition.csv` preserves ten additional raw official records from 2021-09-10 and 2021-09-13, with source/extract hashes and original line numbers in `redemption-transition-provenance.json`. They retain Tesouro Nacional attribution and ODbL terms. `redemption_reference.py` prints independent 60-digit expectations stored in `redemption-transition.json`; normal Go tests consume fixed values without network access or Python.

The transition evidence compares both conventions and deliberately retains failures. Three maturities validate D+1 before the transition and D0 after it. Four longer-maturity records fail with the unchanged verified calendar and remain `calculation_not_validated` in the application. Diagnostic shocks for those mismatching rows are mathematical evidence only, not enabled scenarios. The application selects settlement by date and then validates the official SellPU; it never picks whichever convention fits a price.

Historical knowledge of future holidays can differ from the current fixture. No holiday is removed and no historical calendar is inferred to force a match. The six matching transition records and four original quote-context fixtures cover both directions of hypothetical yield shocks. PU tolerances remain BRL 1e-9 for scenario comparisons and BRL 0.01 for truncated standalone validation; fractional variation uses 1e-12.
