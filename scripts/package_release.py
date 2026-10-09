"""Build local release archives; Python is a maintainer tool, never an app runtime."""

import argparse
import hashlib
import os
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile
import zipfile

ROOT = Path(__file__).resolve().parents[1]
TARGETS = ("linux/amd64", "linux/arm64", "windows/amd64", "darwin/amd64", "darwin/arm64")
DOCUMENTS = {
    "LICENSE": "LICENSE",
    "docs/RELEASE_GUIDE.md": "START_HERE.md",
    "docs/RELEASE_NOTES.md": "RELEASE_NOTES.md",
    "docs/ATTRIBUTION.md": "ATTRIBUTION.md",
}


def licenses(env):
    modules = subprocess.check_output([
        "go", "list", "-deps", "-f",
        "{{if .Module}}{{.Module.Path}}|{{.Module.Version}}|{{.Module.Dir}}{{end}}",
        "./cmd/tesouro-lab",
    ], cwd=ROOT, env=env, text=True)
    goroot = subprocess.check_output(["go", "env", "GOROOT"], text=True).strip()
    go_license = Path(goroot) / "LICENSE"
    if not go_license.exists():
        go_license = Path(goroot).parent / "LICENSE"  # Homebrew layout.
    notices = ["Go standard library\n\n" + go_license.read_text(encoding="utf-8")]
    for module in sorted(set(modules.splitlines())):
        if not module.strip():
            continue
        path, version, directory = module.split("|", 2)
        if not version:
            continue  # Main module license is shipped separately.
        files = sorted(file for file in Path(directory).iterdir()
                       if file.is_file() and file.name.upper().startswith(
                           ("LICENSE", "COPYING", "NOTICE", "AUTHORS", "COPYRIGHT")))
        if not any(file.name.upper().startswith(("LICENSE", "COPYING")) for file in files):
            raise RuntimeError(f"Missing license notice for {path} {version}")
        notices.append(f"\n\n=== {path} {version} ===\n")
        for file in files:
            notices.append(f"\n--- {file.name} ---\n{file.read_text(encoding='utf-8')}\n")
    return "".join(notices)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--version", required=True)
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument("--target", choices=TARGETS, action="append")
    args = parser.parse_args()
    if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9.-]{0,79}", args.version):
        parser.error("Version must use letters, digits, dots, or hyphens.")
    # Refuse to overwrite existing archives or checksums.
    out = args.out.resolve()
    out.mkdir(parents=True, exist_ok=False)
    checksums = []
    for target in args.target or TARGETS:
        system, arch = target.split("/")
        name = f"tesouro-lab-{args.version}-{system}-{arch}"
        with tempfile.TemporaryDirectory(prefix="tlab-package-") as temporary:
            package = Path(temporary) / name
            package.mkdir()
            binary = package / ("tesouro-lab.exe" if system == "windows" else "tesouro-lab")
            env = {**os.environ, "CGO_ENABLED": "0", "GOOS": system, "GOARCH": arch}
            subprocess.run([
                "go", "build", "-trimpath", "-ldflags",
                f"-X github.com/ViniciusReno/tlab/internal/app.BuildVersion={args.version}",
                "-o", str(binary), "./cmd/tesouro-lab",
            ], cwd=ROOT, env=env, check=True)
            for source, destination in DOCUMENTS.items():
                (package / destination).write_bytes((ROOT / source).read_bytes())
            (package / "THIRD_PARTY_LICENSES.txt").write_text(licenses(env), encoding="utf-8")
            if system == "windows":
                archive = out / f"{name}.zip"
                with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED) as zipped:
                    for file in sorted(package.iterdir()):
                        zipped.write(file, f"{name}/{file.name}")
            else:
                archive = out / f"{name}.tar.gz"
                with tarfile.open(archive, "w:gz") as tar:
                    tar.add(package, arcname=name)
            digest = hashlib.sha256(archive.read_bytes()).hexdigest()
            checksums.append(f"{digest}  {archive.name}\n")
            print(f"Built {target}: {archive}", flush=True)
    (out / "SHA256SUMS").write_text("".join(checksums), encoding="ascii")


if __name__ == "__main__":
    main()
