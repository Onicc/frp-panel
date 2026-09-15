# 控制器配置

主要环境变量：

| 变量 | 默认值 | 说明 |
|---|---|---|
| `APP_GLOBAL_SECRET` | 无 | 必填，至少 32 个非空白字符；请使用随机值 |
| `APP_ENABLE_REGISTER` | `false` | 仅在创建首个 Owner 时临时开启 |
| `APP_COOKIE_SECURE` | `true` | 生产环境必须保持开启 |
| `APP_ALLOWED_ORIGINS` | 空 | 需要跨源 WebSocket 时使用逗号分隔的精确 Origin |
| `MASTER_API_HOST` | `MASTER_RPC_HOST` | Agent 可访问的公开 Web/API 主机名 |
| `MASTER_API_SCHEME` | `http` | HTTPS 反向代理场景设为 `https` |
| `MASTER_API_PORT` | `9000` | Web/API 监听端口 |
| `MASTER_RPC_HOST` | `127.0.0.1` | Agent 可访问的公开 RPC 主机名 |
| `MASTER_RPC_PORT` | `9001` | Agent RPC 监听端口 |
| `DB_TYPE` | `sqlite3` | `sqlite3` 或 `postgres` |
| `DB_DSN` | `/data/data.db?...` | SQLite 文件或 PostgreSQL DSN |

不要把 `.env`、数据库文件、注册令牌或 Agent 配置提交到版本库。生产入口应设置 HTTPS、请求大小限制和可信代理规则。
