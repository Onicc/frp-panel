# v2 architecture

frp-panel v2 deliberately breaks v1 packaging and database compatibility. It has three independently managed roles:

1. `frp-panel master` is the single Linux control plane. It serves the Web console/API and accepts managed RPC connections; it carries no tunnel traffic.
2. `frp-panel server` is an independently deployed Linux FRPS data plane. Any number of Server instances enroll with the Master and receive their FRPS configuration.
3. `frp-panel-agent` is the Client Agent for Linux, macOS, and Windows. It runs the FRPC connections required by that Client's Tunnels and reports only the optional capabilities available on that OS.

## Current control flow

The browser authenticates with an HttpOnly, Secure, SameSite=Strict cookie. Master authorizes every API operation from the user's stored role. Creating a Server or Client produces a random, hashed, ten-minute enrollment token. Its first redemption atomically consumes it and returns a distinct long-lived credential; replay is rejected.

After enrollment, Servers pull FRPS configuration and Client Agents pull the internal FRPC connections required by Tunnels. Creating the first Tunnel for a Client and Server pair creates that connection; later Tunnels share it, and deleting the final Tunnel removes it. Online changes are pushed immediately and periodic reconciliation applies offline changes. `internal/protocol/v2` defines capability and revision message types for the next transport migration, but revision-based desired-state delivery is not yet the active Agent transport.

## Storage and networking

- SQLite is the single-instance default; PostgreSQL is supported for external storage.
- A monotonic `schema_migrations` registry replaces startup `AutoMigrate`. v2 does not import a v1 database.
- Public API/RPC addresses are explicit configuration and are never inferred from a listen socket.
- Master and Server containers run as UID/GID `10001`. Master state and each Server credential live in separate `/data` volumes.

## Web console contract

The console is a Vite/React SPA. Mutation dialogs share one contract: success resets the form, closes the dialog, and refreshes the resource; failure keeps the dialog open and displays the server's RFC 9457 problem detail. The same behavior is covered by component and browser tests.
