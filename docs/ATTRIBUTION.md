# License and attribution

Tesouro Lab code is distributed under the MIT license in `LICENSE`.
All application assets are embedded locally; no chart library, font service,
frontend build toolchain, or CDN is required.

Official factual extracts retain their provenance and are not relicensed by the
project's MIT license:

- Tesouro Nacional / Tesouro Direto: daily quote records from the
  [Tesouro Transparente dataset](https://www.tesourotransparente.gov.br/ckan/dataset/taxas-dos-titulos-ofertados-pelo-tesouro-direto).
  Its metadata reports the Open Database License (ODbL). Raw official field names
  are preserved. Fixtures redistribute small factual extracts, not source documents.
- Tesouro Direto: the historical Prefixado demo transcribes factual inputs from
  the [official Prefixado methodology](https://www.tesourodireto.com.br/documents/d/guest/tesouro_prefixado).
- ANBIMA: the business-day calendars transcribe factual holiday dates from the
  [annual financial-market holiday tables](https://www.anbima.com.br/feriados/).
  Source URLs, retrieval dates, hashes, and coverage are recorded in embedded fixtures.

Detailed extract provenance, independent calculation expectations, and limitations
are in `data/demo/README.md`, `data/calendar/README.md`,
`internal/pricing/testdata/README.md`, and `internal/pricing/testdata/M3_README.md`
in the source repository. Hypothetical quantities and derived scenario outputs are
educational inputs/results, not official market records.

## Compiled Go dependencies

The pinned module versions are recorded in `go.mod` and `go.sum`. License notices
for all compiled dependencies are included in each package's
`THIRD_PARTY_LICENSES.txt`; the packager obtains them from the downloaded modules.
This includes modernc.org/sqlite and its transitive dependencies. SQLite itself is
[public domain](https://www.sqlite.org/copyright.html); the Go driver and other
compiled modules retain their respective licenses.
