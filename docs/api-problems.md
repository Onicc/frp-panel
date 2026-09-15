# API problem responses

The v2 HTTP API uses `application/problem+json` following RFC 9457. `type` identifies the documentation, `title` is stable for clients, `status` is the HTTP status, `detail` is safe human-readable context, and `instance` is the request path. Secrets, tokens, and internal stack traces must never appear in a problem response.
