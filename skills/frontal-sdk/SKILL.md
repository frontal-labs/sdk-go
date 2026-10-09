---
name: frontal-sdk-go
description: Build Go integrations with Frontal or change, test, and document the Frontal Go SDK itself.
---

# Frontal Go SDK

Use this skill for downstream integrations and SDK work. Check the committed endpoint inventory and OpenAPI snapshots before relying on a route or wire shape. Do not invent operation-specific schemas.

## Use the client

Import `github.com/frontal-labs/sdk-go/v2` as `frontal`, plus the top-level service package you use. Construct one reusable client with `frontal.New(opts ...frontal.Option)`. Pass `context.Context` as the first argument to every operation.

```go
import (
    "github.com/frontal-labs/sdk-go/v2"
    "github.com/frontal-labs/sdk-go/v2/agents"
)

client, err := frontal.New(frontal.WithAPIKey(apiKey))
if err != nil {
    return err
}
endpoint, ok := client.Agents.Endpoint("GET", "/agents/health")
if !ok {
    return frontal.ErrUnknownEndpoint
}
var result struct {
    Status string `json:"status"`
}
return client.Agents.Call(ctx, agents.Request{Endpoint: endpoint}, &result)
```

Service package `Call` validates the operation against that service's endpoint inventory. Use `client.Call(frontal.Request{...}, out)` for generic inventory-backed requests and `client.Request(ctx, method, path, body, out)` for direct requests. Use `BindOperation` only with request and response shapes from an authoritative contract.

## Configuration and lifecycle

Go does not load `.env` files. Environment defaults are `FRONTAL_API_KEY`, `FRONTAL_API_URL`, `FRONTAL_ENV`, `FRONTAL_DEBUG`, and `FRONTAL_TIMEOUT`; explicit options override them. The default API URL is `https://api.frontal.dev/v1`. Use `WithHTTPClient` and `WithBaseURL` for test servers or compatible gateways. Never put credentials in URLs, logs, or errors.

## Runtime patterns

- Preserve context cancellation and deadlines through network operations.
- Wrap errors with `%w`; inspect `*frontal.APIError` with `errors.As` and use the exported error classifiers.
- Use `FetchPage[T]` for common cursor metadata, `PollUntil[T]` for bounded polling, and `Watch[T]` for JSON SSE. Cancel the context to stop a watch and check its terminal error.
- Automatic retries apply only to safe GET and HEAD operations.

## Change the SDK

Keep the unified client in the root `frontal` package, service clients and catalogs in top-level service packages, and shared implementation under `internal/`. Keep tests beside code; use `tests/` for cross-package integration coverage. Update templates and documentation when public patterns change.

When the endpoint inventory changes, regenerate catalogs with `go generate ./internal/core` and run `python3 scripts/check_contracts.py`. Do not hand-edit generated files.

## Tests and quality

Use `net/http/httptest.Server` with an injected HTTP client for deterministic, offline tests. Cover request method/path, headers, serialization, success decoding, API errors, cancellation, and stream termination. Do not require a live backend or real key in ordinary tests.

```bash
gofumpt -w .
goimports -w .
go generate ./internal/core
go build ./...
golangci-lint run ./...
go test -race ./...
go test -run '^Example' ./...
python3 scripts/check_contracts.py
```

CI supports Go 1.22 and 1.23. See `AGENTS.md`, `docs/ARCHITECTURE.md`, `tests/README.md`, and `CONTRIBUTING.md`.
