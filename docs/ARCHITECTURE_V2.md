# v2 architecture

frp-panel v2 deliberately breaks v1 packaging and database compatibility. It has two deliverables:

1. `frp-panel` is the Linux controller. It serves the Web console and API, accepts Agent RPC connections, and runs the default embedded FRPS.
2. `frp-panel-agent` is the node program for Linux, macOS, and Windows. It runs FRPC and reports only the optional capabilities available on that OS.

## Current control flow

The browser authenticates with an HttpOnly, Secure, SameSite=Strict cookie. The controller authorizes every route from the user's stored role. Creating a node produces a random, hashed, ten-minute enrollment token. Its first redemption atomically consumes it and returns a distinct long-lived node credential; replay is rejected.

After enrollment, the Agent uses the established RPC transport to receive FRPC configuration and report status. `internal/protocol/v2` defines capability and revision message types for the next transport migration, but revision-based desired-state delivery is not yet the active Agent transport. Keeping that boundary explicit prevents the documentation from promising semantics that are not deployed.

## Storage and networking

- SQLite is the single-node default; PostgreSQL is supported for external storage.
- A monotonic `schema_migrations` registry replaces startup `AutoMigrate`. v2 does not import a v1 database.
- Public API/RPC addresses are explicit configuration and are never inferred from a listen socket.
- The controller container runs as UID/GID `10001`; persistent data is under `/data`.

## Web console contract

The console is a Vite/React SPA. Mutation dialogs share one contract: success resets the form, closes the dialog, and refreshes the resource; failure keeps the dialog open and displays the server's RFC 9457 problem detail. The same behavior is covered by component and browser tests.
