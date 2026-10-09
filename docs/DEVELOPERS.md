# Go developer guide

## Tooling

CI supports Go 1.22 and 1.23. Format with gofumpt and goimports, lint with golangci-lint (govet, staticcheck, errcheck, and revive), and run tests with the race detector. The maintenance scripts use Python 3's standard library.

## Change placement

Keep the unified client and common helpers in root package `frontal`. Keep service clients and generated endpoint catalogs in top-level service packages. Keep shared transport behavior private under `internal/`. Use routes from `contracts/sdk-endpoints.json`; do not invent endpoint paths or wire schemas.

Keep unit tests beside the source they cover. Use `net/http/httptest.Server` so tests do not need a Frontal backend. Public README examples should have executable `ExampleXxx` counterparts.

## Contract workflow

`contracts/openapi/` and `contracts/sdk-endpoints.json` are committed inputs. Run `go generate ./internal/core` after changing the inventory, then run `python3 scripts/check_contracts.py` to validate snapshot hashes, catalog parity, and service package bindings. Run `python3 scripts/generate_docs_manifest.py` after changing Markdown documentation.
