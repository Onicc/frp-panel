# Security review and hardening plan

This is the maintained v2 risk register. Severity reflects impact before the listed mitigation.

## Addressed risks

| Risk | Severity | v2 control |
|---|---:|---|
| Shared/default secrets and exposed Client credentials | Critical | Master refuses secrets shorter than 32 characters; API models never serialize persistent credentials; enrollment stores only a token hash and long-lived Client credentials are hashed at rest. |
| Replayable or permanent installation commands | Critical | Cryptographically random enrollment tokens expire after ten minutes and are atomically consumed once. The redeemed Agent secret is different from the bootstrap token. |
| Weak password storage and client-side bearer persistence | High | Passwords use Argon2id (64 MiB, three passes, two lanes). Browser authentication is an HttpOnly, Secure, SameSite=Strict cookie; JWTs are not stored in browser storage or accepted in query strings. |
| Account takeover through unauthenticated password replacement | Critical | Password changes require the current password, use Argon2id for the replacement, clear the current cookie, and force reauthentication. The legacy profile endpoint rejects password fields. |
| FRP credentials exposed through inventory APIs or diagnostic logs | Critical | Client/Server inventory responses omit runtime configuration, and lifecycle logs record only resource identifiers and event types. Configuration bodies, authentication tokens, credential hashes, and RPC payloads are never logged. |
| Excessive authorization | High | Owner, Admin, Operator, and Viewer permissions are derived from the stored role. The generic token-signing endpoint was removed. Public registration works only when explicitly enabled and only while the user table is empty. |
| Cross-origin WebSocket and TLS downgrade | High | Browser upgrades are same-origin unless an exact Origin is allow-listed. Agent TLS verification defaults on with TLS 1.2 minimum; bypass requires an explicit unsafe flag. |
| Arbitrary download or unsafe self-update | High | Automatic workerd download/proxy behavior is disabled by default. Release replacement requires SHA-256, uses an atomic backup, and rolls back failures. |
| Installation in the caller's working directory | High | Bootstrap downloads to an OS temporary directory and installs atomically into OS-owned paths with a dedicated service identity and protected configuration. |
| Accidental privileged feature exposure | High | Functions, remote shell, WireGuard, and download proxies default off; disabled privileged RPC events are rejected. Functions require an explicit absolute `workerd` path and WireGuard is Linux-only. |
| Unbounded or overly informative HTTP failures | Medium | Requests are body-limited, security headers and CSP are applied, sensitive endpoints are rate-limited, and v2 failures use redacted RFC 9457 problem documents. |
| Unreviewed runtime schema mutation | Medium | Numbered, transactional, idempotent migrations support SQLite and PostgreSQL. |
| Supply-chain drift | Medium | CI actions are pinned to commit SHAs; production dependencies, Go vulnerabilities, static analysis, release checksums, binary/image SBOMs, provenance, and multi-architecture builds are gated in CI. |

## Accepted or remaining risks

| Risk | Status / next action |
|---|---|
| Master HTTP has no mandatory built-in public certificate | Terminate HTTPS at a trusted reverse proxy; do not expose port 9000 directly. A future stable profile should refuse non-loopback HTTP without an explicit acknowledgement. |
| Session tokens are signed cookies rather than revocable server-side sessions | Keep expiry short and rotate `APP_GLOBAL_SECRET` after suspected compromise. Add a session table with per-device revocation before multi-tenant use. |
| Some legacy v1 management/RPC handlers remain behind the new console and Agent | Continue resource-by-resource replacement with typed v2 endpoints, object-level authorization tests, and audit records before declaring the API stable. |
| Managed-host credentials must be recoverable locally at runtime | Master stores only hashes. Plaintext FRPS and Agent credentials remain in mode-0600 files on their own persistent hosts; use encrypted disks/volumes and add KMS-backed envelope encryption for hosted deployments. |
| macOS and Windows release binaries are unsigned | SHA-256 and provenance are published. Add Apple notarization and Windows Authenticode before a broadly distributed stable release. |
| Archive/config/WebSocket fuzz coverage is incomplete | Add persistent Go fuzz corpora and malformed-frame tests; keep privileged features disabled by default meanwhile. |
| gVisor WireGuard integration uses a narrow unsafe reflection bridge | It is isolated and scanner-annotated. Replace it when upstream exposes a supported netstack accessor. |

## Operational requirements

- Generate `APP_GLOBAL_SECRET` from at least 32 random bytes and store it in a secret manager.
- Enable `APP_ENABLE_REGISTER` only while creating the initial Owner.
- Restrict Master and Server ingress with firewall rules; expose Master Web/API through HTTPS.
- Back up the database and test restore before upgrades. v2 does not migrate v1 data.
- Keep privileged Client Agent features disabled unless the Client and operator trust boundary requires them.

Report undisclosed vulnerabilities through GitHub private vulnerability reporting for `Onicc/frp-panel`, not a public issue.
