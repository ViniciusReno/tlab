# Official datasource contract — M2.2

Verified on 2026-09-08 against the [official CKAN API](https://www.tesourotransparente.gov.br/ckan/api/3/action/package_show?id=taxas-dos-titulos-ofertados-pelo-tesouro-direto), CSV, and [metadata PDF](https://www.tesourotransparente.gov.br/ckan/dataset/df56aa42-484a-4a59-8184-7676580c81e3/resource/1a8eb2e3-4902-4a38-a1eb-6410f23d90de/download/taxa.pdf). Raw fixtures, hashes, extraction details, and attribution are in [testdata](../internal/datasource/tesouro/testdata/README.md).

## Discovery for the next synchronization step

Use the fixed `package_show` endpoint above. The snapshot reports `success: true`, an active public package with ID `df56aa42-484a-4a59-8184-7676580c81e3`, and two resources: metadata PDF and one active CSV. The CSV ID is `796d2059-14e9-44e3-80c9-2d9e30b405c1`; its URL must be discovered from the response, not treated as permanent.

The snapshot encodes resource `size` as a string, `mimetype` as null, and `hash` as an empty string. Neither CKAN size nor modification timestamps establish download completeness or a market date. The API reports CSV size 14,474,206 bytes, matching the retrieved body. The latest actual quote date in that body is 2026-09-04.

M2.3 must validate API success/package identity/state and select exactly one active CSV resource; missing or ambiguous selection must fail explicitly. Keep URL fetching limited to official HTTPS hosts, including redirect validation. Bound metadata and CSV downloads, use timeouts, validate HTTP content type/body, and never trust the declared size as the sole limit. The inspected resource returned `text/csv`; null CKAN `mimetype` must not prevent checking the HTTP response. No downloader or resource resolver is implemented by M2.2.

## CSV normalization

The downloaded CSV is UTF-8 without BOM, semicolon-delimited, with LF endings. The parser also accepts UTF-8 BOM and CRLF. Dates use exact `DD/MM/YYYY`; numeric values use comma decimals, optionally with grouped thousands separated by periods. Yields are percentages and are divided by 100 once; PUs remain BRL per bond. No machine-locale parsing is used.

| Required source column | Normalized meaning |
|---|---|
| `Tipo Titulo` | Official instrument name |
| `Data Vencimento` | Maturity date |
| `Data Base` | Official quote date |
| `Taxa Compra Manha` | Purchase yield |
| `Taxa Venda Manha` | Early-redemption / base-valuation yield |
| `PU Compra Manha` | Purchase PU, D+1 |
| `PU Venda Manha` | Early-redemption PU, D+1 |
| `PU Base Manha` | Mark-to-market PU, D0 |

Header order may change and extra columns are ignored; all eight columns must exist with unique names. Empty numeric cells remain unavailable, without replacing missing context pairs. Duplicate headers, invalid UTF-8, malformed dates/numbers/row lengths, and nonpositive Prefixado PUs fail with no partial result. Historical records are accepted based on their quote and maturity dates; today's date is irrelevant to parsing.

`Parse` retains the strict 1 MiB demo limit and rejects other instrument types. `ParseDataset` allows up to 32 MiB (including any BOM), enough for the observed 13.8 MiB CSV with bounded growth headroom. It reads the bounded body into memory, returns normalized no-coupon Prefixado quotes, and reports excluded row counts keyed by exact source name. An unsupported-only dataset returns that report with zero accepted quotes; an empty dataset returns `missing_quote`. The future sync must report those counts and must not describe zero accepted rows as a successful Prefixado import.

Unsupported rows still require valid dates, numeric syntax, and finite values, but do not undergo Prefixado price/yield validation or enter storage. Unknown names are explicitly counted, never classified by fuzzy matching. This keeps coupon instruments unsupported and leaves IPCA+/Selic implementation for M3.

## Limits

The fixtures validate ingestion only. They do not extend verified calendar coverage, validate new pricing scenarios, enable synchronized analysis, or change embedded demo values. Calendar/context validation remains a separate M2 step. Duplicate quote/upsert handling, provenance timestamps, failed-sync reporting, and transactional writes belong to M2.3.
