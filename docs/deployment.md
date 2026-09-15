# 部署指南

## 先理解 Master、Server 与 Client

| 组件 | 职责 | 当前交付方式 | 运行位置 |
|---|---|---|---|
| Master | 提供 Web 控制台、HTTP API、用户与配置管理，并接收节点的控制连接 | `onicc/frp-panel` 镜像 | Linux Docker 主机 |
| Server | 运行 FRPS，接收公网流量并将其转发到 Client | 当前版本内置在 Master 进程中，Server ID 为 `default` | 与 Master 相同的容器 |
| Client | 在目标节点上运行 `frp-panel-agent` 与受管 FRPC，接收 Master 下发的配置 | 前端生成的一次性安装命令 | Linux、macOS 或 Windows 节点 |

这里的 **Client** 在控制台中显示为“节点（Node）”。Agent 是系统服务，FRPC 是它管理的数据面进程；不要再使用旧版 `frp-panel client` 命令。

当前 v2 发布物没有可独立启动的 `server` 子命令，也没有远程 Server 安装器。因此，受支持的 Server 部署方式是使用 Docker Compose 启动 Master 时一并启动内置的 `default` Server。不要复制启动第二个控制器容器来冒充独立 Server，也不要只在“Server”页面创建一条记录后就认为远端 FRPS 已经部署。

## 推荐拓扑与端口

推荐让 Web/API 和 Agent RPC 都通过同一个 HTTPS 域名访问：

| 外部入口 | 到达位置 | 用途 | 是否必须 |
|---|---|---|---|
| TCP `443` | 反向代理 → `127.0.0.1:9000` | Web、API、`wss` Agent RPC | 生产环境必须 |
| TCP `80` | 反向代理 | ACME 证书签发或跳转 HTTPS | 取决于证书方案 |
| TCP `7000` | Controller 容器 | 内置 FRPS 的默认绑定端口 | 使用默认 Server 时必须 |
| 代理配置的 remote port | Controller 容器 | 实际的 TCP/UDP 隧道流量 | 按需开放并映射 |
| TCP `9001` | Controller 容器 | 原生 gRPC RPC | 使用推荐的 `wss` 入口时无需发布 |
| TCP `8999` | 仅容器内部 | 内置 Server API | 禁止对外发布 |

DNS 中的面板域名必须指向反向代理，并且该公开地址也应能被 Controller 容器自身访问。这样内置 Server 和外部 Client 都能使用同一组受 TLS 保护的地址。

## 使用 Docker Compose 部署 Master 与 Server

### 1. 准备条件

- 一台安装了 Docker Engine 与 Docker Compose 插件的 Linux 主机。
- 一个已经解析到该主机的域名，例如 `panel.example.com`。
- 一个受信任的 HTTPS 反向代理；下文以 Caddy 为例。
- 防火墙只开放上表列出的必要端口。

在主机上创建独立的部署目录并取得仓库中的 Compose 文件：

```bash
sudo install -d -m 0750 -o "$(id -u)" -g "$(id -g)" /opt/frp-panel
cd /opt/frp-panel
curl -fsSLo compose.yaml \
  https://raw.githubusercontent.com/Onicc/frp-panel/main/compose.yaml
```

### 2. 创建环境文件

先生成密钥：

```bash
openssl rand -hex 32
```

将输出填入 `/opt/frp-panel/.env`。该文件包含密钥，权限应为 `0600`：

```dotenv
FRP_PANEL_IMAGE=onicc/frp-panel:edge
APP_GLOBAL_SECRET=REPLACE_ME
APP_COOKIE_SECURE=true
APP_ENABLE_REGISTER=true

PUBLIC_HOST=panel.example.com
MASTER_API_SCHEME=https
CLIENT_API_URL=https://panel.example.com
CLIENT_RPC_URL=wss://panel.example.com
```

`CLIENT_API_URL` 和 `CLIENT_RPC_URL` 会直接写入前端生成的 Client 安装命令。它们必须是 Client 真正可访问的公网地址；生产环境不要使用 `http://`、`ws://`、`localhost` 或 Docker 服务名。

`REPLACE_ME` 必须替换为上一步的随机输出；该短占位值会被程序拒绝，不能直接启动。

```bash
chmod 0600 /opt/frp-panel/.env
```

`edge` 是随 `main` 更新的滚动版本。正式环境有经过验证的 `v*` 版本后，应把 `FRP_PANEL_IMAGE` 固定到该版本；升级前先备份数据。

### 3. 配置 HTTPS 反向代理

Caddy 的最小站点配置如下。Caddy 会自动代理 WebSocket：

```text
panel.example.com {
    encode zstd gzip
    reverse_proxy 127.0.0.1:9000
}
```

先启动或重新加载反向代理，再启动 frp-panel。若反向代理不在同一台主机，需要相应调整 `compose.yaml` 中 `127.0.0.1:9000:9000` 的绑定地址，并用防火墙限制只有反向代理能访问它。

### 4. 启动并创建 Owner

```bash
cd /opt/frp-panel
sudo docker compose config --quiet
sudo docker compose pull
sudo docker compose up -d
sudo docker compose ps
sudo docker compose logs --tail=100 controller
```

打开 `https://panel.example.com` 并创建唯一的初始 Owner。完成后立即把 `.env` 改为：

```dotenv
APP_ENABLE_REGISTER=false
```

然后重新应用配置：

```bash
sudo docker compose up -d
```

健康检查：

```bash
curl -fsS https://panel.example.com/api/v2/health
```

登录后打开 **Server** 页面，`default` 应为在线状态。它就是和 Master 同容器运行的内置 FRPS。

### 5. 为 FRP 代理映射端口

Docker 只会转发 `compose.yaml` 中显式发布的端口。创建 TCP/UDP 代理前，把它使用的 remote port 加入 `ports`，例如：

```yaml
ports:
  - "127.0.0.1:9000:9000"
  - "7000:7000/tcp"
  - "10000-10100:10000-10100/tcp"
  - "10000-10100:10000-10100/udp"
```

修改后执行 `docker compose up -d`。只映射业务确实需要的端口，并同步限制云安全组和主机防火墙。

### 6. 备份与升级

SQLite 数据位于 Compose 命名卷 `frp-panel-data` 的 `/data` 中。备份时先停止写入，再复制整个目录，同时单独安全保存 `.env`：

```bash
cd /opt/frp-panel
sudo docker compose stop controller
sudo mkdir -p backup
sudo docker compose cp controller:/data backup/
sudo docker compose start controller
```

升级时修改 `FRP_PANEL_IMAGE`，然后执行：

```bash
sudo docker compose pull
sudo docker compose up -d
```

不要执行 `docker compose down -v`，它会删除数据库卷。`APP_GLOBAL_SECRET` 也不应在常规升级中更换。

## 从前端安装 Client

不要手写包含节点密钥的命令，也不要直接运行仓库中的安装脚本而省略注册参数。正确流程是：

1. 登录 Master，打开 **节点（Nodes）→ 添加节点**。
2. 输入只包含字母、数字、下划线或连字符的节点 ID 并确认。
3. 选择目标系统的 **Linux**、**macOS** 或 **Windows** 标签页。
4. 复制页面生成的完整命令，在目标节点上以管理员权限执行。
5. 回到节点列表，确认节点显示为在线。

注册令牌仅在 10 分钟内有效，而且成功兑换一次后立即失效。命令中含有该令牌，不要把它保存到脚本、Shell 历史共享记录、工单或聊天消息中。若命令显示错误的域名或出现生产环境不应使用的 `http://` / `ws://`，请先修正 Master 的 `CLIENT_API_URL` 与 `CLIENT_RPC_URL` 并重启容器，再生成新命令。

安装器会下载对应架构的发布物、校验 `checksums.txt`，然后安装到系统目录并注册服务，不会在当前工作目录留下文件。详细路径、维护、升级和卸载方法见 [Client / Agent 安装与维护](/agent)。

安装后可使用系统服务管理器检查状态：

::: code-group

```bash [Linux]
sudo systemctl status frp-panel-agent
sudo /usr/local/libexec/frp-panel/frp-panel-agent doctor --json
```

```bash [macOS]
sudo launchctl print system/io.github.onicc.frp-panel.agent
sudo /usr/local/libexec/frp-panel/frp-panel-agent doctor --json
```

```powershell [Windows PowerShell]
Get-Service frp-panel-agent
& "$env:ProgramFiles\frp-panel\frp-panel-agent.exe" doctor --json
```

:::

## 仅限本机评估的无 TLS 配置

只在同一台机器上临时评估时，可以不用反向代理：

```dotenv
APP_COOKIE_SECURE=false
PUBLIC_HOST=127.0.0.1
MASTER_API_SCHEME=http
CLIENT_API_URL=http://127.0.0.1:9000
CLIENT_RPC_URL=ws://127.0.0.1:9000
```

浏览器访问 `http://127.0.0.1:9000`。这组配置没有传输加密，不能用于公网、局域网共享或生产环境，也不能用于安装到其他机器上的 Client。

进一步的环境变量说明见 [配置](/configuration)，上线前检查 [安全基线](/SECURITY) 与 [平台支持矩阵](/SUPPORT_MATRIX)。
