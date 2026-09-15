# frp-panel v2

面向 FRP 的开源控制面，包含安全的 Web 控制台与跨平台节点 Agent。

> v2 是全新版本，不迁移 v1 数据库，也不保留旧客户端 CLI 的兼容层。

## 主要变化

- 控制器 `frp-panel` 与轻量节点程序 `frp-panel-agent` 分离发布。
- Agent 支持 Linux、macOS、Windows 的 amd64/arm64 主流平台。
- 安装命令写入系统规范目录，不再污染执行命令时的当前目录。
- 使用 Vite 8 / React 19 重写中英文控制台；创建、编辑成功后统一关闭并重置弹窗，失败时保留现场。
- 安全默认值：校验 TLS、关闭高权限功能、Argon2id 密码、按角色签发权限、同源 WebSocket、0600 Agent 配置。
- 发布 `onicc/frp-panel` 与 `onicc/frp-panel-agent` 两个多架构镜像。
- 发布物包含 SHA-256、SBOM、GitHub provenance，Actions 均固定到提交 SHA。

## 部署 Master、Server 与 Client

Master 提供控制面，默认 Server（FRPS）内置在同一个控制器容器中；Client 是安装到节点上的 `frp-panel-agent` 与受管 FRPC。先为 Docker Compose 创建 `.env`：

```bash
openssl rand -hex 32
```

```dotenv
APP_GLOBAL_SECRET=REPLACE_ME
APP_COOKIE_SECURE=true
APP_ENABLE_REGISTER=true
PUBLIC_HOST=panel.example.com
MASTER_API_SCHEME=https
CLIENT_API_URL=https://panel.example.com
CLIENT_RPC_URL=wss://panel.example.com
```

把 `REPLACE_ME` 替换为上一步生成的随机值。

将 HTTPS 反向代理指向 `127.0.0.1:9000`，然后启动：

```bash
docker compose config --quiet
docker compose pull
docker compose up -d
```

创建初始 Owner 后立即将 `APP_ENABLE_REGISTER=false` 并重新应用 Compose。内置 FRPS 默认使用 `7000`；各代理使用的 remote port 也必须显式映射。

Client 必须使用前端 **节点 → 添加节点** 中生成的 10 分钟一次性命令安装。控制台提供 Linux、macOS 和 Windows 命令，并自动带入公开 API/RPC 地址。

完整的反向代理、端口、备份、升级和 Client 验证步骤见 [部署指南](docs/deployment.md)。

## Client 安装位置

前端命令的安装位置如下：

| 系统 | 程序 | 配置与数据 |
|---|---|---|
| Linux | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/etc/frp-panel/agent.yaml`、`/var/lib/frp-panel` |
| macOS | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/Library/Application Support/frp-panel` |
| Windows | `%ProgramFiles%\frp-panel\frp-panel-agent.exe` | `%ProgramData%\frp-panel` |

系统服务名为 `frp-panel-agent`。程序不会自动加入 `PATH`；请使用上表中的完整路径或系统服务管理器维护它。

## 开发与测试

需要 Go 1.27.1 与 Node.js 24 LTS。

```bash
corepack enable
cd www && pnpm install --frozen-lockfile && pnpm build && cd ..
go test ./...
go build ./cmd/frpp ./cmd/frp-panel-agent
```

进一步阅读：[v2 架构](docs/ARCHITECTURE_V2.md)、[优化审查](docs/OPTIMIZATION.md)、[支持矩阵](docs/SUPPORT_MATRIX.md)、[安全审查](docs/SECURITY.md)、[OpenAPI](api/openapi.yaml)。

## 发布规则

- `main` 通过测试后更新 GitHub `edge` 发布及两个 Docker 镜像的 `edge` 标签。
- `v*` 标签生成稳定 GitHub Release，并发布对应版本与 `latest` 镜像。
- 发布物带校验和、SBOM 与 GitHub 构建证明；目前不提供 Apple 公证及 Windows Authenticode。

## 开源协议与署名

项目使用 AGPL-3.0，由 Onicc 维护，派生自 [VaalaCat/frp-panel](https://github.com/VaalaCat/frp-panel)。详见 [NOTICE](NOTICE)。
