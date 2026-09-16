# Client / Agent 安装与维护

控制台生成的引导命令先下载到临时目录、校验 SHA-256，再把 Agent 原子写入系统目录。服务命令行只包含配置文件路径，节点凭据保存在权限为 `0600` 的配置中。

请从 **节点 → 添加节点** 创建注册命令，选择 Linux、macOS 或 Windows 后复制完整命令到目标机器执行。不要从本文拼接注册命令：其中的一次性令牌只有 10 分钟有效，并在首次成功兑换后失效。

| 系统 | 程序 | 配置与状态 | 服务管理器 |
|---|---|---|---|
| Linux | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/etc/frp-panel/agent.yaml`、`/var/lib/frp-panel` | systemd |
| macOS | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/Library/Application Support/frp-panel` | launchd |
| Windows | `%ProgramFiles%\frp-panel\frp-panel-agent.exe` | `%ProgramData%\frp-panel` | Windows SCM |

程序不会自动加入 `PATH`，请使用完整路径或系统服务管理器。

::: code-group

```bash [Linux]
sudo systemctl status frp-panel-agent
sudo systemctl restart frp-panel-agent
sudo /usr/local/libexec/frp-panel/frp-panel-agent doctor --json
sudo /usr/local/libexec/frp-panel/frp-panel-agent update --version edge
sudo /usr/local/libexec/frp-panel/frp-panel-agent service uninstall --purge
```

```bash [macOS]
sudo launchctl print system/io.github.onicc.frp-panel.agent
sudo /usr/local/libexec/frp-panel/frp-panel-agent service restart
sudo /usr/local/libexec/frp-panel/frp-panel-agent doctor --json
sudo /usr/local/libexec/frp-panel/frp-panel-agent update --version edge
sudo /usr/local/libexec/frp-panel/frp-panel-agent service uninstall --purge
```

```powershell [Windows PowerShell（管理员）]
Get-Service frp-panel-agent
Restart-Service frp-panel-agent
& "$env:ProgramFiles\frp-panel\frp-panel-agent.exe" doctor --json
& "$env:ProgramFiles\frp-panel\frp-panel-agent.exe" update --version edge
& "$env:ProgramFiles\frp-panel\frp-panel-agent.exe" service uninstall --purge
```

:::

`uninstall` 保留配置和状态；只有明确添加 `--purge` 才删除。升级会保留上一版本，并在替换失败时回滚。

## Linux 安装失败后的恢复

注册令牌在服务启动前就会被兑换。如果命令已经写入 `/etc/frp-panel/agent.yaml`，但随后在 systemd 安装或启动阶段失败，请保留该文件并直接重新执行原安装命令；安装器会在令牌无法再次兑换时，仅复用节点 ID、Master 地址均匹配的现有受保护配置。原命令已丢失时，可用现有配置修复安装：

```bash
curl -fsSL https://raw.githubusercontent.com/Onicc/frp-panel/main/install.sh | sudo bash -s -- --config /etc/frp-panel/agent.yaml
```

不要先执行 `uninstall --purge`，否则节点凭据会被删除，必须在 Master 中创建新节点并生成新命令。

macOS 尚未进行 Apple 公证，Windows 尚未提供 Authenticode 签名。稳定部署前请依据 GitHub Release 的 `checksums.txt` 验证下载内容。

引导脚本和 `update` 默认跟随滚动的 `edge` 发布；生产环境应显式传入经过评估的 `v*` 标签，或使用 `--version latest` 选择最新稳定版本。

Master、独立 Server 和 Client 的完整上线顺序及链路分配见 [部署指南](/deployment)。
