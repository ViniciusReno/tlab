# M2 settlement decision — proposed, not approved

Prepared on 2026-09-28. The locked V1 specification and runtime behavior remain unchanged. M2 cannot be marked complete until the maintainer accepts a resolution of its early-redemption requirement.

## Evidence

The [CKAN metadata](https://www.tesourotransparente.gov.br/ckan/dataset/df56aa42-484a-4a59-8184-7676580c81e3/resource/1a8eb2e3-4902-4a38-a1eb-6410f23d90de/download/taxa.pdf) describes SellPU as D+1, BasePU as D0, and publication on the first business day following market close. In contrast, the [B3 announcement of 2021-09-13](https://www.b3.com.br/pt_br/noticias/liquidacao-do-tesouro-direto-passa-a-ser-em-d-0.htm) states that redemption requests before 13:00 under normal market conditions changed to D0 on that date. Requests after that time remain D+1. This announcement concerns operational settlement; mapping the morning CSV field is an inference requiring quote evidence.

Ten raw official Prefixado records, five maturities on each of 2021-09-10 and 2021-09-13, are preserved in `internal/pricing/testdata/redemption-transition.csv`. Adjacent provenance records retrieval date, source/extract SHA-256 and original line numbers. Attribution and ODbL terms follow the existing quote fixtures. No other instruments or later milestones are introduced.

`redemption_reference.py` independently calculates both D0 and D+1 with 60-digit Decimal arithmetic. `redemption-transition.json` records fixed DU, truncated theoretical PU, signed difference from SellPU, and hypothetical shocks. Go tests verify the evidence against the unchanged ANBIMA fixture, without treating mismatches as accepted pricing baselines. Hypothetical outputs for mismatching cases are diagnostic calculations, not validated runtime scenarios.

For maturities 2022-01-01, 2023-01-01, and 2024-07-01, all three 2021-09-10 records match D+1 and all three 2021-09-13 records match D0, with zero truncated difference. The opposite convention misses by BRL 0.28–0.31.

Longer maturities expose a separate historical-calendar issue:

| Quote date | Maturity | Proposed convention | Verified-calendar DU | Signed PU difference |
|---|---|---|---:|---:|
| 2021-09-10 | 2025-01-01 | D+1 | 830 | BRL +0.28 |
| 2021-09-13 | 2025-01-01 | D0 | 830 | BRL +0.28 |
| 2021-09-10 | 2026-01-01 | D+1 | 1082 | BRL +0.51 |
| 2021-09-13 | 2026-01-01 | D0 | 1082 | BRL +0.52 |

Investigation found that adding back the November 20 business days in 2024 and 2025 makes these four differences zero. This is consistent with the later introduction of the national holiday by [Law 14,759 of 2023](https://www.planalto.gov.br/ccivil_03/_ato2023-2026/2023/lei/l14759.htm). It is an explanatory hypothesis, not authorization to remove official holidays or reconstruct a historical calendar from prices. No such adjustment was made to production or the accepted fixtures.

## Proposed amendment for maintainer approval

1. Amend V1 sections 8.8 in AGENTS.md and 14.3/17.4 in V1_SPEC.md, plus corresponding documentation: morning-quote early-redemption scenarios use SellPU/SellYield with D+1 before 2021-09-13 and D0 on/after that date. Purchase remains BuyPU/BuyYield/D+1; mark-to-market remains BasePU/SellYield/D0. Preserve these distinct official fields even where their numbers coincide.
2. For synchronized Prefixado early-redemption scenarios, require standalone validation of the selected SellPU/yield and resolved DU before enabling the anchored scenario. Apply the existing truncated-PU tolerance of BRL 0.01 without widening it. Return `calculation_not_validated` for mismatches, including historical-calendar discrepancies. Do not select whichever convention happens to fit the row.
3. Label the D0 model as a morning-quote, normal-market scenario. It does not model after-13:00 requests, suspended trading, actual execution, taxes, or fees. Preserve gross labeling and exact dates.
4. Keep the verified calendar unchanged. Historical calendar reconstruction is not needed to close M2 if explicitly unavailable mismatching rows are accepted. The ten new records, the existing recent records, and both nonzero shock directions must be covered by deterministic tests before enabling delivery.
5. Update result versioning, CLI/UI context labels, generated commands, gate tests, README and the checkpoint together. Mark M2 complete only after this amendment is accepted, implemented, and all checks pass.

This proposal is intentionally limited to Prefixado M2. IPCA+ will need its own validated evidence in M3. M4 hold-versus-early-exit must use the approved settlement contract and retain unavailable states.

## Alternative

Keep all synchronized early-redemption scenarios blocked and explicitly defer their delivery as a milestone exception. This also needs maintainer acceptance to call M2 complete; it must not be described as full early-redemption support.

## Review validation

The evidence test and repository checks passed locally: `go fmt ./...`, `go vet ./...`, `go test ./...`, `go test -race ./...`, and `git diff --check`. Formatting was rerun with cache access after a sandbox restriction. No production pricing, settlement policy, or calendar fixture was changed. No release, remote CI run, or cross-platform execution was performed.
