# Go developer guide

## Tooling

Use Go 1.22 or later and the native commands listed in [`README.md`](../README.md). The SDK runtime is written in Go. The repository maintenance scripts use Python 3's standard library to parse contract JSON and build the documentation index.

## Change placement

Keep client orchestration in `pkg/resources/`, private HTTP execution in the private transport in `pkg/resources/`, and shared concerns in the matching support packages. Use endpoint paths from the committed inventory. Update generated endpoint descriptors with `go generate ./pkg/resources` after inventory changes. Keep unit tests beside source as `*_test.go`; use `tests/` for integration tests spanning packages.

## Contract workflow

`contracts/openapi/` and `contracts/sdk-endpoints.json` are shared input snapshots. Run `python3 scripts/check_contracts.py` to check the snapshot files and `python3 scripts/generate_docs_manifest.py` after changing Markdown documentation.
