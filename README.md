# frp-panel v2

An open-source FRP control plane with a secure Web console and a cross-platform node Agent.

> v2 is a clean break. It does not migrate v1 databases or preserve the old client CLI.

## What changed

- Separate `frp-panel` controller and lightweight `frp-panel-agent` deliverables.
- Linux amd64/arm64, macOS amd64/arm64, and Windows amd64/arm64 Agent builds.
- Correct system installation paths; the bootstrap command never installs into the current directory.
- Vite 8 / React 19 bilingual console with controlled mutation dialogs.
- Secure defaults: TLS verification on; privileged features off; Argon2id passwords; role-derived permissions; same-origin WebSockets; protected Agent config.
- Reproducible multi-architecture images: `onicc/frp-panel` and `onicc/frp-panel-agent`.
- SHA-256 release checksums, SBOMs, build provenance, and pinned GitHub Actions.

## Deploy Master, Server, and Client

Master provides the control plane, while the default Server (FRPS) is embedded in the same controller container. A Client is `frp-panel-agent` plus managed FRPC on a node. Start by creating the Docker Compose `.env` file:

```bash
openssl rand -hex 32
```

```dotenv
APP_GLOBAL_SECRET=REPLACE_ME
APP_COOKIE_SECURE=true
APP_ENABLE_REGISTER=true
PUBLIC_HOST=panel.example.com
MASTER_API_SCHEME=https
CLIENT_API_URL=https://panel.example.com
CLIENT_RPC_URL=wss://panel.example.com
```

Replace `REPLACE_ME` with the random value generated above.

Point an HTTPS reverse proxy at `127.0.0.1:9000`, then start the stack:

```bash
docker compose config --quiet
docker compose pull
docker compose up -d
```

Immediately set `APP_ENABLE_REGISTER=false` and reapply Compose after creating the initial Owner. Embedded FRPS uses `7000`; every proxy remote port must also be published explicitly.

A Client must be installed with the ten-minute one-time command from **Nodes → Add node**. The console provides Linux, macOS, and Windows commands with the public API/RPC endpoints included.

See the [deployment guide](docs/en/deployment.md) for reverse proxy, port, backup, upgrade, and Client verification instructions.

## Client installation locations

The console-generated command installs the Agent to:

| OS | Binary | Config and state |
|---|---|---|
| Linux | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/etc/frp-panel/agent.yaml`, `/var/lib/frp-panel` |
| macOS | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/Library/Application Support/frp-panel` |
| Windows | `%ProgramFiles%\frp-panel\frp-panel-agent.exe` | `%ProgramData%\frp-panel` |

The service is named `frp-panel-agent`. The binary is not automatically added to `PATH`; use the full path above or the native service manager.

## Development

Requires Go 1.27.1 and Node.js 24 LTS.

```bash
corepack enable
cd www && pnpm install --frozen-lockfile && pnpm build && cd ..
go test ./...
go build ./cmd/frpp ./cmd/frp-panel-agent
```

See [architecture](docs/ARCHITECTURE_V2.md), [optimization review](docs/OPTIMIZATION.md), [platform support](docs/SUPPORT_MATRIX.md), [security review](docs/SECURITY.md), and [OpenAPI](api/openapi.yaml).

## Releases

- Every successful `main` push updates the `edge` binaries and the `edge` tags on both Docker images.
- A `v*` tag creates a stable GitHub release and publishes that tag plus `latest` to Docker Hub.
- Stable releases include checksums, SBOMs, and GitHub artifact attestations. Apple notarization and Windows Authenticode are not currently provided.

## License and attribution

AGPL-3.0. This fork is maintained by Onicc and is derived from [VaalaCat/frp-panel](https://github.com/VaalaCat/frp-panel). See [NOTICE](NOTICE).
