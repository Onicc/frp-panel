# Agent installation and maintenance

The console-generated bootstrap downloads into a temporary directory, verifies SHA-256, then atomically installs the Agent into an OS-owned location. The service command contains only a config path; mode `0600` protects node credentials.

| OS | Binary | Configuration and state | Service manager |
|---|---|---|---|
| Linux | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/etc/frp-panel/agent.yaml`, `/var/lib/frp-panel` | systemd |
| macOS | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/Library/Application Support/frp-panel` | launchd |
| Windows | `%ProgramFiles%\frp-panel\frp-panel-agent.exe` | `%ProgramData%\frp-panel` | Windows SCM |

Maintenance commands:

```bash
frp-panel-agent service status
frp-panel-agent service restart
frp-panel-agent doctor --json
frp-panel-agent update --version edge
frp-panel-agent service uninstall --purge
```

`uninstall` preserves configuration and state unless `--purge` is explicit. Updates retain the previous binary and roll back a failed replacement.

Apple notarization and Windows Authenticode are not provided yet. Verify downloads against `checksums.txt` from the GitHub Release before stable deployments.

The bootstrap and `update` default to the rolling `edge` release. Production should pass an evaluated `v*` tag explicitly, or use `--version latest` for the latest stable release.
