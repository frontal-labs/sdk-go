---
name: go-contract-endpoints
description: Add or maintain contract-backed operations and generated endpoint inventory in the Frontal Go SDK.
---

# Go Contract Endpoints

Use this skill when adding a service operation, changing route inventory, or synchronizing the Go endpoint catalog.

1. Confirm the HTTP method, route, and known wire fields in `contracts/sdk-endpoints.json` and the matching OpenAPI snapshot. Never infer a route from a service name.
2. Follow the existing service dispatch shape (`Endpoint(s)` plus `Call`) and keep response models caller-owned unless the contract and SDK define a stable model.
3. Update the contract inventory only when that source is in scope; regenerate generated catalog files with `go generate ./pkg/resources` instead of editing generated output by hand.
4. Add offline tests for endpoint selection, request encoding, and decoding/errors. Update examples and endpoint docs when public usage changes.
5. Run `python3 scripts/check_contracts.py`, `go build ./...`, and focused package tests; run the full checks listed in `AGENTS.md` for broad changes.

Keep package responsibilities and Go version support described in `AGENTS.md`.
