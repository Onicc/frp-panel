# 控制器配置

主要环境变量：

| 变量 | 默认值 | 说明 |
|---|---|---|
| `APP_GLOBAL_SECRET` | 无 | 必填，至少 32 个非空白字符；请使用随机值 |
| `APP_ENABLE_REGISTER` | `false` | 仅在创建首个 Owner 时临时开启 |
| `APP_COOKIE_SECURE` | `true` | 生产环境必须保持开启 |
| `APP_ALLOWED_ORIGINS` | 空 | 需要跨源 WebSocket 时使用逗号分隔的精确 Origin |
| `MASTER_API_HOST` | `MASTER_RPC_HOST` | Server、Agent 和浏览器可访问的公开 Web/API 主机名 |
| `MASTER_API_SCHEME` | `http` | HTTPS 反向代理场景设为 `https` |
| `MASTER_API_PORT` | `9000` | Web/API 监听端口 |
| `MASTER_RPC_HOST` | `127.0.0.1` | Server 和 Agent 可访问的公开 RPC 主机名 |
| `MASTER_RPC_PORT` | `9001` | 受管组件 RPC 监听端口 |
| `CLIENT_API_URL` | 自动组合 | 前端生成 Server Compose 和 Client 安装命令时使用的完整公开 API URL |
| `CLIENT_RPC_URL` | 自动组合 | 前端生成 Server Compose 和 Client 安装命令时使用的完整公开 RPC URL，推荐 `wss://panel.example.com` |
| `DB_TYPE` | `sqlite3` | `sqlite3` 或 `postgres` |
| `DB_DSN` | `/data/data.db?...` | SQLite 文件或 PostgreSQL DSN |

`compose.yaml` 要求显式填写两个 `CLIENT_*_URL`，防止前端生成指向容器内部地址或错误端口的部署内容。使用推荐的同源 HTTPS/WSS 入口时，不需要对外发布 `MASTER_RPC_PORT`。

独立 FRPS 容器还接受 `SERVER_ENROLLMENT_TOKEN`、`SERVER_CONFIG_PATH`（默认 `/data/server.yaml`）、`CLIENT_API_URL` 和 `CLIENT_RPC_URL`。注册成功后以数据卷中的配置为准，一次性令牌可以移除。

不要把 `.env`、数据库文件、注册令牌或 Agent 配置提交到版本库。生产入口应设置 HTTPS、请求大小限制和可信代理规则。完整示例见 [部署指南](/deployment)。
