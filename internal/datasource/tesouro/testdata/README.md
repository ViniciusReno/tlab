# Official datasource fixtures

Retrieved on 2026-09-08 from Tesouro Nacional / Tesouro Transparente. These are historical parser fixtures, not live quotes or independently validated pricing scenarios. They are not embedded in the demo binary.

## Sources and license

- [Dataset](https://www.tesourotransparente.gov.br/ckan/dataset/taxas-dos-titulos-ofertados-pelo-tesouro-direto)
- [CKAN package_show response](https://www.tesourotransparente.gov.br/ckan/api/3/action/package_show?id=taxas-dos-titulos-ofertados-pelo-tesouro-direto)
- [CSV resource](https://www.tesourotransparente.gov.br/ckan/dataset/df56aa42-484a-4a59-8184-7676580c81e3/resource/796d2059-14e9-44e3-80c9-2d9e30b405c1/download/precotaxatesourodireto.csv)
- [Official metadata PDF](https://www.tesourotransparente.gov.br/ckan/dataset/df56aa42-484a-4a59-8184-7676580c81e3/resource/1a8eb2e3-4902-4a38-a1eb-6410f23d90de/download/taxa.pdf)

The dataset declares `odc-odbl`, the [Open Data Commons Open Database License](https://opendatacommons.org/licenses/odbl/1-0/). These extracts retain the source attribution and applicable ODbL terms; the project MIT license does not relicense them. The metadata PDF is referenced, not redistributed.

## Exact extraction

`package-show.json` is the complete, unmodified API response (5,355 bytes). Its SHA-256 is `49ac06590e6f19b4968d8ea529e029446ce9522941356b06cfe72fbf534c9be6`.

`official-quotes.csv` preserves the original header and raw lines, including names, dates, numbers, encoding, and LF endings. Selected 1-based source line numbers are `2, 6, 11, 20, 26, 27, 33, 41, 118499, 49020`, in that fixture order. The first eight rows represent the first occurrence of each instrument name; line 118499 provides a historical Prefixado with distinct PUs; line 49020 exercises a zero PU in an unsupported coupon instrument. The fixture is 936 bytes with SHA-256 `c406a843deec17e734432678f51d1b8f57d33c7069cd46b9bfefef3bae0c70a7`.

The downloaded CSV was 14,474,206 bytes, SHA-256 `1d8196214a695fd259f4ab600aee6bb709e4f0809d2651134ecc5e080d754fc5`. It contained 175,868 records (28,080 no-coupon Prefixado), with quote dates from 2004-12-31 through 2026-09-04. The full download is not committed. CKAN metadata modification time is 2026-09-08T10:20:08.730501; this is not a quote date.

Recreate the CSV extract from that exact downloaded snapshot with Python standard library:

```python
from pathlib import Path
lines = Path("precotaxatesourodireto.csv").read_bytes().splitlines(keepends=True)
selected = [2, 6, 11, 20, 26, 27, 33, 41, 118499, 49020]
Path("official-quotes.csv").write_bytes(lines[0] + b"".join(lines[n - 1] for n in selected))
```

Source revisions may change hashes and line numbers. Check the full-file hash before reproducing this selection.

## Interpretation and limits

The CKAN Prefixado 2015 row for 2012-01-03 has purchase yield 10.83% and purchase PU BRL 734.86. The separate official methodology example in `data/demo` uses 10.88% and BRL 733.86. Neither source is substituted for the other; this fixture does not revise the demo or establish a pricing/calendar validation baseline.

M2 imports only no-coupon Prefixado. Other names are counted as unsupported in this milestone, including IPCA+ and Selic pending M3. Unsupported rows have their dates and numeric syntax checked but never become domain prices. A zero PU in an excluded coupon record is not inferred to mean missing and is not accepted as a valid Prefixado price.

No empty numeric fields were observed in the downloaded Prefixado records. Nullable-field, reordered-header, BOM, extra-column, malformed-input, and size-limit tests use explicitly synthetic variations. Empty numeric fields remain nil; missing columns fail. Parser comparisons use absolute tolerances of 1e-12 for normalized yields and BRL 1e-9 for PUs. These are parsing tests, not financial repricing validation.
