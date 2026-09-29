# Independent IPCA+ and Selic fixtures

`m3-quotes.csv` preserves four raw official CSV records downloaded on 2026-09-27. `m3-provenance.json` records the official source URL, full source size/SHA-256, original line numbers, extract hash, and the separate demo extract hash. `data/demo/ipca.csv` preserves the historical IPCA+ row and header unchanged. Attribution: Tesouro Nacional / Tesouro Transparente, under the dataset's [ODbL terms](https://opendatacommons.org/licenses/odbl/1-0/). The project MIT license does not relicense these source extracts.

| Instrument / maturity | Official quote date | Buy PU / yield | Sell PU / yield | Base PU |
|---|---|---|---|---|
| IPCA+ 2015-05-15 | 2012-02-17 | BRL 1,843.43 / 4.47% real | BRL 1,841.15 / 4.51% real | BRL 1,839.28 |
| IPCA+ 2029-05-15 | 2026-09-04 | BRL 3,876.96 / 7.80% real | BRL 3,865.75 / 7.92% real | BRL 3,865.75 |
| IPCA+ 2050-08-15 | 2026-09-04 | BRL 889.56 / 7.28% real | BRL 866.28 / 7.40% real | BRL 866.28 |
| Selic 2029-03-01 | 2026-09-04 | BRL 19,810.47 / 0.03% spread | BRL 19,795.28 / 0.04% spread | BRL 19,795.28 |

These are historical records, not current live values. Selic has no hypothetical scenario expectations; it is used only to verify quote ingestion/display and the unsupported analytical boundary.

## Fixed-indexation reference

The [official IPCA+ methodology, pages 1–4](https://www.tesourodireto.com.br/documents/d/guest/tesouro_ipca_juros_semestrais), reviewed on 2026-09-29, relates the no-coupon price to an inflation-adjusted base and a real-yield discount factor. V1 section 17 deliberately holds the official quote's indexation base constant for same-date scenarios rather than projecting future inflation.

For official PU `P`, real annual yield `y`, and verified term `t = DU/252`, the independent reference constructs the implied anchor `A = P × (1+y)^t` once. It then computes `A/(1+scenario_yield)^t` for each shock and fractional variation `scenario_PU/P - 1`. `A` is derived from the official pair; it is not an independently observed or reconstructed official VNA. Each quote context has its own anchor, preserved across its shocks. Production uses the algebraically equivalent anchored ratio formula.

`ipca_reference.py` uses 60-digit Decimal arithmetic and never imports Go code. Its fixed results are stored in `ipca-scenarios.json`. No reference program runs during Go tests. To inspect them without changing fixtures:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 internal/pricing/testdata/ipca_reference.py
```

| Quote date | Context | Settlement | DU |
|---|---|---|---:|
| 2012-02-17 | purchase / early_exit | 2012-02-22 | 812 |
| 2012-02-17 | mark_to_market | 2012-02-17 | 813 |
| 2026-09-04 (2029 maturity) | purchase | 2026-09-08 | 669 |
| 2026-09-04 (2029 maturity) | mark_to_market / early_exit | 2026-09-04 | 670 |
| 2026-09-04 (2050 maturity) | purchase | 2026-09-08 | 5994 |
| 2026-09-04 (2050 maturity) | mark_to_market / early_exit | 2026-09-04 | 5995 |

Count settlement inclusive to maturity exclusive using the versioned 2002–2050 [ANBIMA fixture](../../../data/calendar/README.md). Carnival and the September 7 holiday explain the D+1 boundaries. Full-week/remainder counting minus weekday holidays is independently cross-checked by date enumeration. Demo calendar classifications match the expanded fixture over 2012–2015.

For example, the 2012 purchase pair at 4.47% and BRL 1,843.43 has DU 812. At a hypothetical 3.47% real yield, the price is BRL 1,901.4562720782521 and fractional variation 0.0314773395671396. The same reference records zero and +100 bps shocks, plus both D0/D+1 context sets.

Scenario PU tolerance is absolute BRL 1e-9; decimal yields and fractional variation use 1e-12. Nonzero expectations detect a one-business-day term error. No intermediate rounding, future inflation assumption, Prefixado BRL 1,000 face-value formula, or artificial Selic repricing is used.
