# v2 optimization review

## Completed in this rewrite

| Area | Previous failure mode | v2 result |
|---|---|---|
| Packaging | Bootstrap left binaries/config beside the copied command | Temporary, checksum-verified download followed by atomic OS-native installation |
| Platform support | Linux-only client entry point | Dedicated Agent builds for Linux, macOS, and Windows on amd64/arm64, with native service definitions |
| User experience | Mutation dialogs remained open after successful create/update | Shared controlled-dialog behavior, inline server errors, deterministic refresh, component and browser tests |
| Frontend | Large legacy Next.js/component surface with duplicated patterns | Small Vite 8, React 19, TypeScript 6 SPA with responsive bilingual navigation and reusable primitives |
| Authentication | Browser-managed bearer token and generic token signing | Secure cookie session, role-derived authorization, bootstrap-only Owner registration |
| Enrollment | Long-lived credential in copied commands | Expiring, one-use, hashed token exchanged for a distinct hashed Agent secret |
| Persistence | Implicit schema mutation at startup | Ordered, transactional, idempotent SQLite/PostgreSQL migration registry |
| Runtime | Monolithic Linux client/controller packaging | Separate controller and Agent commands/images; non-root controller image |
| Delivery | Floating CI dependencies and incomplete artifacts | SHA-pinned Actions, cross-build gates, checksums, SBOMs, provenance, dual multi-architecture images |
| Maintainability | Historical generated UI and unused compatibility layers | Old UI, client executable, obsolete build scripts, and unused database adapter removed |

## Next optimization sequence

1. Replace remaining v1 management and RPC handlers resource-by-resource with typed v2 endpoints; add object-level authorization and audit tests with each replacement.
2. Activate revision-based desired state and acknowledgements from `internal/protocol/v2`, including reconnect and partial-apply integration tests.
3. Add revocable server-side sessions and envelope encryption for recoverable FRPS secrets.
4. Add native Windows and macOS CI runners for service lifecycle smoke tests, then sign/notarize stable installers.
5. Establish latency, memory, reconnect, and 10k-resource benchmarks before changing polling/storage internals; gate regressions with recorded baselines.

The sequence favors verified behavior and removal of obsolete code over compatibility shims. A v2 stable tag should not be cut until the remaining security gates in [SECURITY.md](./SECURITY.md) are accepted or closed.
