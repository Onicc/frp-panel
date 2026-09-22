# v2 architecture

frp-panel v2 deliberately breaks v1 packaging and database compatibility. It has three independently managed roles:

1. `frp-panel master` is the single Linux control plane. It serves the Web console/API and accepts managed RPC connections; it carries no tunnel traffic.
2. `frp-panel server` is an independently deployed Linux FRPS data plane. Any number of Server instances enroll with the Master and receive their FRPS configuration.
3. `frp-panel-agent` is the Client Agent for Linux, macOS, and Windows. It runs the FRPC connections required by that Client's Tunnels and reports only the optional capabilities available on that OS.

## Current control flow

The browser authenticates with an HttpOnly, Secure, SameSite=Strict cookie. Master authorizes every API operation from the user's stored role. Creating a Server or Client produces a random, hashed, ten-minute enrollment token. Its first redemption atomically consumes it and returns a distinct long-lived credential; replay is rejected.

After enrollment, Servers and Client Agents establish their managed RPC streams, while their startup and periodic legacy configuration pulls remain active. v2 Tunnel CRUD stores desired rows in Master and, when the Client is online and both resources are enrolled, best-effort pushes an assembled FRPC configuration to the Client. The first Tunnel for a Client–Server pair creates its FRPC connection; later Tunnels for that pair reuse it. A physical Client has a dedicated FRP proxy namespace, so different Clients can use the same Tunnel name on one Server, while a name remains unique within one Client. Deleting the last Tunnel of a pair sends an empty configuration for that pair; the Agent stops that FRPC before removing it from its registry. On Client reconnect, Master removes all prior runtime FRPC connections and replays durable Tunnel rows. On Server reconnect or configuration change, Master reapplies its Client pairs. Managed FRPC connections retry login while FRPS is still starting; a stopped FRPC is recreated on a later update. Delivery is still best-effort, not revision-acknowledged; `internal/protocol/v2` defines capability and revision message types for the next transport migration.

Client map location uses a manually configured public IP first, a recent direct Agent public-IP probe second, and the Master-observed connection IP last. The observed address may be a VPN or proxy exit. The Agent bypasses HTTP proxy environment variables for the probe, but a transparent/TUN proxy can still intercept it; use the Client location override in that case. Agent reports are authenticated with its long-lived credential and do not disclose it to the browser.

## Storage and networking

- SQLite is the single-instance default; PostgreSQL is supported for external storage.
- A monotonic `schema_migrations` registry replaces startup `AutoMigrate`. v2 does not import a v1 database.
- Public API/RPC addresses are explicit configuration and are never inferred from a listen socket.
- Master and Server containers run as UID/GID `10001`. Master state and each Server credential live in separate `/data` volumes.

## Web console contract

The console is a Vue 3/Vite SPA. Its layout and common primitives are adapted
only from the sub2api component set (see the frontend attribution notice).
Mutation dialogs share one contract: success resets the form, closes the dialog,
and refreshes the resource; failure keeps the dialog open and displays the
server's RFC 9457 problem detail.
