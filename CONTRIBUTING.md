# Contributing to Tesouro Lab

Tesouro Lab implements the M1 offline demo, M2 official synchronization/market/history, and M3 IPCA+ scenarios within verified calendar coverage plus Selic quote support. M4 adds local portfolio positions, official gross valuation, and validated Prefixado comparisons. See [M4 notes](docs/M4_IMPLEMENTATION.md) for the current behavior and validation, and [M3 notes](docs/M3_IMPLEMENTATION.md) for the versioned calendar coverage. Broader usability work and Brazilian Portuguese support are deferred until the planned V1 work is complete; authored project content remains English.

## Read first

1. Read [AGENTS.md](AGENTS.md).
2. Read the relevant sections of the [V1 specification](docs/V1_SPEC.md).
3. Inspect the affected package boundaries, tests, and fixture provenance.

V1 is an educational, local-first application. Keep changes focused on verified data, transparent calculations, and understandable explanations. Scope changes require an explicit maintainer decision.

## Language and repository structure

Write code identifiers, comments, test names, documentation, UI copy, help text, errors, logs, commit messages, issues, pull requests, and release notes in English. Preserve official proper names and verbatim source data, and explain them in English.

Application code, CLI, browser assets, migrations, fixtures, tests, and documentation belong in this repository and share one root Go module and release lifecycle. Follow the [proposed layout](docs/V1_SPEC.md#10-proposed-repository-layout), creating packages only when the current milestone needs them.

## Development workflow

The root module pins Go 1.25.0 as its minimum version and records dependency versions. With Go 1.25 or newer installed, run:

```sh
go run ./cmd/tesouro-lab demo
```

Run it from the repository root, open `http://127.0.0.1:8080`, and stop with Ctrl+C. The first source build may download dependencies. The packaged demo must work offline. Neither path requires Docker or a frontend build toolchain.

The demo must use private, temporary SQLite state and must not access the persistent database. Do not add real portfolio data, local database files, or private notes to the repository or test fixtures.

## Validation

Run the relevant tests while developing and these checks before submitting a normal code change:

```sh
go fmt ./...
go vet ./...
go test ./...
```

Run `go test -race ./...` where supported and relevant. Automated checks must also verify a build without CGO. Keep routine tests deterministic and offline; separate explicitly requested live-source integration tests.

Sync tests replace the HTTP transport with committed official metadata/CSV fixtures; routine tests never contact the live source. For a manual live check, run `go run ./cmd/tesouro-lab sync --data-dir PATH` against a separate temporary directory and record the actual report. Do not use a personal database for test imports.

The optional `python3 internal/pricing/testdata/reference.py` command prints the independent high-precision context expectations for review. Python is not required for Go tests or application use. Never regenerate expected values at Go test runtime. See [context validation and the known source discrepancy](internal/pricing/testdata/README.md).

Financial changes require documented assumptions, official baseline provenance, independently checked expected results, explicit floating-point tolerances, and relevant edge cases. Calendar changes require revalidation against affected official fixtures. Simple/advanced views and CLI/browser delivery must use the same calculation.

For documentation-only changes, check links, examples, terminology, and consistency with the specification. Report exactly what you verified and what you could not run.

## Pull requests

Explain the concrete problem and resulting behavior, why the change belongs in V1, the checks performed, and any limitations. Update affected documentation in the same change. Label planned or unverified behavior explicitly.

Keep feature work, broad refactoring, and dependency upgrades separate unless they are necessary for the same change. Do not add investment advice, predictions, unneeded infrastructure, or unsupported financial approximations.

## Release documentation

Document supported operating systems/architectures, checksums, launch commands, known limitations, and migration or compatibility changes. Verify the extracted release's demo command without requiring Go or network access before claiming the release supports that workflow.

Record implemented behavior and validation in English release notes. Never present future milestones or pending checks as completed work.
