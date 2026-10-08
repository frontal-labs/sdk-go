---
name: frontal-sdk-go
description: Build Go integrations with Frontal or change, test, and document the Frontal Go SDK itself.
---

# Frontal Go SDK

Use this skill for both downstream Go integrations and work inside this SDK repository. Check the repository source and committed API contracts before relying on a method or wire shape; the public contract inventory is `contracts/sdk-endpoints.json` and the OpenAPI snapshots under `contracts/openapi/`.

## Use the client

Import `github.com/frontal-labs/sdk-go` as `frontal`. Construct one client with `frontal.New(opts ...frontal.Option)` and reuse it across service calls. Pass `context.Context` as the first argument to every operation. Service fields cover the contract inventory; inspect `client.<Service>.Endpoints()` or `Endpoint(method, path)` before dispatching through `Call(ctx, resources.Request, out)`.

```go
client, err := frontal.New(frontal.WithAPIKey(apiKey))
if err != nil {
	return err
}
var result map[string]any
err = client.Agents.Call(ctx, resources.Request{
	Endpoint: resources.Endpoint{Service: "agents", Method: http.MethodGet, Path: "/agents/health"},
}, &result)
```

Keep caller-owned request and response structs local when the contract does not define a stable SDK model. Do not invent routes, fields, or convenience method names. Use the service convenience API only when present in this checkout.

## Configuration and lifecycle

Go does not load `.env` files. Environment defaults are `FRONTAL_API_KEY`, `FRONTAL_API_URL`, `FRONTAL_ENV`, `FRONTAL_DEBUG`, and `FRONTAL_TIMEOUT`; explicit functional options override them. The default API URL is `https://api.frontal.dev/v1`. Use `WithHTTPClient` and `WithBaseURL` for test servers or compatible gateways. Never put credentials in a URL, log, or returned error.

## Runtime patterns

- Preserve context cancellation and deadlines through network operations.
- Wrap errors with `%w` so `errors.Is` and `errors.As` keep working.
- Inspect `*frontal.APIError` with `errors.As`; use exported classifiers for auth, rate limit, validation, server, and network failures. Preserve request IDs when reporting failures.
- Use `FetchPage[T]` for cursor metadata, `PollUntil[T]` for bounded polling, and `Watch[T]` for JSON SSE. Cancel the context to stop a watch and always check its terminal error.
- Avoid retrying non-idempotent operations unless the API contract and caller provide an idempotency guarantee.

## Change the SDK

Follow the existing package boundaries: unified API in the root `frontal` package, shared transport and endpoint catalog in `pkg/resources`, auth in `pkg/authentication`, response/error/SSE handling in `pkg/handlers`, defaults in `pkg/headers`, and shared helpers in `pkg/utils`. Keep tests beside package code; use `tests/` for cross-package integration coverage. Update templates and docs when a public pattern changes.

When endpoint inventory changes, update the contract source as instructed by the task, regenerate the catalog with `go generate ./pkg/resources`, and run the contract checker. Do not hand-edit generated catalog output unless the generator itself is being fixed.

## Tests and quality

Use `net/http/httptest.Server` with an injected HTTP client for deterministic, offline HTTP tests. Cover request method/path, headers, serialization, success decoding, API errors, cancellation, and stream termination where relevant. Do not require a live backend or real key in ordinary tests.

Useful checks, from the repository root:

```bash
gofumpt -w .
goimports -w .
go generate ./pkg/resources
go build ./...
golangci-lint run ./...
go test -race ./...
go test -run '^Example' ./...
python3 scripts/check_contracts.py
```

CI supports Go 1.22 and 1.23. See `AGENTS.md`, `docs/ARCHITECTURE.md`, `tests/README.md`, and `CONTRIBUTING.md` for repository-specific details.
