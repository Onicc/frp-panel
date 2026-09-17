# frp-panel v2

面向 FRP 的开源控制面，包含安全的 Web 控制台与跨平台 Client Agent。

> v2 是全新版本，不迁移 v1 数据库，也不保留旧客户端 CLI 的兼容层。

## 主要变化

- Master 控制面、可独立部署的 Server（FRPS）数据面与轻量 Client Agent `frp-panel-agent` 相互分离。
- Agent 支持 Linux、macOS、Windows 的 amd64/arm64 主流平台。
- 安装命令写入系统规范目录，不再污染执行命令时的当前目录。
- 使用 Vue 3 / Vite 重写中英文控制台；只使用 sub2api 中的布局、表格、表单和对话框基础组件，成功后统一关闭并重置弹窗，失败时保留现场。
- 安全默认值：校验 TLS、关闭高权限功能、Argon2id 密码、按角色签发权限、同源 WebSocket、0600 Agent 配置。
- 发布 `onicc/frp-panel` 与 `onicc/frp-panel-agent` 两个多架构镜像。
- 发布物包含 SHA-256、SBOM、GitHub provenance，Actions 均固定到提交 SHA。

## 部署 Master、Server 与 Client

使用仓库 Compose 文件部署唯一的 Master。Master 只负责 Web 控制台和期望配置，不承载代理流量。在 **Servers** 中创建任意数量的 Server；每次创建都会生成一份带一次性令牌的 FRPS Docker Compose 部署。在 **Clients** 中创建 Client，并使用生成的 Linux、macOS 或 Windows 命令安装 Client Agent。创建 Tunnel 时直接选择 Client 与 Server；同一个 Client 可创建多条 Tunnel，每条可选择不同 Server，Client 在线且双方已完成注册时，底层 FRPC 连接由 Master 按需尽力建立和复用，连接回收受当前 Client 级协议限制。

Server 和 Client 的注册令牌都在 10 分钟后过期且只能兑换一次。长期凭据在 Master 中仅保存哈希，明文只写入受管主机的受保护文件或数据卷。

完整的 Master → Server → Client 顺序、原始 Compose 文件、环境变量、端口与备份检查见 [部署指南](docs/deployment.md)。

## Client 安装位置

前端命令的安装位置如下：

| 系统 | 程序 | 配置与数据 |
|---|---|---|
| Linux | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/etc/frp-panel/agent.yaml`、`/var/lib/frp-panel` |
| macOS | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/Library/Application Support/frp-panel` |
| Windows | `%ProgramFiles%\frp-panel\frp-panel-agent.exe` | `%ProgramData%\frp-panel` |

Linux 和 Windows 的系统服务名为 `frp-panel-agent`；macOS 的 launchd 标识为 `io.github.onicc.frp-panel.agent`。程序不会自动加入 `PATH`；请使用上表中的完整路径或系统服务管理器维护它。

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
