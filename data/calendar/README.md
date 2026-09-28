# Verified financial-market calendar

`anbima-2002-2032-v1.json` contains 381 dates extracted from ANBIMA's annual national financial-market holiday tables, retrieved on 2026-09-10. Coverage is inclusive from 2002-01-01 through 2032-12-31. Each annual source URL, raw HTML SHA-256, and extracted holiday count is recorded in the fixture. Only factual dates and provenance are redistributed; attribution remains with ANBIMA, and the project's MIT license does not relicense source material.

The range covers the years specified for the source dataset through the latest no-coupon Prefixado maturity observed in the downloaded CSV (2032-01-01). It does not imply quote availability for every date, or support for other instruments.

## Calendar rule

Weekends and the listed holidays are excluded. The [ANBIMA notes](https://www.anbima.com.br/feriados/) distinguish financial-market holidays from municipal holidays, elections, and year-end bank public-service closures. No extra closures are inserted. For example, 2026-12-31 is a business day in this fixture. The [2023 table](https://www.anbima.com.br/feriados/fer_nacionais/2023.asp) does not list November 20; [2024](https://www.anbima.com.br/feriados/fer_nacionais/2024.asp) and subsequent covered years do. This transition is taken directly from the tables, not inferred from the current year.

For scenarios, count settlement inclusive and maturity exclusive. Both endpoints must lie within verified coverage, even when the maturity endpoint is excluded from the count. D+1 advances to the next financial-market business day, including after weekends and Carnival; it does not mean the next calendar day. D0 retains the official quote date. A term with no remaining business days is unavailable.

Dataset-age warnings also use this unchanged fixture. `Previous` selects the business day strictly before the local date, failing if no predecessor is within verified coverage. Lag counts business dates after the stored dataset date through that expected date; this does not change scenario settlement/counting rules. See the [warning policy](../../docs/OFFICIAL_DATASOURCE.md#dataset-age-warning-policy).

## Extraction and independent verification

The annual HTML is ISO-8859-1. Extract only date cells matching `>\s*(\d{1,2}/\d{1,2}/\d{2})\s*</td>`, parse each as `DD/MM/YY`, and require the year to match the page heading and URL. The 2002–2023 pages each contain 12 unique ordered dates; 2024–2032 each contain 13. The fixture normalizes those factual dates to ISO format; it does not use a formula to invent future holidays.

All 2012–2015 business-day classifications match the original demo calendar exactly. The original file and version `anbima-2012-2015-v1` remain unchanged for demo reproducibility. The expanded version is embedded for the verified quote-context workflow; synchronized purchase/base CLI/browser analysis uses this version.

Tests cover all annual business-day totals, explicit settlement boundaries, bounds, and the official demo count of 755 days. Independent reference counts use complete weeks plus remainder weekdays, minus weekday holidays; quote fixture counts are also cross-checked by separate date enumeration. See the [quote fixtures and reference script](../../internal/pricing/testdata/README.md).

Future source revisions require review and a new version; no calendar is fetched at application runtime. Dates after 2032-12-31 or before 2002-01-01 return `calendar_out_of_range` when this calendar is selected. The demo retains its narrower coverage.
