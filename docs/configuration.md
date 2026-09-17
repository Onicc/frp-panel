# Master 配置

主要环境变量。下表描述程序支持的变量；如果使用仓库提供的 `compose.yaml`，只有该文件 `environment` 段中列出的变量会从宿主机 `.env` 传入容器。`APP_ALLOWED_ORIGINS`、`MASTER_*` 和 `DB_*` 需要显式加入 Compose，不能只写在 `.env` 中。

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
| `DB_DSN` | 程序默认 `/data/data.db?...`；官方 Master 镜像默认 `/data/frp-panel.db?...` | SQLite 文件或 PostgreSQL DSN |

`PUBLIC_URL` 不允许包含账号、路径、查询参数或片段。`https://` 自动派生 `wss://`，`http://` 自动派生 `ws://`；使用同源 WSS 时不需要对外发布 `MASTER_RPC_PORT`。底层 `MASTER_*` 和 `CLIENT_*` 地址变量仅保留给特殊网络拓扑，不应与 `PUBLIC_URL` 混用。

官方 Master 镜像通过镜像环境变量将 SQLite 文件设为 `/data/frp-panel.db`。如果修改 `MASTER_API_PORT`、`MASTER_RPC_PORT` 或数据库配置，除了传入环境变量，还必须同步修改 Compose 的端口映射、健康检查和数据卷策略；仓库默认 Compose 只映射 `127.0.0.1:9000`。

独立 FRPS 容器接受相同的 `PUBLIC_URL`，以及 `SERVER_ENROLLMENT_TOKEN`、`SERVER_CONFIG_PATH`（默认 `/data/server.yaml`）。注册成功后以数据卷中的配置为准，一次性令牌可以移除。

Server Compose 还可以设置 `SERVER_API_PORT`（默认 `8999`），用于 FRPS 鉴权插件访问的本机回环 API。它必须与 Server 的 FRPS 绑定端口和所有 Tunnel 远端端口不同，并在目标主机上保持空闲；该端口不应暴露到公网。

不要把 `.env`、数据库文件、注册令牌或 Agent 配置提交到版本库。生产入口应设置 HTTPS、请求大小限制和可信代理规则。完整示例见 [部署指南](/deployment)。
