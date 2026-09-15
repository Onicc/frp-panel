# 快速开始

Master 提供控制面，Server 是 Master 内置的默认 FRPS，Client 则是安装在 Linux、macOS 或 Windows 节点上的 Agent 与受管 FRPC。当前版本使用一个 Docker Compose 服务同时部署 Master 和 Server，再从前端生成 Client 安装命令。

## 1. 配置 Master 与 Server

复制仓库的 `compose.yaml`，生成至少 32 字节的随机密钥，并创建 `.env`：

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

首次启动时临时开放注册，在 Web 控制台创建唯一的初始 Owner。创建成功后将 `APP_ENABLE_REGISTER=false`，再执行 `docker compose up -d`。内置 FRPS 使用 `7000`；每个代理使用的 remote port 也必须在 Compose 中映射。

## 2. 安装 Client

登录后打开 **节点 → 添加节点**，填写名称并创建 10 分钟有效的一次性注册命令。选择 Linux、macOS 或 Windows 标签页并在目标机器执行。令牌兑换一次即失效。

## 3. 检查状态

在控制台确认 `default` Server 与新节点在线，再使用对应系统的服务管理器检查 `frp-panel-agent`。完整命令见 [Client / Agent 安装与维护](/agent)。

生产部署、反向代理、端口、备份与升级步骤见 [部署指南](/deployment)。上线前同时阅读 [配置](/configuration) 与 [安全基线](/SECURITY)。
