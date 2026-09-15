# Quick start

## 1. Start the controller

```bash
git clone https://github.com/Onicc/frp-panel.git
cd frp-panel
export APP_GLOBAL_SECRET="$(openssl rand -hex 32)"
export PUBLIC_HOST="panel.example.com"
export APP_ENABLE_REGISTER=true
docker compose up -d
```

Registration is enabled only while you create the initial Owner in the Web console. Set `APP_ENABLE_REGISTER=false` afterward and run `docker compose up -d` again. The Owner manages subsequent users; the public endpoint cannot create more accounts.

Ports are Web/API `9000`, Agent RPC `9001`, and embedded FRPS `7000`. Expose Web/API through HTTPS in production and keep `APP_COOKIE_SECURE=true`.

## 2. Add a node

Open **Nodes → Add node**, enter a name, and create a one-time command valid for ten minutes. Select Linux, macOS, or Windows and run it on that node. Redemption permanently invalidates the token.

## 3. Check it

```bash
frp-panel-agent service status
frp-panel-agent doctor --json
```

Continue with [Agent installation](/en/agent), [configuration](/en/configuration), and the [security baseline](/SECURITY).
