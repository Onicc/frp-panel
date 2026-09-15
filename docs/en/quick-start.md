# Quick start

frp-panel uses **one Master, multiple independent FRPS Servers, and multiple FRPC Clients**. Master is control plane only; business traffic passes directly through the selected FRPS.

## 1. Master

Deploy the single Master with the repository `compose.yaml`. Configure `APP_GLOBAL_SECRET` and the public HTTPS API/WSS RPC endpoints from `.env.sample`. Enable registration only long enough to create the initial Owner.

## 2. Server (FRPS)

Open **Servers → Create server**, then enter the target host's public address and bind port. The console displays a complete `compose.yaml`; deploy it on that Linux server. Create every FRPS host separately and never share generated files or data volumes.

## 3. Client (FRPC)

Open **Nodes → Add node**, select Linux, macOS, or Windows, and run the generated one-use command on the target node. The installer uses OS-owned locations and registers the native service.

## 4. Assign routes

On **Nodes**, assign the FRPS routes used by each Client. Different Clients may use different Servers, and one Client may connect to multiple Servers. Configure tunnels against an explicit node and Server pair.

See the [deployment guide](/en/deployment) for raw Compose files, environment settings, ports, acceptance checks, and backup guidance.
