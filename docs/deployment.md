# 部署指南

## 1. 先理解三个角色

| 角色 | 数量 | 职责 | 默认部署方式 |
|---|---:|---|---|
| Master | 1 | Web 控制台、API、用户、配置和连接调度；不承载代理流量 | Docker Compose |
| Server（FRPS） | 1 至多个 | 公网数据入口，接收 FRPC 连接和业务流量 | 每台服务器使用 Docker Compose |
| Client（FRPC） | 1 至多个 | 业务主机，运行 Client Agent 与 Master 管理的 FRPC | 使用控制台生成的系统安装命令 |

一个 Master 管理全部 Server、Client 和 Tunnel。Client 是物理业务主机，Agent 是安装在 Client 上的管理程序。每条 Tunnel 直接选择一个 Client 和一个 Server；Master 按需创建底层 FRPC 连接，同一 Client 与 Server 组合的多条 Tunnel 共用连接。

```text
                         ┌── Server A / FRPS ── 公网入口与 remote ports
浏览器 ── Master ────────┼── Server B / FRPS ── 公网入口与 remote ports
          控制面         └── Server ...
             │
             └──────────── Client A/B/... / Client Agent + FRPC
                              └── 每条 Tunnel 选择目标 Server
```

推荐按 **Master → Server → Client → Tunnel** 的顺序部署。

## 2. 部署唯一的 Master

为 Master 保留一个目录，其中只有 `compose.yaml` 和 `.env`。仓库根目录提供的 [compose.yaml](https://github.com/Onicc/frp-panel/blob/main/compose.yaml) 与下方内容一致。

### compose.yaml

```yaml
services:
  master:
    image: ${FRP_PANEL_IMAGE:-onicc/frp-panel:edge}
    restart: unless-stopped
    environment:
      APP_GLOBAL_SECRET: ${APP_GLOBAL_SECRET:?set a random 32+ character secret}
      APP_COOKIE_SECURE: ${APP_COOKIE_SECURE:-true}
      APP_ENABLE_REGISTER: ${APP_ENABLE_REGISTER:-false}
      APP_AGENT_INSTALL_URL: ${APP_AGENT_INSTALL_URL:-https://raw.githubusercontent.com/Onicc/frp-panel/main}
      PUBLIC_URL: ${PUBLIC_URL:?set PUBLIC_URL in .env}
    ports:
      - "127.0.0.1:9000:9000"
    volumes:
      - frp-panel-data:/data
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:9000/api/v2/health"]
      interval: 30s
      timeout: 5s
      retries: 3

volumes:
  frp-panel-data:
```

### .env

```dotenv
FRP_PANEL_IMAGE=onicc/frp-panel:edge

APP_GLOBAL_SECRET=REPLACE_WITH_A_RANDOM_32_BYTE_OR_LONGER_SECRET
APP_COOKIE_SECURE=true
APP_ENABLE_REGISTER=true
# Optional fork or internal mirror for install.sh/install.ps1.
# APP_AGENT_INSTALL_URL=https://raw.githubusercontent.com/Onicc/frp-panel/main

PUBLIC_URL=https://panel.example.com
```

配置要求：

- `APP_GLOBAL_SECRET` 必须是唯一的高强度随机值，至少 32 字节；需要长期保存，丢失或变更会使现有凭据失效。
- `APP_ENABLE_REGISTER` 只在创建首个 Owner 时设为 `true`，创建后立即改为 `false` 并重新应用 Compose 配置。
- `PUBLIC_URL` 是唯一需要配置的 Master 公开地址，必须是无路径的完整 `http://` 或 `https://` URL。HTTPS 会自动派生 `wss://` RPC 地址。
- `APP_AGENT_INSTALL_URL` 仅在使用 fork 或内部镜像时设置；它必须包含 `install.sh` 与 `install.ps1`。
- 使用 `APP_COOKIE_SECURE=true` 时，浏览器入口必须是 HTTPS；推荐让同机反向代理转发至 `127.0.0.1:9000`，并支持 WebSocket。
- `edge` 会随 `main` 更新；生产环境应固定到已验证的 `v*` 镜像标签。

Master 只需开放 Web/API/RPC 入口，不应映射 `7000` 或任何业务 remote port。

首次打开 `PUBLIC_URL` 时会自动显示“创建首个所有者账户”：填写用户名、邮箱、密码和确认密码后，页面会自动登录并进入控制台，浏览器也能按标准登录表单识别并保存凭据。确认可以重新登录后，将 `APP_ENABLE_REGISTER` 改为 `false` 并重新应用 Compose 配置。登录后可从右上角进入 **账户设置** 修改密码；系统会校验当前密码并在修改成功后要求重新登录。

## 3. 部署一个或多个 Server（FRPS）

每个 FRPS 都是独立数据面，部署在真正接收公网流量的 Linux 服务器上：

1. 登录 Master，打开 **Servers → 创建 Server**。
2. 填写唯一 ID、该机器供 FRPC 连接的公网域名/IP，以及 FRPS 绑定端口（默认 `7000`）。
3. 创建后，弹出页会关闭，页面上方会显示该 Server 专属的完整 `compose.yaml`。
4. 将文件保存到目标 Server 主机并应用。Server 首次启动会兑换一次性令牌，并把长期凭据写入 Docker 数据卷 `/data/server.yaml`。
5. 返回 Server 列表，状态变为“在线”即完成。重复以上步骤即可增加更多 FRPS。

令牌有效期为 10 分钟且只能使用一次。令牌超时或部署结果丢失时，删除尚未注册且没有依赖 Tunnel 的 Server 后，再用相同 ID 创建；已注册 Server 应使用“轮换凭据”。不要在多台机器上复用同一份生成文件或同一个数据卷。

控制台生成文件的结构如下；Master 会自动写入已配置的 `PUBLIC_URL` 和一次性令牌，用户不需要再次配置域名：

```yaml
services:
  frps:
    image: onicc/frp-panel:edge
    restart: unless-stopped
    network_mode: host
    command:
      - server
      - --config
      - /data/server.yaml
    environment:
      PUBLIC_URL: "MASTER_PUBLIC_URL_INSERTED_BY_CONSOLE"
      SERVER_ENROLLMENT_TOKEN: "ONE_TIME_TOKEN_FROM_MASTER"
    volumes:
      - frp-panel-server-data:/data

volumes:
  frp-panel-server-data:
```

这里使用 Linux 的 host network，使 FRPS 后续新增任意 TCP/UDP remote port 时不必反复修改容器端口映射。必须在 Server 主机防火墙和云安全组中放行：

- FRPS 绑定端口，例如 `7000/tcp`，供各 Client 连接；
- 每条 Tunnel 实际使用的 TCP/UDP remote port，供业务访问；
- Server 到 Master 的 HTTPS/WSS 出站访问。

容器内部的 `127.0.0.1:8999` 仅供 FRPS 鉴权插件使用，不应对外开放。

Server 必须使用控制台生成的文件，不要手工复制 Master 的 `.env`。首次注册成功后，重启会直接读取数据卷中的受保护凭据。

## 4. 安装一个或多个 Client（FRPC）

Client 不使用手写 Compose 或手工拼接参数：

1. 在 Master 打开 **Clients → 添加 Client**，输入唯一 Client ID。
2. 选择 Linux、macOS 或 Windows，复制控制台生成的完整安装命令。
3. 在目标 Client 主机以管理员权限执行。脚本会安装 Client Agent、写入受保护配置并注册系统服务，不会使用当前命令目录作为安装位置。
4. 返回 Client 列表，Client Agent 状态变为“在线”即完成。每个 Client 都必须使用自己生成的命令。

默认安装位置：

| 系统 | 二进制 | 配置与数据 |
|---|---|---|
| Linux | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/etc/frp-panel/agent.yaml`、`/var/lib/frp-panel` |
| macOS | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/Library/Application Support/frp-panel` |
| Windows | `%ProgramFiles%\frp-panel\frp-panel-agent.exe` | `%ProgramData%\frp-panel` |

注册令牌同样只在 10 分钟内有效并只能兑换一次。详细的服务管理、升级和卸载说明见 [Client / Agent 安装与维护](/agent)。

## 5. 创建 Tunnel 并选择 Server

Client 上线后无需预先绑定 Server。打开 **Tunnels → 创建 Tunnel**：

1. 选择承载本地服务的 Client。
2. 选择提供公网入口的 Server。
3. 填写 TCP/UDP、本地地址、本地端口与公网端口并确认。

一台 Client 可以创建多条 Tunnel，每条 Tunnel 可以选择不同 Server。Master 会自动为首次使用的“Client + Server”组合创建 FRPC 连接，同一组合的 Tunnel 共用连接；删除该组合的最后一条 Tunnel 时连接自动回收。Client 与 Server 的底层连接不作为用户资源单独管理。

公网端口监听在所选 Server（FRPS）上，并转发至所选 Client 能够访问的本地服务；业务流量不会经过 Master。Client 离线时配置仍会保存，并在 Agent 恢复连接后同步。

## 6. 网络端口

| 位置 | 端口 | 用途 | 建议 |
|---|---|---|---|
| Master | `443/tcp` | Web、API、WSS RPC | 对浏览器、Server、Client 开放 |
| Master | `9000/tcp` | 容器内 Web/API | 仅反向代理访问 |
| Master | `9001/tcp` | 原生 gRPC | 使用 WSS 时不发布 |
| 每台 Server | `7000/tcp` 或自定义值 | FRPC 连接入口 | 对 Client 开放 |
| 每台 Server | remote ports | Tunnel 业务入口 | 按业务逐项开放 TCP/UDP |
| 每台 Server | `8999/tcp` | 本地鉴权 API | 仅回环地址，不开放 |

## 7. 验收与备份

- Master 健康检查地址 `/api/v2/health` 可访问，且公开注册已经关闭。
- 所有 FRPS 在 Server 列表显示在线；Server 主机只开放计划内的绑定端口和 remote ports。
- 所有 Client Agent 在 Client 列表显示在线；每条 Tunnel 显示正确的 Client、Server 与端口。
- 已分别备份 Master 的 `frp-panel-data` 卷、每台 Server 的 `frp-panel-server-data` 卷以及 Master 的 `.env`。
- 恢复时保持原 `APP_GLOBAL_SECRET` 和数据卷；不要让两台 FRPS 同时使用同一个 Server 数据卷副本。

上线前同时检查 [安全基线](/SECURITY)、[配置说明](/configuration) 和 [平台支持矩阵](/SUPPORT_MATRIX)。
