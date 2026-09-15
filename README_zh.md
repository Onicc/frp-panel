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

## 运行控制器

先生成随机密钥：

```bash
export APP_GLOBAL_SECRET="$(openssl rand -hex 32)"
export PUBLIC_HOST="panel.example.com"
export APP_ENABLE_REGISTER=true
docker compose up -d
```

在 Web 控制台创建初始 Owner 后，将 `APP_ENABLE_REGISTER=false` 并再次执行 `docker compose up -d`。Web/API 端口为 `9000`，Agent RPC 为 `9001`，内置默认 FRPS 为 `7000`。生产环境应在 Web 入口前配置 HTTPS，并保持 `APP_COOKIE_SECURE=true`。

## 安装 Agent

在 **节点 → 添加节点** 中生成 10 分钟有效的注册命令，选择对应系统后复制。安装位置如下：

| 系统 | 程序 | 配置与数据 |
|---|---|---|
| Linux | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/etc/frp-panel/agent.yaml`、`/var/lib/frp-panel` |
| macOS | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/Library/Application Support/frp-panel` |
| Windows | `%ProgramFiles%\frp-panel\frp-panel-agent.exe` | `%ProgramData%\frp-panel` |

系统服务名为 `frp-panel-agent`。可使用 `service status/restart/uninstall --purge` 管理服务，使用 `doctor --json` 检查能力与配置。

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
