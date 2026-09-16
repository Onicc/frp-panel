# Master configuration

Primary environment variables:

| Variable | Default | Purpose |
|---|---|---|
| `APP_GLOBAL_SECRET` | none | Required; at least 32 non-whitespace characters from a random source |
| `APP_ENABLE_REGISTER` | `false` | Enable temporarily only to create the initial Owner |
| `APP_COOKIE_SECURE` | `true` | Keep enabled in production |
| `APP_ALLOWED_ORIGINS` | empty | Comma-separated exact Origins when cross-origin WebSockets are required |
| `PUBLIC_URL` | none | Required single public endpoint such as `https://panel.example.com`; derives the HTTPS/WSS addresses and hostnames |
| `MASTER_API_PORT` | `9000` | Web/API listen port |
| `MASTER_RPC_PORT` | `9001` | Managed-component RPC listen port |
| `DB_TYPE` | `sqlite3` | `sqlite3` or `postgres` |
| `DB_DSN` | `/data/data.db?...` | SQLite file or PostgreSQL DSN |

`PUBLIC_URL` cannot contain credentials, a path, query, or fragment. `https://` derives `wss://`, while `http://` derives `ws://`; same-origin WSS does not require publishing `MASTER_RPC_PORT`. Low-level `MASTER_*` and `CLIENT_*` address variables remain for unusual network topologies and should not be mixed with `PUBLIC_URL`.

An independent FRPS container accepts the same `PUBLIC_URL`, plus `SERVER_ENROLLMENT_TOKEN` and `SERVER_CONFIG_PATH` (default `/data/server.yaml`). After enrollment, the volume configuration is authoritative and the one-use token may be removed.

Never commit `.env`, database files, enrollment tokens, or Agent configuration. Production ingress should enforce HTTPS, request-size limits, and trusted-proxy rules. See the complete [deployment guide](/en/deployment).
