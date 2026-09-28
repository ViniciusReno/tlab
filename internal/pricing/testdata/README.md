# Independent Prefixado quote-context validation

These are official historical validation fixtures, not current market values and not additional demo quotes. `quote-contexts.csv` preserves four complete raw lines and the original header from the [official CSV](https://www.tesourotransparente.gov.br/ckan/dataset/df56aa42-484a-4a59-8184-7676580c81e3/resource/796d2059-14e9-44e3-80c9-2d9e30b405c1/download/precotaxatesourodireto.csv), retrieved on 2026-09-10. `provenance.json` records the URL, source byte count/SHA-256, original 1-based line numbers, and extract SHA-256. Raw names, dates, decimal commas, and line endings are preserved.

Attribution: Tesouro Nacional / Tesouro Transparente. These source extracts retain the dataset's [ODbL terms](https://opendatacommons.org/licenses/odbl/1-0/); the project MIT license does not relicense them. Derived hypothetical outputs are identified separately in `scenarios.json`. Calendar attribution remains with ANBIMA.

## Official inputs and verified terms

`scenarios.json` fixes quote context, settlement, calendar version, independent DU, annual DU contributions, original normalized PU/yield, standalone theoretical PU, and expected nonzero-shock outputs. Yields are annual decimal fractions and PU is BRL per bond. All three contexts use the locked V1 contract: purchase uses BuyPU/BuyYield/D+1, mark-to-market uses BasePU/SellYield/D0, and early redemption uses SellPU/SellYield/D+1.

| Quote date | Maturity | D+1 settlement | D+1 DU | D0 DU | Standalone validation |
|---|---|---|---:|---:|---|
| 2012-01-03 | 2015-01-01 | 2012-01-04 | 755 | 756 | All three contexts |
| 2016-02-05 | 2018-01-01 | 2016-02-10 | 475 | 476 | All three contexts; Carnival boundary |
| 2024-11-19 | 2027-01-01 | 2024-11-21 | 529 | 530 | Purchase and base; November 20 holiday |
| 2026-09-04 | 2032-01-01 | 2026-09-08 | 1331 | 1332 | Purchase and base; weekend and September 7 holiday |

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

## Known source/specification discrepancy

The two recent early-redemption cases are explicitly marked `calculation_not_validated` and contain no accepted shock expectations. They are regression evidence of a failed standalone validation, not fixtures made to pass by changing the calendar or tolerance:

| Quote / maturity | Official SellPU | D+1 theoretical PU, truncated | Difference |
|---|---:|---:|---:|
| 2024-11-19 / 2027-01-01 | BRL 765.76 | BRL 766.14 | BRL 0.38 |
| 2026-09-04 / 2032-01-01 | BRL 490.42 | BRL 490.68 | BRL 0.26 |

Those source SellPUs equal BasePU and match the standalone D0 calculation after truncation. The [metadata PDF](https://www.tesourotransparente.gov.br/ckan/dataset/df56aa42-484a-4a59-8184-7676580c81e3/resource/1a8eb2e3-4902-4a38-a1eb-6410f23d90de/download/taxa.pdf) still describes SellPU as D+1. An [official announcement dated 2021-09-13](https://www.gov.br/tesouronacional/pt-br/noticias/tesouro-direto-passa-a-ter-resgate-no-mesmo-dia/) describes D0 redemption for requests before 13:00 under normal market conditions. This is evidence worth investigating; that announcement alone does not establish a universal historical mapping for the CSV column.

The locked V1 D+1 rule is preserved. The production basis resolver follows that rule; it is not a runtime standalone-validation gate and does not reinterpret recent rows as D0. M2.5b enables synchronized purchase/base scenarios. The application service blocks all synchronized early-redemption scenarios with `calculation_not_validated` after resolving the required pair and term, including older rows, until an explicit historical settlement contract is validated. It does not infer a safe transition date. Before enabling recent early-redemption scenarios, reconcile the source contract with the maintainer and add official transition fixtures; do not silently change settlement, widen tolerances, or infer a historical switch from PU equality. Purchase/base validation and historical early-redemption validation remain useful independently of this unresolved case.

## Transition evidence added during the M2 closure review

`redemption-transition.csv` preserves ten additional raw official records from 2021-09-10 and 2021-09-13, with source/extract hashes and original line numbers in `redemption-transition-provenance.json`. They retain Tesouro Nacional attribution and the ODbL terms above. `redemption_reference.py` prints independent 60-digit expectations stored in `redemption-transition.json`; normal Go tests use those fixed values without network access or Python.

These fixtures compare both settlement conventions and deliberately preserve failures. Diagnostic shocks for a mismatching baseline do not certify its settlement contract. Three maturities support a D+1-to-D0 transition; two longer maturities also expose a historical-calendar discrepancy. See the [proposed decision](../../../docs/M2_SETTLEMENT_DECISION.md). Production early-redemption scenarios remain blocked pending approval and implementation; these fixtures alone do not change runtime policy.
