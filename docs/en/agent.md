# Client / Agent installation and maintenance

The console-generated bootstrap downloads into a temporary directory, verifies SHA-256, then atomically installs the Client Agent into an OS-owned location. The service command contains only a config path; mode `0600` protects Client credentials.

Create the enrollment command under **Clients → Add Client**, select Linux, macOS, or Windows, and run the complete copied command on the target Client host. Do not assemble an enrollment command from this page: its one-time token is valid for ten minutes and expires after its first successful redemption.

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
sudo /usr/local/libexec/frp-panel/frp-panel-agent update --version v2.1.0
sudo /usr/local/libexec/frp-panel/frp-panel-agent service uninstall --purge
```

```bash [macOS]
sudo launchctl print system/io.github.onicc.frp-panel.agent
sudo /usr/local/libexec/frp-panel/frp-panel-agent service restart
sudo /usr/local/libexec/frp-panel/frp-panel-agent doctor --json
sudo /usr/local/libexec/frp-panel/frp-panel-agent update --version v2.1.0 --restart-service=false && sudo /usr/local/libexec/frp-panel/frp-panel-agent service restart
sudo /usr/local/libexec/frp-panel/frp-panel-agent service uninstall --purge
```

```powershell [Windows PowerShell (Administrator)]
Get-Service frp-panel-agent
Restart-Service frp-panel-agent
& "$env:ProgramFiles\frp-panel\frp-panel-agent.exe" doctor --json
& "$env:ProgramFiles\frp-panel\frp-panel-agent.exe" update --version v2.1.0
& "$env:ProgramFiles\frp-panel\frp-panel-agent.exe" service uninstall --purge
```

:::

`uninstall` preserves configuration and state unless `--purge` is explicit. Updates retain the previous binary and roll back a failed replacement.

## Recovering a failed Linux installation

The enrollment token is redeemed before the service starts. If the command wrote `/etc/frp-panel/agent.yaml` and then failed while installing or starting systemd, keep that file and rerun the original installation command. When the token cannot be redeemed again, the installer reuses the protected configuration only if both its Client ID and Master endpoints match. If the original command is no longer available, repair the installation from the existing configuration:

```bash
curl -fsSL https://raw.githubusercontent.com/Onicc/frp-panel/main/install.sh | sudo bash -s -- --config /etc/frp-panel/agent.yaml
```

Do not run `uninstall --purge` first. Purging removes the Client credential, so a new Client and installation command must then be created in Master.

Apple notarization and Windows Authenticode are not provided yet. Verify downloads against `checksums.txt` from the GitHub Release before stable deployments.

The bootstrap and `update` default to the latest stable release; console-generated install commands pin the Master's `vX.X.X` release. The `v2.1.0` examples show the first stable release; copy the current target command from Clients for future updates. Existing `edge` Agents can use that command for a one-time migration without deleting their protected configuration.

See the [deployment guide](/en/deployment) for the complete Master, independent Server, Client, and Tunnel rollout.
