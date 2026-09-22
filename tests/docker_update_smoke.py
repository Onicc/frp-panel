#!/usr/bin/env python3
"""Isolated Docker launcher promotion/rollback test; no production volumes."""

import json
import pathlib
import shutil
import socket
import subprocess
import tempfile
import time
import urllib.request
import uuid


IMAGE = "frpp-local-smoke-master:20260922"


def docker(*args, check=True):
    return subprocess.run(["docker", *args], check=check, capture_output=True, text=True)


def free_port():
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


def wait_for(label, check, timeout=90):
    deadline = time.monotonic() + timeout
    last = None
    while time.monotonic() < deadline:
        try:
            if check():
                print(f"PASS {label}", flush=True)
                return
        except (OSError, ValueError, AssertionError) as error:
            last = error
        time.sleep(0.5)
    raise AssertionError(f"timed out: {label}; last={last}")


def main():
    suffix = uuid.uuid4().hex[:10]
    name = f"frpp-update-smoke-{suffix}"
    extractor = f"{name}-extract"
    port = free_port()
    image_env = docker("image", "inspect", "--format", "{{range .Config.Env}}{{println .}}{{end}}", IMAGE).stdout.splitlines()
    image_commit = next(value.split("=", 1)[1] for value in image_env if value.startswith("FRP_PANEL_IMAGE_COMMIT="))
    try:
        with tempfile.TemporaryDirectory(prefix="frpp-update-smoke-") as directory:
            root = pathlib.Path(directory)
            data = root / "data"
            data.mkdir(mode=0o777)
            data.chmod(0o777)
            docker("create", "--name", extractor, IMAGE)
            image_binary = root / "image-binary"
            docker("cp", f"{extractor}:/usr/local/bin/frp-panel", str(image_binary))
            docker("rm", extractor)

            commit_ok = "b" * 40
            commit_bad = "c" * 40
            update = data / "update"
            update.mkdir(mode=0o777)
            update.chmod(0o777)

            def pending(commit, executable, operation):
                binary_dir = update / "bin" / commit
                binary_dir.mkdir(parents=True, mode=0o777)
                binary_dir.chmod(0o777)
                binary = binary_dir / "frp-panel"
                shutil.copyfile(executable, binary)
                binary.chmod(0o755)
                (update / "pending.json").write_text(json.dumps({
                    "operationId": operation, "commit": commit, "version": "edge",
                    "imageCommit": image_commit,
                }))

            pending(commit_ok, image_binary, "success-test")
            docker("run", "-d", "--name", name, "-p", f"127.0.0.1:{port}:9000",
                   "-v", f"{data}:/data", "-e", "APP_GLOBAL_SECRET=isolated-test-secret-at-least-32-chars",
                   "-e", "APP_COOKIE_SECURE=false", IMAGE)

            def healthy():
                with urllib.request.urlopen(f"http://127.0.0.1:{port}/api/v2/health", timeout=2) as response:
                    return response.status == 200

            wait_for("staged Master starts", healthy)
            wait_for("staged binary promoted", lambda: json.loads((update / "status.json").read_text())["state"] == "succeeded")
            assert json.loads((update / "active.json").read_text())["commit"] == commit_ok

            docker("stop", name)
            failed_binary = root / "failed-binary"
            failed_binary.write_text("#!/bin/sh\nexit 1\n")
            pending(commit_bad, failed_binary, "failure-test")
            docker("start", name)
            wait_for("failed startup rolled back", lambda: json.loads((update / "status.json").read_text()).get("operationId") == "failure-test" and json.loads((update / "status.json").read_text())["state"] == "rolled_back")
            wait_for("previous Master remains healthy", healthy)
            assert json.loads((update / "active.json").read_text())["commit"] == commit_ok

            docker("stop", name)
            active_binary = update / "bin" / commit_ok / "frp-panel"
            active_binary.write_text("#!/bin/sh\nexit 1\n")
            active_binary.chmod(0o755)
            docker("start", name)
            wait_for("failed active binary falls back to image", lambda: not (update / "active.json").exists())
            wait_for("image Master remains healthy", healthy)
    finally:
        docker("rm", "-f", "-v", name, check=False)
        docker("rm", "-f", "-v", extractor, check=False)
        print(f"Cleaned isolated Docker update test {name}", flush=True)


if __name__ == "__main__":
    main()
