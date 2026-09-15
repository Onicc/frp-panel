# Deployment guide

## 1. Deployment model

| Role | Count | Responsibility | Default deployment |
|---|---:|---|---|
| Master | 1 | Web console, API, identity, configuration, and connection orchestration; it carries no proxy traffic | Docker Compose |
| Server (FRPS) | 1 or more | Public data-plane entry points for FRPC and tunnel traffic | Docker Compose on every Server host |
| Client (FRPC) | 1 or more | Agent and managed FRPC on a node that exposes services | Console-generated OS installation command |

One Master manages every Server and Client. A Client may be assigned one or more Server routes from the console; Master delivers a separate managed FRPC configuration for each route.

Deploy in this order: **Master → Server → Client → assign routes**.

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
      MASTER_API_HOST: ${PUBLIC_HOST:?set the public controller hostname}
      MASTER_API_SCHEME: ${MASTER_API_SCHEME:-https}
      MASTER_RPC_HOST: ${PUBLIC_HOST:?set the public controller hostname}
      CLIENT_API_URL: ${CLIENT_API_URL:?set the public API URL used by managed components}
      CLIENT_RPC_URL: ${CLIENT_RPC_URL:?set the public RPC URL used by managed components}
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

PUBLIC_HOST=panel.example.com
MASTER_API_SCHEME=https
CLIENT_API_URL=https://panel.example.com
CLIENT_RPC_URL=wss://panel.example.com
```

- Keep `APP_GLOBAL_SECRET` unchanged and private. It must contain at least 32 random bytes; changing or losing it invalidates existing credentials.
- Enable registration only while creating the first Owner. Set `APP_ENABLE_REGISTER=false` immediately afterward and reapply the Compose configuration.
- `PUBLIC_HOST` contains no scheme, port, or path.
- `CLIENT_API_URL` and `CLIENT_RPC_URL` are the public endpoints actually reachable by every Server and Client. They are embedded in console-generated deployments.
- With secure cookies enabled, terminate HTTPS at a reverse proxy that forwards Web, API, and WebSocket traffic to `127.0.0.1:9000`.
- Pin a tested `v*` image tag in production; `edge` follows `main`.

Master does not publish port `7000` or any tunnel remote port.

## 3. Deploy one or more Servers (FRPS)

Each FRPS is an independent data plane on the Linux host that receives public traffic:

1. Open **Servers → Create server** in Master.
2. Enter a unique ID, the public DNS name/IP used by FRPC, and the FRPS bind port (`7000` by default).
3. The successful dialog closes and the page displays a complete Server-specific `compose.yaml`.
4. Save it on the target Server host and apply it. The first start redeems a ten-minute one-use token and stores the permanent credential in `/data/server.yaml` inside the named volume.
5. Wait for the Server status to become online. Repeat for every additional FRPS host.

Never reuse one generated file or Server data volume on multiple machines. If the token expires before the volume is initialized, create the same ID again to replace that not-yet-enrolled record.

The generated file has this structure; the console replaces every placeholder:

```yaml
services:
  frps:
    image: onicc/frp-panel:edge
    restart: unless-stopped
    network_mode: host
    command:
      - server
      - --config
      - /data/server.yaml
      - --enrollment-token
      - "ONE_TIME_TOKEN_FROM_MASTER"
      - --api-url
      - "https://panel.example.com"
      - --rpc-url
      - "wss://panel.example.com"
    volumes:
      - frp-panel-server-data:/data

volumes:
  frp-panel-server-data:
```

Linux host networking lets FRPS accept new TCP/UDP remote ports without changing container port mappings. Restrict the host firewall and cloud security group to the FRPS bind port and the exact remote ports required by tunnels. The internal authentication API listens only on `127.0.0.1:8999` and must not be exposed.

For manual maintenance, this is the equivalent environment-based template.

### compose.yaml

```yaml
services:
  frps:
    image: ${FRP_PANEL_IMAGE:-onicc/frp-panel:edge}
    restart: unless-stopped
    network_mode: host
    command: ["server", "--config", "/data/server.yaml"]
    environment:
      SERVER_ENROLLMENT_TOKEN: ${SERVER_ENROLLMENT_TOKEN:-}
      CLIENT_API_URL: ${CLIENT_API_URL:?set the Master API URL}
      CLIENT_RPC_URL: ${CLIENT_RPC_URL:?set the Master RPC URL}
    volumes:
      - frp-panel-server-data:/data

volumes:
  frp-panel-server-data:
```

### .env

```dotenv
FRP_PANEL_IMAGE=onicc/frp-panel:edge
SERVER_ENROLLMENT_TOKEN=ONE_TIME_TOKEN_FROM_MASTER
CLIENT_API_URL=https://panel.example.com
CLIENT_RPC_URL=wss://panel.example.com
```

After the first successful enrollment, remove `SERVER_ENROLLMENT_TOKEN` from `.env`; restarts use the protected volume credential.

## 4. Install one or more Clients (FRPC)

Do not hand-write a Client Compose file or command:

1. Open **Nodes → Add node** in Master and enter a unique node ID.
2. Select Linux, macOS, or Windows and copy the complete generated command.
3. Run it with administrator privileges on the target node. It installs the binary, protected configuration, and native service in system locations, never in the command's current directory.
4. Wait for the node to become online. Generate a separate command for every Client.

| OS | Binary | Configuration and state |
|---|---|---|
| Linux | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/etc/frp-panel/agent.yaml`, `/var/lib/frp-panel` |
| macOS | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/Library/Application Support/frp-panel` |
| Windows | `%ProgramFiles%\frp-panel\frp-panel-agent.exe` | `%ProgramData%\frp-panel` |

See [Client / Agent operations](/en/agent) for service management, upgrade, and uninstall details.

## 5. Assign FRPC routes in Master

A newly installed Client does not connect to FRPS until Master assigns a route:

1. On **Nodes**, select a Server in the node's **FRPS routes** column and add it.
2. Add more Servers when one physical Agent must run multiple independent FRPC routes.
3. Remove an unused route in the same column. An online Agent applies changes immediately; changes made while offline apply no later than the next configuration sync.

The mapping is explicit per node and Server. Different Clients can use different FRPS hosts, one Client can use multiple FRPS hosts, and business traffic never passes through Master.

## 6. Ports and acceptance

| Location | Port | Purpose |
|---|---|---|
| Master | `443/tcp` | Web, API, and WSS RPC |
| Master | `9000/tcp` | Container Web/API; reverse proxy only |
| Master | `9001/tcp` | Native gRPC; do not publish when using WSS |
| Every Server | `7000/tcp` or configured port | FRPC ingress |
| Every Server | tunnel remote ports | Public TCP/UDP business ingress |
| Every Server | `8999/tcp` | Loopback-only authentication API |

Verify that Master health is available at `/api/v2/health`, registration is closed, every FRPS and Agent is online, and each node shows the intended FRPS routes. Back up the Master `frp-panel-data` volume, every Server `frp-panel-server-data` volume, and the Master `.env`. Preserve the original `APP_GLOBAL_SECRET` during restore.

Review the [security baseline](/SECURITY), [configuration reference](/en/configuration), and [support matrix](/SUPPORT_MATRIX) before production use.
