---
name: go-context-http
description: Design, implement, and test context-aware Go HTTP client behavior in the Frontal SDK.
---

# Go Context and HTTP

Use this skill when changing request execution, cancellation, timeouts, retries, or HTTP behavior in `sdk-go`.

- Thread `context.Context` from the public call to every request and blocking operation. Do not replace caller context with `context.Background()` inside library code.
- Preserve cancellation and deadlines; release response bodies and stop goroutines on every exit path.
- Wrap underlying failures with `%w`; keep `errors.Is` and `errors.As` useful. Preserve typed API error status, code, request ID, and retryability.
- Keep credentials out of URL parameters, logs, and error text. Prefer existing auth/header helpers over duplicating header logic.
- Use `net/http/httptest` with injected client/base URL to verify method, path, headers, body, decoding, cancellation, and error handling without live services.
- Retry only safe operations. Backoff and retry decisions must honor context cancellation and the SDK's established option defaults.

Read `AGENTS.md` and the owning package implementation before changing transport behavior.
