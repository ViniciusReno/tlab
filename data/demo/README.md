# M1 historical demo fixtures

Fixture version: `prefixado-2012-v1`. Retrieved: 2026-09-05.

## Official baseline

`prefixado.csv` is a **transcription of an official worked example**, placed in the Tesouro dataset's column layout to exercise Brazilian CSV parsing. It is not a downloaded CKAN row or a current market quote.

Source: [Tesouro Direto, Prefixado methodology](https://www.tesourodireto.com.br/documents/d/guest/tesouro_prefixado), PDF pages 1–3.

| Field | Value |
|---|---|
| Instrument | Tesouro Prefixado 2015 (LTN), no coupon |
| Quote/purchase date | 2012-01-03 |
| Settlement | 2012-01-04 |
| Maturity | 2015-01-01 |
| Official purchase PU | BRL 733.86 |
| Annual yield | 10.88% |
| Official remaining business days | 755 |

The worked example supplies no sell-yield/base-PU pair for this date. Those CSV fields are deliberately empty, not zero or reconstructed. Early-exit and mark-to-market scenarios therefore return `missing_quote`. Additional official quote contexts belong to M2. IPCA+ belongs to M3.

## Calendar

`calendar.json` transcribes the dates in ANBIMA's [2012](https://www.anbima.com.br/feriados/fer_nacionais/2012.asp), [2013](https://www.anbima.com.br/feriados/fer_nacionais/2013.asp), [2014](https://www.anbima.com.br/feriados/fer_nacionais/2014.asp), and [2015](https://www.anbima.com.br/feriados/fer_nacionais/2015.asp) national financial-market holiday tables. Only 2012-01-01 through 2015-12-31 is covered.

Weekends and these listed holidays are excluded. The [ANBIMA calendar notes](https://www.anbima.com.br/feriados/) distinguish this list from municipal holidays and bank public-service closures at year end. No extra closures are guessed. Count settlement inclusive and maturity exclusive. The official example independently establishes 755 business days. D0 on 2012-01-03 gives 756; Carnival settlement from Friday 2012-02-17 advances to Wednesday 2012-02-22.

## Independent scenario expectations

`scenarios.json` contains derived hypothetical results, not official quotes. The baseline is fixed at 733.86, baseline yield at 0.1088, and DU at 755.

These fixed results were calculated independently of Go production code using Ruby standard-library BigDecimal/BigMath with 60-digit precision:

```ruby
require 'bigdecimal'
require 'bigdecimal/math'
precision = 60
base = BigDecimal('733.86')
term = BigDecimal('755') / 252
factor = base * BigMath.exp(BigMath.log(BigDecimal('1.1088'), precision) * term, precision)
['0.0888', '0.1088', '0.1288'].each do |yield_text|
  pu = factor / BigMath.exp(BigMath.log(1 + BigDecimal(yield_text), precision) * term, precision)
  puts [pu.to_s('F'), (pu / base - 1).to_s('F')]
end
```

Theoretical baseline before official truncation: 733.868652711263568.
After two-decimal truncation: 733.86.
Scenario prices: 774.991898043259639 and 695.588950836608457.
Scenario fractional changes: 0.056048698720818193 and -0.052150340887078656.

Tests use absolute PU tolerance 1e-9 and fractional-change tolerance 1e-12. These tolerances detect a one-business-day error in either nonzero shock. Standalone official pricing uses the specification's BRL 0.01 tolerance after truncation.

Only factual numeric/date extracts are bundled; source documents and their explanatory prose are not redistributed. Attribution remains with Tesouro Nacional/Tesouro Direto and ANBIMA. The project MIT license does not relicense third-party source material. These fixtures are historical educational data, not executable quotes or investment advice.
