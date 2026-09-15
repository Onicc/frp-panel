# 快速开始

## 1. 启动控制器

```bash
git clone https://github.com/Onicc/frp-panel.git
cd frp-panel
export APP_GLOBAL_SECRET="$(openssl rand -hex 32)"
export PUBLIC_HOST="panel.example.com"
export APP_ENABLE_REGISTER=true
docker compose up -d
```

首次启动时临时开放注册，在 Web 控制台创建唯一的初始 Owner。创建成功后将 `APP_ENABLE_REGISTER=false`，再执行 `docker compose up -d`。后续用户应由 Owner 管理，公共注册接口不会继续创建账户。

端口：Web/API `9000`、Agent RPC `9001`、内置 FRPS `7000`。生产环境必须通过 HTTPS 暴露 Web/API，并保持 `APP_COOKIE_SECURE=true`。

## 2. 添加节点

登录后打开 **节点 → 添加节点**，填写名称并创建 10 分钟有效的一次性注册命令。选择 Linux、macOS 或 Windows标签页并在目标机器执行。令牌兑换一次即失效。

## 3. 检查状态

```bash
frp-panel-agent service status
frp-panel-agent doctor --json
```

更多信息见 [Agent 安装](/agent)、[配置](/configuration) 与 [安全基线](/SECURITY)。
