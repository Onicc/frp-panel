# Controller configuration

Primary environment variables:

| Variable | Default | Purpose |
|---|---|---|
| `APP_GLOBAL_SECRET` | none | Required; at least 32 non-whitespace characters from a random source |
| `APP_ENABLE_REGISTER` | `false` | Enable temporarily only to create the initial Owner |
| `APP_COOKIE_SECURE` | `true` | Keep enabled in production |
| `APP_ALLOWED_ORIGINS` | empty | Comma-separated exact Origins when cross-origin WebSockets are required |
| `MASTER_API_HOST` | `MASTER_RPC_HOST` | Public Web/API hostname reachable by Servers, Agents, and browsers |
| `MASTER_API_SCHEME` | `http` | Set to `https` behind HTTPS termination |
| `MASTER_API_PORT` | `9000` | Web/API listen port |
| `MASTER_RPC_HOST` | `127.0.0.1` | Public RPC hostname reachable by Servers and Agents |
| `MASTER_RPC_PORT` | `9001` | Managed-component RPC listen port |
| `CLIENT_API_URL` | derived | Complete public API URL used in generated Server Compose files and Client installation commands |
| `CLIENT_RPC_URL` | derived | Complete public RPC URL used in generated Server Compose files and Client installation commands; `wss://panel.example.com` is recommended |
| `DB_TYPE` | `sqlite3` | `sqlite3` or `postgres` |
| `DB_DSN` | `/data/data.db?...` | SQLite file or PostgreSQL DSN |

`compose.yaml` requires both `CLIENT_*_URL` values explicitly so the console cannot generate deployments that point at an internal container address or the wrong port. The recommended same-origin HTTPS/WSS entry does not require publishing `MASTER_RPC_PORT`.

An independent FRPS container also accepts `SERVER_ENROLLMENT_TOKEN`, `SERVER_CONFIG_PATH` (default `/data/server.yaml`), `CLIENT_API_URL`, and `CLIENT_RPC_URL`. After enrollment, the volume configuration is authoritative and the one-use token may be removed.

Never commit `.env`, database files, enrollment tokens, or Agent configuration. Production ingress should enforce HTTPS, request-size limits, and trusted-proxy rules. See the complete [deployment guide](/en/deployment).
