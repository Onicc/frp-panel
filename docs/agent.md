# Agent 安装与维护

控制台生成的引导命令先下载到临时目录、校验 SHA-256，再把 Agent 原子写入系统目录。服务命令行只包含配置文件路径，节点凭据保存在权限为 `0600` 的配置中。

| 系统 | 程序 | 配置与状态 | 服务管理器 |
|---|---|---|---|
| Linux | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/etc/frp-panel/agent.yaml`、`/var/lib/frp-panel` | systemd |
| macOS | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/Library/Application Support/frp-panel` | launchd |
| Windows | `%ProgramFiles%\frp-panel\frp-panel-agent.exe` | `%ProgramData%\frp-panel` | Windows SCM |

维护命令：

```bash
frp-panel-agent service status
frp-panel-agent service restart
frp-panel-agent doctor --json
frp-panel-agent update --version edge
frp-panel-agent service uninstall --purge
```

`uninstall` 保留配置和状态；只有明确添加 `--purge` 才删除。升级会保留上一版本，并在替换失败时回滚。

macOS 尚未进行 Apple 公证，Windows 尚未提供 Authenticode 签名。稳定部署前请依据 GitHub Release 的 `checksums.txt` 验证下载内容。

引导脚本和 `update` 默认跟随滚动的 `edge` 发布；生产环境应显式传入经过评估的 `v*` 标签，或使用 `--version latest` 选择最新稳定版本。
