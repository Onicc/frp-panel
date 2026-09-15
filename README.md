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

## Run the controller

Generate a secret first:

```bash
export APP_GLOBAL_SECRET="$(openssl rand -hex 32)"
export PUBLIC_HOST="panel.example.com"
export APP_ENABLE_REGISTER=true
docker compose up -d
```

Create the initial Owner in the Web console, then set `APP_ENABLE_REGISTER=false` and run `docker compose up -d` again. The Web/API listener is `9000`, Agent RPC is `9001`, and the built-in default FRPS uses `7000`. Put the Web endpoint behind HTTPS and leave `APP_COOKIE_SECURE=true` in production.

## Install an Agent

Create a 10-minute enrollment command in **Nodes → Add node**, select Linux, macOS, or Windows, and copy the generated command. The Agent installs to:

| OS | Binary | Config and state |
|---|---|---|
| Linux | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/etc/frp-panel/agent.yaml`, `/var/lib/frp-panel` |
| macOS | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/Library/Application Support/frp-panel` |
| Windows | `%ProgramFiles%\frp-panel\frp-panel-agent.exe` | `%ProgramData%\frp-panel` |

The service is named `frp-panel-agent`. Use `frp-panel-agent service status`, `restart`, or `uninstall --purge`; run `frp-panel-agent doctor --json` for capability and configuration diagnostics.

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
