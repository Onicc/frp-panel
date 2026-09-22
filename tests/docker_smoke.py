#!/usr/bin/env python3
"""Isolated Docker smoke test for multi-Client Tunnel identity and cleanup.

From the repository root, build the images with:
  docker build --target master -t frpp-local-smoke-master:20260922 .
  docker build --target agent -t frpp-local-smoke-agent:20260922 .
Then run this script. It removes all containers, anonymous volumes, temporary
credentials and the test network; remove the two test image tags afterwards.
"""

import http.cookiejar
import json
import pathlib
import socket
import subprocess
import tempfile
import time
import urllib.error
import urllib.request
import uuid


MASTER_IMAGE = "frpp-local-smoke-master:20260922"
AGENT_IMAGE = "frpp-local-smoke-agent:20260922"


def docker(*args, check=True):
    return subprocess.run(
        ["docker", *args], check=check, text=True, capture_output=True
    )


def free_port():
    with socket.socket() as listener:
        listener.bind(("127.0.0.1", 0))
        return listener.getsockname()[1]


def eventually(label, fn, timeout=75):
    deadline = time.monotonic() + timeout
    last = None
    while time.monotonic() < deadline:
        try:
            result = fn()
            if result:
                print(f"PASS {label}", flush=True)
                return result
        except (AssertionError, OSError, subprocess.CalledProcessError) as exc:
            last = exc
        time.sleep(1)
    raise AssertionError(f"timed out: {label}; last={last}")


def main():
    suffix = uuid.uuid4().hex[:10]
    network = f"frpp-smoke-{suffix}"
    names = []
    port = free_port()
    base = f"http://127.0.0.1:{port}/api/v2"
    opener = urllib.request.build_opener(
        urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar())
    )

    def api(method, path, payload=None, status=200, headers=None):
        body = None if payload is None else json.dumps(payload).encode()
        request = urllib.request.Request(
            base + path,
            data=body,
            headers={"Content-Type": "application/json", **(headers or {})},
            method=method,
        )
        try:
            with opener.open(request, timeout=8) as response:
                actual = response.status
                data = response.read()
        except urllib.error.HTTPError as exc:
            actual = exc.code
            data = exc.read()
        if actual != status:
            raise AssertionError(f"{method} {path}: {actual} != {status}: {data[:500]!r}")
        return json.loads(data) if data else None

    def start(name, image, options=(), command=()):
        full = f"{network}-{name}"
        result = docker("run", "-d", "--name", full, "--network", network, *options, image, *command)
        names.append(full)
        return result.stdout.strip()

    def endpoint(server, remote_port, expected):
        result = docker(
            "run", "--rm", "--network", network, "alpine:3.23",
            "wget", "-qO-", "-T", "2", f"http://{network}-{server}:{remote_port}/",
            check=False,
        )
        return result.returncode == 0 and expected in result.stdout

    docker("network", "create", network)
    try:
        with tempfile.TemporaryDirectory(prefix="frpp-docker-smoke-") as scratch:
            directory = pathlib.Path(scratch)
            start(
                "master", MASTER_IMAGE,
                options=(
                    "-p", f"127.0.0.1:{port}:9000",
                    "-e", "APP_GLOBAL_SECRET=smoke-only-secret-value-with-32-chars",
                    "-e", "APP_COOKIE_SECURE=false",
                    "-e", "APP_ENABLE_REGISTER=true",
                    "-e", f"PUBLIC_URL=http://{network}-master:9000",
                ),
            )
            eventually("Master health", lambda: api("GET", "/health")["status"] == "ok")
            api("POST", "/auth/register", {
                "username": "smoke", "email": "smoke@example.test",
                "password": "smoke-password-for-local-only",
            }, 201)

            def add_server(name, bind_port):
                created = api("POST", "/servers", {
                    "serverId": name, "address": f"{network}-{name}",
                    "bindPort": bind_port,
                }, 201)
                start(
                    name, MASTER_IMAGE,
                    options=(
                        "-e", f"PUBLIC_URL=http://{network}-master:9000",
                        "-e", f"SERVER_ENROLLMENT_TOKEN={created['enrollment']['token']}",
                    ),
                    command=("server", "--config", "/data/server.yaml"),
                )
                return created["server"]["id"]

            server_a = add_server("server-a", 7001)
            for label in ("a", "b"):
                content = directory / label
                content.mkdir()
                (content / "index.html").write_text(f"backend-{label}\n")
                start(
                    f"backend-{label}", "alpine:3.23",
                    options=("-v", f"{content}:/srv:ro"),
                    command=("sh", "-c", "apk add --no-cache busybox-extras >/dev/null && httpd -f -p 7443 -h /srv"),
                )

            client_ids = {}
            for label in ("a", "b"):
                created = api("POST", "/clients", {"clientId": label}, 201)
                client_id = created["client"]["id"]
                client_ids[label] = client_id
                enrolled = api("POST", "/agent/enroll", {
                    "token": created["enrollment"]["token"]
                }, 200)
                config = directory / f"agent-{label}.yaml"
                config.write_text(
                    "version: 3\n"
                    f"master:\n  api_url: http://{network}-master:9000\n"
                    f"  rpc_url: ws://{network}-master:9000\n"
                    f"credentials:\n  client_id: {client_id}\n"
                    f"  secret: {enrolled['secret']}\n"
                )
                config.chmod(0o644)
                start(
                    f"agent-{label}", AGENT_IMAGE,
                    options=("-v", f"{config}:/etc/frp-panel/agent.yaml:ro"),
                )

            def add_tunnel(label, server_id, name, remote_port):
                return api("POST", "/tunnels", {
                    "name": name, "clientId": client_ids[label],
                    "serverId": server_id, "type": "tcp",
                    "localHost": f"{network}-backend-{label}",
                    "localPort": 7443, "remotePort": remote_port,
                }, 201)["tunnel"]

            a = add_tunnel("a", server_a, "same", 60001)
            b = add_tunnel("b", server_a, "same", 60002)
            api("POST", "/tunnels", {
                "name": "same", "clientId": client_ids["a"],
                "serverId": server_a, "type": "tcp",
                "localHost": f"{network}-backend-a", "localPort": 7443,
                "remotePort": 60003,
            }, 409)
            print("PASS same name across Clients; duplicate within Client rejected", flush=True)
            eventually("Client A proxy", lambda: endpoint("server-a", 60001, "backend-a"))
            eventually("Client B proxy", lambda: endpoint("server-a", 60002, "backend-b"))

            server_b = add_server("server-b", 7002)
            secondary = add_tunnel("a", server_b, "secondary", 60003)
            eventually("new Server proxy", lambda: endpoint("server-b", 60003, "backend-a"))
            eventually("existing proxy survives Server addition", lambda: endpoint("server-a", 60001, "backend-a"))
            eventually("other Client survives Server addition", lambda: endpoint("server-a", 60002, "backend-b"))

            docker("restart", f"{network}-server-a")
            eventually("Server restart restores first Client", lambda: endpoint("server-a", 60001, "backend-a"))
            eventually("Server restart restores second Client", lambda: endpoint("server-a", 60002, "backend-b"))

            docker("restart", f"{network}-agent-a")
            eventually("reconnect restores first pair", lambda: endpoint("server-a", 60001, "backend-a"))
            eventually("reconnect restores second pair", lambda: endpoint("server-b", 60003, "backend-a"))
            api("DELETE", f"/tunnels/{secondary['id']}", status=204)
            eventually("last Tunnel cleans exact pair", lambda: not endpoint("server-b", 60003, "backend-a"))
            eventually("other Server pair remains", lambda: endpoint("server-a", 60001, "backend-a"))
            api("DELETE", f"/tunnels/{a['id']}", status=204)
            eventually("first Client proxy removed", lambda: not endpoint("server-a", 60001, "backend-a"))
            eventually("second Client same-name proxy remains", lambda: endpoint("server-a", 60002, "backend-b"))
            api("DELETE", f"/tunnels/{b['id']}", status=204)
            eventually("second Client proxy removed", lambda: not endpoint("server-a", 60002, "backend-b"))
            for server in ("server-a", "server-b"):
                logs = docker("logs", f"{network}-{server}")
                if "already exists" in logs.stdout or "already exists" in logs.stderr:
                    raise AssertionError(f"duplicate FRPS proxy registration on {server}")
            print("PASS no duplicate FRPS proxy registration", flush=True)
    except Exception:
        for name in names:
            logs = docker("logs", "--tail", "50", name, check=False)
            print(f"--- {name} ---\n{logs.stdout[-3000:]}\n{logs.stderr[-3000:]}", flush=True)
        raise
    finally:
        for name in reversed(names):
            docker("rm", "-f", "-v", name, check=False)
        docker("network", "rm", network, check=False)
        print(f"Cleaned Docker containers and network for {network}", flush=True)


if __name__ == "__main__":
    main()
