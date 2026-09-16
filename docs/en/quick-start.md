# Quick start

frp-panel uses **one Master, multiple independent FRPS Servers, and multiple FRPC Clients**. Master is control plane only; business traffic passes directly through the selected FRPS.

## 1. Master

Deploy the single Master with the repository `compose.yaml`. Configure `APP_GLOBAL_SECRET` and the public HTTPS API/WSS RPC endpoints from `.env.sample`. Enable registration only long enough to create the initial Owner.

## 2. Server (FRPS)

Open **Servers → Create server**, then enter the target host's public address and bind port. The console displays a complete `compose.yaml`; deploy it on that Linux server. Create every FRPS host separately and never share generated files or data volumes.

## 3. Client (FRPC)

Open **Clients → Add Client**, select Linux, macOS, or Windows, and run the generated one-use command on the target Client host. The installer places the Client Agent in OS-owned locations and registers the native service.

## 4. Create a Tunnel

Open **Tunnels → Create Tunnel**, directly select a Client and Server, then configure the local service and public port. A Client may own multiple Tunnels and each may select a different Server; Master automatically creates, shares, and removes the underlying FRPC connections.

See the [deployment guide](/en/deployment) for raw Compose files, environment settings, ports, acceptance checks, and backup guidance.
