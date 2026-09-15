# 部署指南

## 组件说明

| 组件 | 作用 | 部署方式 |
|---|---|---|
| Master | Web 控制台、API、用户和配置管理、节点控制连接 | Docker Compose |
| Server | FRPS 数据入口，将公网流量转发到 Client | 内置在 Master 容器中，ID 为 `default` |
| Client | 节点上的 `frp-panel-agent` 和受管 FRPC | 使用前端生成的安装命令 |

控制台将 Client 显示为“节点（Node）”。当前版本不提供独立的 Server 运行程序；启动 Master 时会同时启动内置 Server，不需要部署第二个容器。

## 部署文件

建议为部署保留一个独立目录：

```text
frp-panel/
├── compose.yaml
└── .env
```

### compose.yaml

以下是可直接保存的完整文件。TCP/UDP 代理使用的 remote port 需要额外添加到 `ports`。

```yaml
services:
  controller:
    image: ${FRP_PANEL_IMAGE:-onicc/frp-panel:edge}
    restart: unless-stopped
    environment:
      APP_GLOBAL_SECRET: ${APP_GLOBAL_SECRET:?set a random 32+ character secret}
      APP_COOKIE_SECURE: ${APP_COOKIE_SECURE:-true}
      APP_ENABLE_REGISTER: ${APP_ENABLE_REGISTER:-false}
      MASTER_API_HOST: ${PUBLIC_HOST:?set the public controller hostname}
      MASTER_API_SCHEME: ${MASTER_API_SCHEME:-https}
      MASTER_RPC_HOST: ${PUBLIC_HOST:?set the public controller hostname}
      CLIENT_API_URL: ${CLIENT_API_URL:?set the public API URL used by Agents}
      CLIENT_RPC_URL: ${CLIENT_RPC_URL:?set the public RPC URL used by Agents}
    ports:
      - "127.0.0.1:9000:9000"
      - "7000:7000"
      # Publish every TCP/UDP remote port used by a proxy, for example:
      # - "10000-10100:10000-10100/tcp"
      # - "10000-10100:10000-10100/udp"
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

生产环境示例：

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

配置要点：

- `APP_GLOBAL_SECRET` 必须替换为至少 32 字节的随机值，并长期安全保存。
- `APP_ENABLE_REGISTER` 仅在创建首个 Owner 时设为 `true`；创建完成后改为 `false` 并重新应用 Compose 配置。
- `PUBLIC_HOST` 只填写域名，不包含协议或路径。
- `CLIENT_API_URL` 和 `CLIENT_RPC_URL` 会直接进入前端生成的 Client 安装命令，必须是节点实际可访问的公网地址。
- 生产环境推荐使用同域名的 `https://` 与 `wss://`，不要使用 `localhost`、容器名、`http://` 或 `ws://`。
- `edge` 随 `main` 更新；稳定环境建议固定到经过验证的 `v*` 镜像标签。
- `.env` 包含密钥，文件权限应限制为仅部署管理员可读。

更多可选项见 [配置说明](/configuration)。

## 网络与反向代理

推荐由同机 HTTPS 反向代理接收公网请求，并转发到 `127.0.0.1:9000`。该入口同时承载 Web、API 和 `wss` Agent RPC。

| 端口 | 用途 | 暴露建议 |
|---|---|---|
| `443/tcp` | HTTPS、WebSocket RPC | 对外开放 |
| `9000/tcp` | Master Web/API | 仅供反向代理访问 |
| `7000/tcp` | 内置 FRPS | 使用默认 Server 时对 Client 开放 |
| remote port | TCP/UDP 代理流量 | 按业务需要在 Compose 和防火墙中同时开放 |
| `9001/tcp` | 原生 gRPC RPC | 使用推荐的 WSS 入口时无需发布 |
| `8999/tcp` | 内置 Server API | 不对外发布 |

面板域名需要正确解析到反向代理，并且 Master 容器自身也应能够访问 `CLIENT_API_URL` 和 `CLIENT_RPC_URL`。

如果反向代理不在 Docker 主机上，需要调整 `compose.yaml` 中 `127.0.0.1:9000:9000` 的绑定地址，并通过防火墙限制来源。

## Server 配置

Master 启动后会自动创建并运行 `default` Server。登录控制台后，可在 **Server** 页面确认其在线状态和配置。

Docker 只转发 `compose.yaml` 中明确发布的端口。创建 TCP 或 UDP 代理前，必须将对应 remote port 添加到 `ports`；否则控制台配置虽然存在，公网流量仍无法进入容器。

当前版本不支持部署独立的远程 Server。在 Server 页面新增记录不会自动在另一台机器上安装或启动 FRPS。

## Client 安装

Client 必须使用控制台生成的命令安装：

1. 登录 Master，打开 **节点（Nodes）→ 添加节点**。
2. 输入节点 ID 并创建注册信息。
3. 选择目标系统：Linux、macOS 或 Windows。
4. 复制页面显示的完整安装命令，在目标节点上以管理员权限执行。
5. 返回节点列表，确认节点在线。

注册令牌仅在 10 分钟内有效，成功使用一次后立即失效。不要保存或分享包含令牌的安装命令。

如果命令中的域名或协议不正确，应先修正 `.env` 中的 `CLIENT_API_URL`、`CLIENT_RPC_URL`，重新应用 Master 配置，然后生成新的安装命令。

安装路径、服务管理、升级和卸载说明见 [Client / Agent 安装与维护](/agent)。

## 部署验收

- `https://panel.example.com/api/v2/health` 可正常访问。
- 首个 Owner 创建后，公开注册已关闭。
- 控制台中的 `default` Server 显示在线。
- Client 使用前端命令安装后显示在线。
- FRPS `7000` 和实际使用的 remote port 已在 Compose、防火墙和云安全组中放行。
- 已备份 `.env` 和 `frp-panel-data` 数据卷；升级或移除服务时不删除该数据卷。

上线前请同时检查 [安全基线](/SECURITY) 和 [平台支持矩阵](/SUPPORT_MATRIX)。
