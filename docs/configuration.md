# Master 配置

主要环境变量：

| 变量 | 默认值 | 说明 |
|---|---|---|
| `APP_GLOBAL_SECRET` | 无 | 必填，至少 32 个非空白字符；请使用随机值 |
| `APP_ENABLE_REGISTER` | `false` | 仅在创建首个 Owner 时临时开启 |
| `APP_COOKIE_SECURE` | `true` | 生产环境必须保持开启 |
| `APP_ALLOWED_ORIGINS` | 空 | 需要跨源 WebSocket 时使用逗号分隔的精确 Origin |
| `PUBLIC_URL` | 无 | 必填的唯一公开入口，例如 `https://panel.example.com`；自动生成 HTTPS/WSS 地址与主机名 |
| `MASTER_API_PORT` | `9000` | Web/API 监听端口 |
| `MASTER_RPC_PORT` | `9001` | 受管组件 RPC 监听端口 |
| `DB_TYPE` | `sqlite3` | `sqlite3` 或 `postgres` |
| `DB_DSN` | `/data/data.db?...` | SQLite 文件或 PostgreSQL DSN |

`PUBLIC_URL` 不允许包含账号、路径、查询参数或片段。`https://` 自动派生 `wss://`，`http://` 自动派生 `ws://`；使用同源 WSS 时不需要对外发布 `MASTER_RPC_PORT`。底层 `MASTER_*` 和 `CLIENT_*` 地址变量仅保留给特殊网络拓扑，不应与 `PUBLIC_URL` 混用。

独立 FRPS 容器接受相同的 `PUBLIC_URL`，以及 `SERVER_ENROLLMENT_TOKEN`、`SERVER_CONFIG_PATH`（默认 `/data/server.yaml`）。注册成功后以数据卷中的配置为准，一次性令牌可以移除。

不要把 `.env`、数据库文件、注册令牌或 Agent 配置提交到版本库。生产入口应设置 HTTPS、请求大小限制和可信代理规则。完整示例见 [部署指南](/deployment)。
