# Official datasource contract — M2.2/M2.3

Verified on 2026-09-08 against the [official CKAN API](https://www.tesourotransparente.gov.br/ckan/api/3/action/package_show?id=taxas-dos-titulos-ofertados-pelo-tesouro-direto), CSV, and [metadata PDF](https://www.tesourotransparente.gov.br/ckan/dataset/df56aa42-484a-4a59-8184-7676580c81e3/resource/1a8eb2e3-4902-4a38-a1eb-6410f23d90de/download/taxa.pdf). Raw fixtures, hashes, extraction details, and attribution are in [testdata](../internal/datasource/tesouro/testdata/README.md).

## Discovery and explicit synchronization

Use the fixed `package_show` endpoint above. The snapshot reports `success: true`, an active public package with ID `df56aa42-484a-4a59-8184-7676580c81e3`, and two resources: metadata PDF and one active CSV. The CSV ID is `796d2059-14e9-44e3-80c9-2d9e30b405c1`; its URL must be discovered from the response, not treated as permanent.

The snapshot encodes resource `size` as a string, `mimetype` as null, and `hash` as an empty string. Neither CKAN size nor modification timestamps establish download completeness or a market date. The API reports CSV size 14,474,206 bytes, matching the retrieved body. The latest actual quote date in that body is 2026-09-04.

M2.3 validates API success, active public package identity, and exactly one active CSV resource belonging to that package. Missing or ambiguous selection fails explicitly. Fetching and redirects are limited to `https://www.tesourotransparente.gov.br` without credentials, fragments, or alternate ports. Metadata is limited to 1 MiB and CSV to 32 MiB, including streaming bodies with no declared size. Each request has a 60-second timeout and at most four followed redirects. HTTP status must be 200; content types must be `application/json` and `text/csv` respectively (parameters such as charset are accepted). The inspected resource returned `text/csv`; null CKAN `mimetype` must not prevent checking the HTTP response. The M2.3 client discovers the resource for each explicit synchronization; users cannot supply arbitrary URLs.

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
| `PU Venda Manha` | Morning early-redemption PU; metadata says D+1, runtime uses the approved date-dependent contract below |
| `PU Base Manha` | Mark-to-market PU, D0 |

Header order may change and extra columns are ignored; all eight columns must exist with unique names. Empty numeric cells remain unavailable, without replacing missing context pairs. Duplicate headers, duplicate supported bond/date rows, invalid UTF-8, malformed dates/numbers/row lengths, and nonpositive Prefixado PUs fail with no partial result. Duplicate rows are rejected rather than selecting an order-dependent winner; corrections between successful imports update the same natural key. Historical records are accepted based on their quote and maturity dates; today's date is irrelevant to parsing.

`Parse` retains the strict 1 MiB demo limit and rejects other instrument types. `ParseDataset` allows up to 32 MiB (including any BOM), enough for the observed 13.8 MiB CSV with bounded growth headroom. It reads the bounded body into memory, returns normalized no-coupon Prefixado quotes, and reports excluded row counts keyed by exact source name. An unsupported-only dataset returns that report with zero accepted quotes; an empty dataset returns `missing_quote`. Sync reports those counts and treats zero accepted rows as a failure, preserving the previous dataset.

Unsupported rows still require valid dates, numeric syntax, and finite values, but do not undergo Prefixado price/yield validation or enter storage. Unknown names are explicitly counted, never classified by fuzzy matching. This keeps coupon instruments unsupported and leaves IPCA+/Selic implementation for M3.

## Limits

The fixtures validate ingestion only. They do not extend verified calendar coverage, validate new pricing scenarios, enable synchronized analysis, or change embedded demo values. M2.4 calendar/context validation is recorded in the [independent fixtures](../internal/pricing/testdata/README.md), including the evidence for the approved SellPU settlement amendment and remaining historical-calendar limitations. Ingestion retains all raw quote contexts without claiming every row has validated standalone pricing. M2.3 stores the CKAN endpoint, resource ID, final download URL, import timestamp, excluded counts, and maximum validated source quote date in each successful sync run. The maximum date includes excluded instruments. Quotes retain their final source URL and import timestamp. Failed runs preserve prior quotes and successful-run metadata. The local import timestamp does not establish market freshness; M2.5 market/history views distinguish missing, older, latest-in-dataset, and unknown-freshness states. M2.5c adds the conservative dataset-age warning documented below.


## Dataset-age warning policy

The [official resource description](https://www.tesourotransparente.gov.br/ckan/dataset/taxas-dos-titulos-ofertados-pelo-tesouro-direto/resource/796d2059-14e9-44e3-80c9-2d9e30b405c1), checked on 2026-09-27, describes a daily list. It does not establish an intraday publication cutoff. The following is an explicit application assumption, not an official publication SLA:

- Expected base date: the verified financial-market business day strictly before the user's local calendar date. Do not advance this expectation during the day.
- Lag: number of verified business dates strictly after `dataset_max_quote_date` and through the expected base date. A quote on or after the expected date has zero lag, unless it lies in the future relative to the user's local date.
- Warn at lag >= 2; tolerate a one-business-day discrepancy. A warning describes the stored local dataset, not a confirmed outage at the official source.
- Always retain the actual dataset date and successful-import timestamp. Import time is not an input to the age calculation. Failed syncs cannot replace successful-import metadata.
- Missing/invalid metadata, future quote dates, or insufficient verified calendar coverage yield an unavailable assessment. No weekday-only fallback or guessed holiday is allowed.

For example, on local 2026-09-10 the expected base date is 2026-09-09. A stored 2026-09-08 dataset is one business day behind and does not warn. A 2026-09-04 dataset is two business days behind (September 8 and 9) and does warn; the weekend and September 7 holiday add no lag. This does not alter per-bond `latest`, `older_quote`, `no_quote`, or `matured` states. A bond can be latest within a dataset that itself appears stale.

The embedded `anbima-2002-2032-v1` calendar is unchanged. The warning works offline, does not trigger sync, and does not certify that a quote is live when no warning is shown.


## Approved morning-redemption settlement contract

The maintainer approved the [M2 settlement amendment](M2_SETTLEMENT_DECISION.md) on 2026-09-28. The metadata's unconditional SellPU D+1 wording conflicts with the 2021 operational change and the transition records. Runtime Prefixado morning-redemption scenarios now use D+1 before 2021-09-13 and D0 thereafter, and require per-record standalone PU validation (truncated to cents; BRL 0.01 tolerance). The original official SellPU remains the scenario anchor. Historical-calendar mismatches return `calculation_not_validated`; raw history stays visible. This does not model actual execution, after-13:00 requests, or suspended trading. Purchase and BasePU contexts remain unchanged.

The metadata PDF also explicitly states publication on the first business day after secondary-market close. This supports the previous-business-day reference in the dataset-age policy; no intraday publication time is assumed.
