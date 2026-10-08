# Go developer guide

## Tooling

Use Go 1.22 or later and the native commands listed in [`README.md`](../README.md). The SDK runtime is written in Go. Two small repository maintenance scripts use Python 3's standard library to parse contract JSON and build the documentation index.

## Change placement

Keep transport concerns in `internal/core`, SDK construction in `pkg/sdk`, and endpoint behavior in its matching `pkg/<service>` package. Update API types, examples, and the generated migration matrix with each implemented operation. Put tests next to source files as `*_test.go`.

## Contract workflow

`contracts/openapi/` and `contracts/sdk-endpoints.json` are shared input snapshots. Run `python3 scripts/check_contracts.py` to check the snapshot files and `python3 scripts/generate_docs_manifest.py` after changing Markdown documentation.
