# Deployment guide

## 1. Deployment model

| Role | Count | Responsibility | Default deployment |
|---|---:|---|---|
| Master | 1 | Web console, API, identity, configuration, and connection orchestration; it carries no proxy traffic | Docker Compose |
| Server (FRPS) | 1 or more | Public data-plane entry points for FRPC and tunnel traffic | Docker Compose on every Server host |
| Client (FRPC) | 1 or more | Workload host running the Client Agent and managed FRPC | Console-generated OS installation command |

One Master manages every Server, Client, and Tunnel. A Client is the physical workload host; the Agent is its management process. Every Tunnel directly selects one Client and one Server. Master stores desired state and best-effort creates the underlying FRPC connection when the Client is online and both resources are enrolled; Tunnels for the same Client and Server share it.

Deploy in this order: **Master → Server → Client → Tunnel**.

## 2. Deploy the single Master

Keep `compose.yaml` and `.env` in a dedicated directory. The repository [compose.yaml](https://github.com/Onicc/frp-panel/blob/main/compose.yaml) is identical to this file.

### compose.yaml

```yaml
services:
  master:
    image: ${FRP_PANEL_IMAGE:-onicc/frp-panel:edge}
    restart: unless-stopped
    environment:
      APP_GLOBAL_SECRET: ${APP_GLOBAL_SECRET:?set a random 32+ character secret}
      APP_COOKIE_SECURE: ${APP_COOKIE_SECURE:-true}
      APP_ENABLE_REGISTER: ${APP_ENABLE_REGISTER:-false}
      APP_AGENT_INSTALL_URL: ${APP_AGENT_INSTALL_URL:-https://raw.githubusercontent.com/Onicc/frp-panel/main}
      PUBLIC_URL: ${PUBLIC_URL:?set PUBLIC_URL in .env}
    ports:
      - "127.0.0.1:9000:9000"
    volumes:
      - frp-panel-data:/data
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:9000/api/v2/health"]
      interval: 30s
      timeout: 5s
      retries: 3

volumes:
  frp-panel-data:
```

### .env

```dotenv
FRP_PANEL_IMAGE=onicc/frp-panel:edge

APP_GLOBAL_SECRET=REPLACE_WITH_A_RANDOM_32_BYTE_OR_LONGER_SECRET
APP_COOKIE_SECURE=true
APP_ENABLE_REGISTER=true
# Optional fork or internal mirror for install.sh/install.ps1.
# APP_AGENT_INSTALL_URL=https://raw.githubusercontent.com/Onicc/frp-panel/main

PUBLIC_URL=https://panel.example.com
```

- Keep `APP_GLOBAL_SECRET` unchanged and private. It must contain at least 32 random bytes; changing or losing it invalidates existing credentials.
- Enable registration only while creating the first Owner. Set `APP_ENABLE_REGISTER=false` immediately afterward and reapply the Compose configuration.
- `PUBLIC_URL` is the only Master public address to configure. It must be a complete `http://` or `https://` URL without a path; HTTPS automatically derives the `wss://` RPC endpoint.
- Set `APP_AGENT_INSTALL_URL` only for a fork or internal mirror; it must contain both `install.sh` and `install.ps1`.
- In production with `APP_COOKIE_SECURE=true`, an HTTPS reverse proxy must forward Web, API, and WebSocket traffic to `127.0.0.1:9000`; the repository Compose file does not provide public TLS.
- Pin a tested `v*` image tag in production; `edge` follows `main`.

Master does not publish port `7000` or any tunnel remote port.

The first visit to `PUBLIC_URL` shows the sign-in form. When `APP_ENABLE_REGISTER=true` and the database has no users, the page provides a **Create Owner** switch. Enter a username, email, a password of at least 12 characters, and the confirmation; the console signs in automatically. After confirming that the account can sign in again, set `APP_ENABLE_REGISTER=false` and re-apply the Compose configuration. Use **Account settings** in the top-right corner to change the password; Master verifies the current password and requires a new sign-in after a successful change.

The repository Compose file passes only the variables listed under its `environment` section into the container. Putting `APP_ALLOWED_ORIGINS`, `MASTER_*`, or `DB_*` only in the host `.env` does not apply them; add them explicitly to Compose and update port mappings and the healthcheck when customizing them.

## 3. Deploy one or more Servers (FRPS)

Each FRPS is an independent data plane on the Linux host that receives public traffic:

1. Open **Servers → Create Server** in Master.
2. Enter a unique ID, the public DNS name/IP used by FRPC, the FRPS bind port (`7000` by default), and the local Server authentication API port (`SERVER_API_PORT`, `8999` by default). The two ports must be different, and both must be free on the target host.
3. The successful dialog closes and the page displays a complete Server-specific `compose.yaml`.
4. Save it on the target Server host and apply it. The first start redeems a ten-minute one-use token and stores the permanent credential in `/data/server.yaml` inside the named volume.
5. Wait for the Server status to become online. Repeat for every additional FRPS host.

Never reuse one generated file or Server data volume on multiple machines. If a token expires or the generated file is lost, delete the un-enrolled Server (when it has no dependent Tunnels) and create the same ID again; use credential rotation for an enrolled Server.

The generated file has this structure. Master inserts its configured `PUBLIC_URL` and the one-use token automatically, so the domain is not configured again:

```yaml
services:
  frps:
    image: ${FRP_PANEL_IMAGE:-onicc/frp-panel:edge}
    restart: unless-stopped
    network_mode: host
    command:
      - server
      - --config
      - /data/server.yaml
    environment:
      PUBLIC_URL: "MASTER_PUBLIC_URL_INSERTED_BY_CONSOLE"
      SERVER_ENROLLMENT_TOKEN: "ONE_TIME_TOKEN_FROM_MASTER"
      SERVER_API_PORT: "8999"
    volumes:
      - frp-panel-server-data:/data

volumes:
  frp-panel-server-data:
```

Linux host networking lets FRPS accept new TCP/UDP remote ports without changing container port mappings. The generated `FRP_PANEL_IMAGE` value is resolved by Compose on the Server host; pin it there to the same tested version as Master in production. Restrict the host firewall and cloud security group to the FRPS bind port and the exact remote ports required by tunnels. Because host networking is enabled, `SERVER_API_PORT` (default `127.0.0.1:8999`) is the Server host loopback and is used only by the FRPS authentication plugin; it must not be exposed. It must be free and different from the FRPS bind port and every Tunnel remote port. If multiple FRPS instances share one host, give each instance a different `SERVER_API_PORT`.

Always use the file generated by the console; do not copy the Master's `.env` to a Server. After the first successful enrollment, restarts use the protected volume credential.

When a new token is deployed to a data volume that previously ran the same Server, the current release compares a non-secret token fingerprint and automatically re-enrolls. Restarts with the same token continue using the persisted credential. Legacy `server.yaml` files are migrated opportunistically. If logs still show `invalid secret`, verify that the current Server-specific Compose file is being used and remove only that Server's data volume before deploying again; never remove the Master data volume.

## 4. Install one or more Clients (FRPC)

Do not hand-write a Client Compose file or command:

1. Open **Clients → Add Client** in Master and enter a unique Client ID.
2. Select Linux, macOS, or Windows and copy the complete generated command.
3. Run it with administrator privileges on the target Client host. It installs the Client Agent, protected configuration, and native service in system locations, never in the command's current directory.
4. Wait for the Client Agent to become online. Generate a separate command for every Client.

| OS | Binary | Configuration and state |
|---|---|---|
| Linux | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/etc/frp-panel/agent.yaml`, `/var/lib/frp-panel` |
| macOS | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/Library/Application Support/frp-panel` |
| Windows | `%ProgramFiles%\frp-panel\frp-panel-agent.exe` | `%ProgramData%\frp-panel` |

See [Client / Agent operations](/en/agent) for service management, upgrade, and uninstall details.

## 5. Create Tunnels and select Servers

A newly installed Client needs no preassigned Server. Open **Tunnels → Create Tunnel**:

1. Select the Client that can reach the local service.
2. Select the Server that provides the public entry point.
3. Enter TCP/UDP, local address, local port, and public port, then confirm.

A Client may own multiple Tunnels and every Tunnel may use a different Server. Different Clients may reuse the same Tunnel name; names must be unique within a Client. Master stores each Tunnel as desired state and, when the Client is online and both Client and Server are enrolled, best-effort pushes the assembled FRPC configuration to the Client; Tunnels for the same pair share a connection. Removing the last Tunnel of a pair stops that pair's FRPC without interrupting the Client's other Server connections. The underlying Client-to-Server connection is not a separate user-managed resource.

The public port listens on the selected Server (FRPS) and forwards to the service reachable by the selected Client. Business traffic never passes through Master. A Tunnel created while a Client is offline remains stored in Master; reconnect clears old runtime FRPC connections and rebuilds them from saved Tunnels. Local ports may be `1–65535`; Server bind ports and remote ports must be `1024–65535`.

The map uses a manually set public IP in Client editing first, a recent direct Agent public-IP probe second, and the Master-observed connection IP last. The probe bypasses `HTTP_PROXY`/`HTTPS_PROXY`, but a transparent or TUN proxy may still change the result; set the Client's actual public IP manually in that case. The observed fallback is labeled as a possible proxy exit. Upgrade Master, Servers, and Client Agents together to enable the fixes and remove old Agent runtime connections; updating the Web console alone is insufficient.

## 6. Ports and acceptance

| Location | Port | Purpose |
|---|---|---|
| Master | `443/tcp` | Web, API, and WSS RPC |
| Master | `9000/tcp` | Container Web/API; reverse proxy only |
| Master | `9001/tcp` | Native gRPC; do not publish when using WSS |
| Every Server | `7000/tcp` or configured port | FRPC ingress |
| Every Server | tunnel remote ports | Public TCP/UDP business ingress |
| Every Server | `8999/tcp` or custom `SERVER_API_PORT` | Loopback-only authentication API |

Verify that Master health is available at `/api/v2/health`, registration is closed, every Server and Client Agent is online, and every Tunnel shows the intended Client, Server, and ports. Back up the Master `frp-panel-data` Compose volume, every Server `frp-panel-server-data` Compose volume, and the Master `.env`. Docker usually prefixes the actual volume name with the Compose project name, so confirm it with `docker volume ls` or `docker volume inspect` before backup. Preserve the original `APP_GLOBAL_SECRET` during restore.

Review the [security baseline](/SECURITY), [configuration reference](/en/configuration), and [support matrix](/SUPPORT_MATRIX) before production use.
