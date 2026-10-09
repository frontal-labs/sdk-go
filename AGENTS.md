# Frontal Go SDK - Agent Instructions

## Repository scope

This is the v2 Go module at `github.com/frontal-labs/sdk-go/v2`. Top-level service packages own their endpoint clients and generated endpoint catalogs. Shared transport and implementation details are private under `internal/`. The committed OpenAPI snapshots and `contracts/sdk-endpoints.json` remain the API source of truth.

## Layout

- Root `frontal` package — unified client, common configuration, errors, raw calls, pagination, polling, and stream helpers.
- Top-level service packages — service-scoped client, endpoint catalog, JSON calls, and streaming operations.
- `internal/core/` — shared authenticated HTTP client, retry/redirect policy, and endpoint inventory.
- `internal/authentication/` — API key validation and Bearer authentication.
- `internal/handlers/` — HTTP request/response handling, API errors, and event stream decoding.
- `internal/headers/` — HTTP header names and defaults.
- `internal/utils/` — URL, timeout, and retry helpers.
- `contracts/` — OpenAPI snapshots and endpoint inventory.
- `templates/` — Go starter projects using this SDK.

## Go conventions

Pass `context.Context` through network operations and wrap errors with `%w`. Use `net/http/httptest` for deterministic HTTP tests. Keep API keys out of URLs, logs, and errors. Do not invent endpoint-specific method names or response schemas; use committed contracts. Regenerate catalogs with `go generate ./internal/core` after updating the inventory.

The module path is `github.com/frontal-labs/sdk-go/v2`. The client defaults to `https://api.frontal.dev/v1`. Go does not load `.env` files automatically.

## Key commands

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
