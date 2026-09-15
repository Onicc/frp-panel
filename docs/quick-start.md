# 快速开始

frp-panel 采用 **一个 Master、多个独立 FRPS Server、多个 FRPC Client** 的结构。Master 只负责控制面，业务流量直接经过所选 FRPS。

## 1. Master

使用仓库 `compose.yaml` 部署唯一的 Master，并按 `.env.sample` 设置 `APP_GLOBAL_SECRET`、公网 HTTPS API 与 WSS RPC 地址。首次创建 Owner 时临时启用注册，创建完成后立即关闭。

## 2. Server（FRPS）

登录后打开 **服务端 → 创建服务端**，填写目标机器的公网地址和绑定端口。保存后控制台会显示完整 `compose.yaml`；将它部署到对应 Linux 服务器。每个 FRPS 主机都应单独创建，不共享生成文件或数据卷。

## 3. Client（FRPC）

打开 **节点 → 添加节点**，选择 Linux、macOS 或 Windows，并在目标节点执行控制台生成的一次性安装命令。安装程序会写入系统规范目录并注册本机服务。

## 4. 分配链路

在 **节点** 页为每个 Client 添加需要连接的 FRPS。可以让不同 Client 使用不同 Server，也可以让同一 Client 同时连接多个 Server。之后再为“节点 + Server”组合配置隧道。

完整的原始 Compose、环境变量、端口、验收和备份说明见 [部署指南](/deployment)。
