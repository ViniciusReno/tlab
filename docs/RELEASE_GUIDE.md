# Launching Tesouro Lab

This package contains one application binary and documentation. No Go, Python,
Node.js, database installation, configuration file, or Docker is needed to run it.

## Verify and extract

Choose the archive matching your operating system and processor. Verify its
SHA-256 against the accompanying `SHA256SUMS` before extraction:

- Linux: `sha256sum --check SHA256SUMS` (with the downloaded archives alongside it).
- macOS: `shasum -a 256 tesouro-lab-VERSION-darwin-arm64.tar.gz`, then compare the
  printed digest with that archive's line in `SHA256SUMS`.
- Windows PowerShell: `Get-FileHash .\tesouro-lab-VERSION-windows-amd64.zip -Algorithm SHA256`,
  then compare the digest with that archive's line in `SHA256SUMS`.

The Linux command requires every listed archive to be present; missing downloads
are reported separately. Substitute the actual version and architecture in names.

Extract the archive and open a terminal in its extracted directory.

## Offline historical demo

Linux and macOS:

```sh
./tesouro-lab demo
```

Windows PowerShell:

```powershell
.\tesouro-lab.exe demo
```

Open `http://127.0.0.1:8080` in a browser. Stop with Ctrl+C. If the port is already
occupied, use `demo --port 8081` and open the URL printed by the program.
The demo works offline with embedded historical fixtures and a private in-memory
database. Its portfolio edits disappear at shutdown. It never opens your persistent
database or synchronizes official quotes. External attribution links need internet.

## Browser language and input formats

Choose **English** or **Português** above the main navigation. English is the
default; a local browser cookie remembers an explicit choice. Both languages
work offline and without JavaScript. CLI help and JSON remain English/canonical.

Both interfaces use explicit BRL amounts, decimal points, and ISO `YYYY-MM-DD`
dates. Enter numbers without grouping separators, such as `10000.50` for an
amount or `8.88` for a yield percentage. The interface explains these conventions
in the selected language. Raw official data and private portfolio notes are
preserved unchanged.

## Persistent local data

Run `./tesouro-lab sync`, then `./tesouro-lab` (PowerShell:
`.\tesouro-lab.exe sync`, then `.\tesouro-lab.exe`). Synchronization needs internet;
the server reads stored quotes without automatic synchronization.
The default database is `tesouro-lab.db` inside `tesouro-lab` under the operating
system's user configuration directory. The server prints the resolved directory.
For a custom location, pass the same `--data-dir PATH` to both commands.

Back up the data directory with the application stopped before replacing a binary.
Versioned migrations run at startup. This build adds no migration beyond M4's
`003_portfolio.sql`; it reuses existing quote and portfolio data. Downgrading after
a later schema change is not guaranteed.

## Scope and support

Prefixado and IPCA+ without coupons have same-date hypothetical yield scenarios.
Selic supports quotes, history, and official base-PU portfolio valuation only.
Values are gross; no tax engine, inflation forecast, or financial recommendation
is provided. Read `RELEASE_NOTES.md` for validation status and known limits.

Maintainer build targets are Linux amd64/arm64, Windows amd64, and macOS
amd64/arm64. A successful cross-build is not execution validation. Native support
must be confirmed by the release's CI results before publication. Packages are
currently unsigned; no code-signing or notarization claim is made.
