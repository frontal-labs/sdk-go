# Frontal Go SDK - Agent Instructions

## Repository scope

This is one Go module. The SDK lives under `pkg/`; use the committed OpenAPI snapshots and `contracts/sdk-endpoints.json` as the endpoint source of truth.

## Layout

- `pkg/resources/` — public API client, request/endpoint types, and generated endpoint catalog.
- `pkg/authentication/` — API key validation and Bearer authentication.
- `pkg/handlers/` — HTTP request/response handling, API errors, and event stream decoding.
- `pkg/headers/` — HTTP header names and defaults.
- `pkg/utils/` — URL, timeout, and retry helpers.
- `tests/` — cross-package integration tests; keep unit tests beside the code they cover.
- `templates/` — Go starter projects using this SDK.

## Go conventions

Pass `context.Context` through network operations and wrap errors with `%w`. Use `net/http/httptest` for deterministic HTTP tests. Keep API keys out of URLs, logs, and errors. Do not invent endpoint-specific method names or response schemas; use the committed contracts. Regenerate the endpoint catalog with `go generate ./pkg/resources` after updating the inventory.

The module path is `github.com/frontal-labs/sdk-go`. The client defaults to `https://api.frontal.dev/v1`; Go does not load `.env` files automatically.

## Key commands

```bash
gofmt -w ./pkg
go generate ./pkg/resources
go build ./...
go vet ./...
go test -race ./...
python3 scripts/check_contracts.py
```
