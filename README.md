# frp-panel v2

An open-source FRP control plane with a secure Web console and a cross-platform node Agent.

> v2 is a clean break. It does not migrate v1 databases or preserve the old client CLI.

## What changed

- Separate Master control plane, independently deployable FRPS data planes, and lightweight `frp-panel-agent` nodes.
- Linux amd64/arm64, macOS amd64/arm64, and Windows amd64/arm64 Agent builds.
- Correct system installation paths; the bootstrap command never installs into the current directory.
- Vite 8 / React 19 bilingual console with controlled mutation dialogs.
- Secure defaults: TLS verification on; privileged features off; Argon2id passwords; role-derived permissions; same-origin WebSockets; protected Agent config.
- Reproducible multi-architecture images: `onicc/frp-panel` and `onicc/frp-panel-agent`.
- SHA-256 release checksums, SBOMs, build provenance, and pinned GitHub Actions.

## Deploy Master, Server, and Client

Deploy exactly one Master with the repository Compose file. Master owns the Web console and desired state but carries no proxy traffic. Create any number of Servers in **Servers**; each creation returns a one-time Docker Compose deployment for an independent FRPS host. Create Clients in **Nodes**, install them with the generated Linux, macOS, or Windows command, then assign their FRPS routes from the same page. One Client may use multiple Servers.

Both Server and Client bootstrap tokens expire after ten minutes and can be redeemed only once. Persistent credentials are hashed in Master and stored only in protected files or volumes on the managed host.

See the [deployment guide](docs/en/deployment.md) for the complete Master → Server → Client rollout, raw Compose files, environment configuration, ports, and backup checks.

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
