# Supported platforms

| Component | Platform | Architectures | Service manager | FRPC | Functions | Remote shell | WireGuard |
|---|---|---|---|---|---|---|---|
| Controller | Linux / Docker | amd64, arm64 | container | n/a | controller API | n/a | Linux host support required |
| Agent | Linux | amd64, arm64 | systemd | yes | optional | optional | optional |
| Agent | macOS 13.5+ | amd64, arm64 | launchd | yes | optional | optional | no |
| Agent | Windows 11 / Server 2022+ | amd64, arm64 | Windows SCM | yes | no | optional | no |

The Agent does not run FRPS in v2; the controller owns the embedded FRPS. Optional privileged features are disabled unless explicitly enabled in protected Agent configuration. Enabling Functions also requires an absolute path to an operator-installed `workerd` binary. Unsupported capabilities fail validation instead of being silently attempted, and disabled privileged RPC events are rejected by the Agent.

## Installation locations

- Linux: binary `/usr/local/libexec/frp-panel/frp-panel-agent`, config `/etc/frp-panel/agent.yaml`, data `/var/lib/frp-panel`.
- macOS: binary `/usr/local/libexec/frp-panel/frp-panel-agent`, config/data `/Library/Application Support/frp-panel`, LaunchDaemon `/Library/LaunchDaemons/io.github.onicc.frp-panel.agent.plist`.
- Windows: binary `%ProgramFiles%\frp-panel\frp-panel-agent.exe`, config/data `%ProgramData%\frp-panel`, service `frp-panel-agent`.

## Verification level

- macOS arm64: native staged install/uninstall, protected paths, enrollment, controller connection, and diagnostics.
- Linux amd64: container-staged systemd layout, ownership/modes, enrollment, and purge uninstall.
- Linux controller amd64: non-root Docker runtime with SQLite and PostgreSQL migration/restart checks.
- Other published targets: compile and packaging validation. Native OS service smoke tests remain release-gating work for dedicated runners.
