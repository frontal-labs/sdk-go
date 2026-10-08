# Frontal Go SDK - Agent Instructions

## Repository scope

This is one Go module for the Frontal API. The root client provides generic contract-backed JSON and raw HTTP requests; endpoint-specific convenience methods are not generated. Treat committed OpenAPI snapshots and `contracts/sdk-endpoints.json` as the source for endpoint shapes.

## Layout

- `client.go` — public `frontal.Client` and request entry points.
- `authentication/`, `headers/`, `handlers/`, `models/`, and `utils/` — shared public SDK components.
- `internal/core/` — private HTTP transport.
- `contracts/` — API snapshots, endpoint inventory, and conformance reports.
- `docs/`, `examples/`, and `templates/` — developer material and starter applications.

## Go conventions

Keep one `go.mod` at the repository root. Pass `context.Context` through network operations and wrap errors with `%w`. Put tests beside source in `*_test.go` files and use `net/http/httptest` for deterministic transport tests. Keep API keys out of URLs, logs, and errors. Do not invent endpoint-specific method names or response schemas; use the committed contracts. Regenerate endpoint descriptors with `go generate ./models` after updating the endpoint inventory.

The module path is `github.com/frontal-labs/sdk-go`. The client defaults to `https://api.frontal.dev/v1`; Go does not load `.env` files automatically.

## Key commands

```bash
gofmt -w client.go authentication headers handlers models utils internal
go generate ./models
go build ./...
go vet ./...
go test -race ./...
python3 scripts/check_contracts.py
```

The maintenance scripts use Python 3's standard library and do not add Python as a runtime dependency.
