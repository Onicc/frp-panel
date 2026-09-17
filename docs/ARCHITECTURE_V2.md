# v2 architecture

frp-panel v2 deliberately breaks v1 packaging and database compatibility. It has three independently managed roles:

1. `frp-panel master` is the single Linux control plane. It serves the Web console/API and accepts managed RPC connections; it carries no tunnel traffic.
2. `frp-panel server` is an independently deployed Linux FRPS data plane. Any number of Server instances enroll with the Master and receive their FRPS configuration.
3. `frp-panel-agent` is the Client Agent for Linux, macOS, and Windows. It runs the FRPC connections required by that Client's Tunnels and reports only the optional capabilities available on that OS.

## Current control flow

The browser authenticates with an HttpOnly, Secure, SameSite=Strict cookie. Master authorizes every API operation from the user's stored role. Creating a Server or Client produces a random, hashed, ten-minute enrollment token. Its first redemption atomically consumes it and returns a distinct long-lived credential; replay is rejected.

After enrollment, Servers and Client Agents establish their managed RPC streams, while their startup and periodic legacy configuration pulls remain active. v2 Tunnel CRUD stores desired rows in Master and, when the Client is online and both resources are enrolled, best-effort pushes an assembled FRPC configuration to the Client. The first Tunnel for a pair creates the shared connection; later Tunnels for that pair reuse it. Deletion sends a physical Client-level removal event only when that Client has no other active Tunnel, so pair-specific cleanup is not guaranteed when the Client serves another Server. A Tunnel created while the Client is offline remains durable, but the current reconnect path does not automatically re-apply v2 Tunnel configuration; another Tunnel update is required to trigger delivery. `internal/protocol/v2` defines capability and revision message types for the next transport migration, but revision-based desired-state delivery is not yet the active Agent transport.

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
