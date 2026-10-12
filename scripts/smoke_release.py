"""Verify an extracted native release with an empty PATH and loopback-only HTTP."""

import argparse
import http.cookiejar
import json
import os
from pathlib import Path
import signal
import socket
import subprocess
import tempfile
import time
import urllib.error
import urllib.request


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=Path)
    parser.add_argument("--version", required=True)
    parser.add_argument("--port", type=int, default=8080,
                        help="Local test port; release CI uses the documented default 8080.")
    args = parser.parse_args()
    if not 1 <= args.port <= 65535:
        parser.error("Port must be between 1 and 65535.")
    # Fail before launch if another application owns the test address. Never
    # mistake its pages for the extracted binary under review.
    with socket.socket() as probe:
        probe.bind(("127.0.0.1", args.port))
    base_url = f"http://127.0.0.1:{args.port}"
    binary = args.binary.resolve(strict=True)
    with tempfile.TemporaryDirectory(prefix="tlab-smoke-") as directory:
        env = {**os.environ, "PATH": "", "GOPROXY": "off", "GOSUMDB": "off",
               "HTTP_PROXY": "http://127.0.0.1:1", "HTTPS_PROXY": "http://127.0.0.1:1",
               "HOME": directory, "USERPROFILE": directory, "XDG_CONFIG_HOME": directory,
               "APPDATA": directory, "LOCALAPPDATA": directory}
        def run(*arguments):
            return subprocess.check_output([str(binary), *arguments], cwd=directory,
                                           env=env, text=True, timeout=20)
        assert run("version").strip() == args.version
        for instrument, date, annual_yield in [
            ("prefixado:2015-01-01", "2012-01-03", "8.88"),
            ("ipca:2015-05-15", "2012-02-17", "3.47"),
        ]:
            result = json.loads(run("analyze", instrument, "--source", "demo", "--basis",
                                    "purchase", "--date", date, "--yield", annual_yield))
            assert result["build_version"] == args.version and result["quote_date"] == date
        opener = urllib.request.build_opener(urllib.request.ProxyHandler({}),
                                            urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
        # Exercise the documented default port and command, rather than a custom launcher.
        with tempfile.TemporaryFile(mode="w+") as log:
            command = [str(binary), "demo"]
            if args.port != 8080:
                command.extend(["--port", str(args.port)])
            process = subprocess.Popen(command, cwd=directory, env=env,
                                       stdout=log, stderr=log,
                                       creationflags=subprocess.CREATE_NEW_PROCESS_GROUP if os.name == "nt" else 0)
            try:
                deadline = time.monotonic() + 20
                while True:
                    if process.poll() is not None:
                        log.seek(0)
                        raise RuntimeError(log.read())
                    try:
                        with opener.open(base_url + "/", timeout=1) as response:
                            assert b"Historical demo" in response.read()
                        break
                    except (urllib.error.URLError, TimeoutError):
                        if time.monotonic() >= deadline:
                            raise
                        time.sleep(0.1)
                for path, expected in [
                    ("/learn", b"Learn"), ("/portfolio", b"temporary"),
                    ("/static/style.css", b"--ink"),
                    ("/api/scenario?bond=ipca:2015-05-15&source=demo&basis=mark_to_market"
                     "&date=2012-02-17&yield=3.51", b'"scenario_pu"'),
                ]:
                    with opener.open(base_url + path, timeout=5) as response:
                        assert expected in response.read(), path
                for path, expected in [
                    ("/playground?lang=pt-BR", "Preço no cenário hipotético"),
                    ("/learn", "Fontes e limitações"),
                    ("/portfolio", "alterações temporárias"),
                ]:
                    with opener.open(base_url + path, timeout=5) as response:
                        assert response.headers["Content-Language"] == "pt-BR", path
                        body = response.read().decode("utf-8")
                        assert '<html lang="pt-BR">' in body and expected in body, path
                with opener.open(base_url + "/api/scenario?bond=prefixado:2015-01-01"
                                 "&source=demo&basis=purchase&date=2012-01-03&yield=8.88",
                                 timeout=5) as response:
                    result = json.load(response)
                    assert result["build_version"] == args.version
                    assert result["quote_date"] == "2012-01-03"
                with opener.open(base_url + "/learn?lang=en", timeout=5) as response:
                    assert response.headers["Content-Language"] == "en"
                    assert b'<html lang="en">' in response.read()
                log.seek(0)
                startup = log.read()
                assert f"127.0.0.1:{args.port}" in startup and "Ctrl+C" in startup
            finally:
                if process.poll() is None:
                    process.send_signal(signal.CTRL_BREAK_EVENT if os.name == "nt" else signal.SIGINT)
                process.wait(timeout=10)
            assert process.returncode == 0, "Shutdown failed"
        assert not list(Path(directory).rglob("*.db")), "Demo wrote a persistent database"
    print("Extracted demo, both browser languages, embedded pages/assets, both analytical instruments, and shutdown passed.")
    print("No Go or runtime executable was available through PATH. External proxies were disabled.")
    print("This smoke check is not an OS-level network isolation test.")


if __name__ == "__main__":
    main()
