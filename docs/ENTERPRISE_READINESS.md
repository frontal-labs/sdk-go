# Go SDK release readiness

The Go module provides one context-first client with service namespaces for every group in the committed endpoint inventory. It includes Bearer authentication, per-attempt request IDs, configurable environment headers, bounded JSON decoding, safe-method retries, typed API errors, generic pagination, cancellable SSE channels, and contract checks.

CI checks Go 1.22 and 1.23, formatting, build, lint, race tests, executable examples, package documentation, and contract drift. Tagged releases run `apidiff` against the preceding tag and use GoReleaser to publish a GitHub release. The first v1.0.0 tag has no earlier version for API comparison.

Service operations use contract-backed generic request and response values. Endpoint-specific request structs and higher-level domain workflows can be added later when they can be traced to stable OpenAPI schemas or behavior in the TypeScript reference.
