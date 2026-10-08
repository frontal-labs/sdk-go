# Project scripts

- `check_contracts.py` verifies OpenAPI snapshot hashes, all spec operations, every SDK endpoint descriptor and client namespace, the complete AI endpoint mapping, and generated catalog drift. `go test ./contracts` resolves every OpenAPI operation to a valid HTTP request and verifies service namespaces.
- `check_apidiff.sh <tag>` compares exported module APIs against an earlier release tag. The release workflow skips the comparison for the first release.
- `generate_docs_manifest.py` refreshes the root and docs MCP indexes.
- `generate_endpoints.py` regenerates the endpoint catalog from `contracts/sdk-endpoints.json`.

The commit hooks are configured in `lefthook.yml`; `commitlint.config.cjs` applies Conventional Commits header rules.
