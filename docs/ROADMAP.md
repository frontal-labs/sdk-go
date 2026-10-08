# Go SDK roadmap

- [x] Create one Go module with repository health files and Go 1.22/1.23 CI.
- [x] Wire gofumpt, goimports, golangci-lint, race tests, executable examples, documentation, and contract gates.
- [x] Implement context-first HTTP transport, Bearer auth, request IDs, environment headers, retries, timeouts, and typed API errors.
- [x] Expose all committed service groups from one `frontal.Client` and dispatch every catalog operation through its scoped service.
- [x] Cover retries, error mapping, pagination, configuration, and SSE behavior with mock HTTP servers.
- [x] Keep README Go examples executable under `go test` and verify the committed OpenAPI snapshots and generated catalog.
- [x] Add SDK guidance files and tagged GoReleaser release automation with API compatibility checks.
- [ ] Publish the first `v1.0.0` tag after release review.
- [ ] Add endpoint-specific request and response types where the OpenAPI contract declares stable schemas.
