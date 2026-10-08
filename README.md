# Frontal Go SDK

![Frontal Banner](./banner.png)

One context-first Go client for Frontal AI, agents, workflows, and every other service in the committed endpoint inventory.

## Quickstart

With `FRONTAL_API_KEY` set and `ctx` in scope, call a service operation:

```go
client, err := frontal.New()
if err != nil { log.Fatal(err) }
var health map[string]any
err = client.Agents.Call(ctx, resources.Request{Endpoint: resources.Endpoint{Service: "agents", Method: "GET", Path: "/agents/health"}}, &health)
if err != nil { log.Fatal(err) }
```

The [`ExampleNew`](./example_test.go) test runs the same client call against an `httptest.Server`, so the quickstart behavior stays executable without a Frontal backend.

## Client and services

`frontal.New(opts ...Option)` reads `FRONTAL_API_KEY`, `FRONTAL_API_URL`, `FRONTAL_ENV`, `FRONTAL_DEBUG`, and `FRONTAL_TIMEOUT`. Go does not load `.env` files automatically. `FRONTAL_ENV` defaults to `development`. Options include `WithAPIKey`, `WithBaseURL`, `WithHTTPClient`, `WithTimeout`, and `WithMaxRetries`.

The returned client exposes `AI`, `Agents`, `Workflows`, `Audit`, `Auth`, `Billing`, `Blob`, `Connectors`, `Data`, `Governance`, `Lineage`, `Observability`, `Ontology`, `Pipelines`, `Sandbox`, `Schedules`, and `Webhooks`. Each service lists its contract operations with `Endpoints` and sends an operation with `Call(ctx, resources.Request, out)`. Use `client.Core` or `client.Request` for direct HTTP access.

```go
endpoint, _ := client.AI.Endpoint(http.MethodPost, "/ai/chat/completions")
var result map[string]any
err = client.AI.Call(ctx, resources.Request{Endpoint: endpoint, Body: map[string]any{"model": "model-id", "messages": messages}}, &result)
```

[`ExampleService_Call`](./example_test.go) runs this AI operation against an `httptest.Server`.

`FetchPage[T]` decodes collection and cursor metadata. `Watch[T]` streams JSON SSE events to a receive-only channel; cancel its context to close the request. `PollUntil[T]` polls with a caller-supplied fetch and completion check.

API failures are `*frontal.APIError` values discoverable through `errors.As`, with status, code, request ID, and retryability. Use `IsAuthError`, `IsRateLimitError`, `IsValidationError`, `IsServerError`, and `IsNetworkError` to classify failures.

## Package layout

| Path | Purpose |
| --- | --- |
| `client.go`, `service.go` | Unified client, service namespaces, pagination, polling, and streams |
| `pkg/resources/` | Shared HTTP client and generated endpoint catalog |
| `pkg/authentication/` | API key validation and Bearer authentication |
| `pkg/handlers/` | HTTP request/response handling, API errors, and SSE decoding |
| `pkg/headers/` | HTTP header names and defaults |
| `pkg/utils/` | URL, timeout, and retry helpers |
| `contracts/` | Committed OpenAPI snapshots and route inventory |

## Development

The module supports Go 1.22 and 1.23. CI uses gofumpt, goimports, golangci-lint (govet, staticcheck, errcheck, revive), and runs tests with the race detector.

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

See [`CONTRIBUTING.md`](./CONTRIBUTING.md), [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md), and [`docs/RELEASING.md`](./docs/RELEASING.md).

## License

Apache-2.0. See [`LICENSE.md`](./LICENSE.md).
