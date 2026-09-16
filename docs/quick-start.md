# 快速开始

frp-panel 采用 **一个 Master、多个独立 FRPS Server、多个 FRPC Client** 的结构。Master 只负责控制面，业务流量直接经过所选 FRPS。

## 1. Master

使用仓库 `compose.yaml` 部署唯一的 Master，并按 `.env.sample` 设置 `APP_GLOBAL_SECRET`、公网 HTTPS API 与 WSS RPC 地址。首次创建 Owner 时临时启用注册，创建完成后立即关闭。

## 2. Server（FRPS）

登录后打开 **Servers → 创建 Server**，填写目标机器的公网地址和绑定端口。保存后控制台会显示完整 `compose.yaml`；将它部署到对应 Linux 服务器。每个 FRPS 主机都应单独创建，不共享生成文件或数据卷。

## 3. Client（FRPC）

打开 **Clients → 添加 Client**，选择 Linux、macOS 或 Windows，并在目标 Client 主机执行控制台生成的一次性安装命令。安装程序会把 Client Agent 写入系统规范目录并注册本机服务。

## 4. 创建 Tunnel

打开 **Tunnels → 创建 Tunnel**，直接选择 Client 与 Server，再设置本地服务和公网端口。每个 Client 可创建多条 Tunnel，每条可选择不同 Server；底层 FRPC 连接由 Master 自动创建、复用和回收。

完整的原始 Compose、环境变量、端口、验收和备份说明见 [部署指南](/deployment)。
