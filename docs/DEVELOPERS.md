# Go developer guide

## Tooling

CI supports Go 1.22 and 1.23. Format with gofumpt and goimports, lint with golangci-lint (govet, staticcheck, errcheck, and revive), and run tests with the race detector. The maintenance scripts use Python 3's standard library.

Install the hooks with `lefthook install`. The commit message hook uses `npx` to run commitlint with the repository's Conventional Commits rules.

## Change placement

Use the root `frontal` package for the unified client and service namespaces. Keep shared HTTP behavior in `pkg/resources`, authentication in `pkg/authentication`, HTTP handling and API errors in `pkg/handlers`, and general helpers in `pkg/utils`. Use contract routes from `contracts/sdk-endpoints.json`; do not invent endpoint paths or response schemas.

Keep unit tests beside the source they cover. Use `net/http/httptest.Server` so tests never need a Frontal backend. `ExampleXxx` functions should execute with `go test` and cover public Go examples in the README.

## Contract workflow

`contracts/openapi/` and `contracts/sdk-endpoints.json` are committed inputs. Run `go generate ./pkg/resources` after changing the inventory, then run `python3 scripts/check_contracts.py` to validate snapshot hashes, catalog parity, and the AI operation mapping. Run `python3 scripts/generate_docs_manifest.py` after changing Markdown documentation.
