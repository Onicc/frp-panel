# Controller configuration

Primary environment variables:

| Variable | Default | Purpose |
|---|---|---|
| `APP_GLOBAL_SECRET` | none | Required; at least 32 non-whitespace characters from a random source |
| `APP_ENABLE_REGISTER` | `false` | Enable temporarily only to create the initial Owner |
| `APP_COOKIE_SECURE` | `true` | Keep enabled in production |
| `APP_ALLOWED_ORIGINS` | empty | Comma-separated exact Origins when cross-origin WebSockets are required |
| `MASTER_API_HOST` | `MASTER_RPC_HOST` | Public Web/API hostname reachable by Agents |
| `MASTER_API_SCHEME` | `http` | Set to `https` behind HTTPS termination |
| `MASTER_API_PORT` | `9000` | Web/API listen port |
| `MASTER_RPC_HOST` | `127.0.0.1` | Public RPC hostname reachable by Agents |
| `MASTER_RPC_PORT` | `9001` | Agent RPC listen port |
| `DB_TYPE` | `sqlite3` | `sqlite3` or `postgres` |
| `DB_DSN` | `/data/data.db?...` | SQLite file or PostgreSQL DSN |

Never commit `.env`, database files, enrollment tokens, or Agent configuration. Production ingress should enforce HTTPS, request-size limits, and trusted-proxy rules.
