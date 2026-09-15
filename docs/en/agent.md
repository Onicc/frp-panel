# Client / Agent installation and maintenance

The console-generated bootstrap downloads into a temporary directory, verifies SHA-256, then atomically installs the Agent into an OS-owned location. The service command contains only a config path; mode `0600` protects node credentials.

Create the enrollment command under **Nodes → Add node**, select Linux, macOS, or Windows, and run the complete copied command on the target machine. Do not assemble an enrollment command from this page: its one-time token is valid for ten minutes and expires after its first successful redemption.

| OS | Binary | Configuration and state | Service manager |
|---|---|---|---|
| Linux | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/etc/frp-panel/agent.yaml`, `/var/lib/frp-panel` | systemd |
| macOS | `/usr/local/libexec/frp-panel/frp-panel-agent` | `/Library/Application Support/frp-panel` | launchd |
| Windows | `%ProgramFiles%\frp-panel\frp-panel-agent.exe` | `%ProgramData%\frp-panel` | Windows SCM |

The binary is not automatically added to `PATH`. Use its full path or the native service manager.

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

```powershell [Windows PowerShell (Administrator)]
Get-Service frp-panel-agent
Restart-Service frp-panel-agent
& "$env:ProgramFiles\frp-panel\frp-panel-agent.exe" doctor --json
& "$env:ProgramFiles\frp-panel\frp-panel-agent.exe" update --version edge
& "$env:ProgramFiles\frp-panel\frp-panel-agent.exe" service uninstall --purge
```

:::

`uninstall` preserves configuration and state unless `--purge` is explicit. Updates retain the previous binary and roll back a failed replacement.

Apple notarization and Windows Authenticode are not provided yet. Verify downloads against `checksums.txt` from the GitHub Release before stable deployments.

The bootstrap and `update` default to the rolling `edge` release. Production should pass an evaluated `v*` tag explicitly, or use `--version latest` for the latest stable release.

See the [deployment guide](/en/deployment) for the complete Master, independent Server, Client, and route-assignment rollout.
