# Go SDK release readiness

The Go module provides one context-first client with service namespaces for every group in the committed endpoint inventory. It includes Bearer authentication, request IDs stable across retries, configurable environment headers, bounded JSON decoding, safe-method retries, typed API errors, generic pagination, cancellable SSE channels, and contract checks.

CI checks Go 1.22 and 1.23, formatting, build, lint, race tests, executable examples, package documentation, and contract drift. Tagged releases compare the public Go API with the latest earlier release in the same major version, run build, race, and contract checks, then use GoReleaser to publish a GitHub release. API comparison is skipped when crossing major versions. The v2 module path marks the breaking package redesign.

Service operations use contract-backed generic request and response values. Add endpoint-specific request structs and higher-level helpers when the committed OpenAPI snapshots define stable wire schemas and behavior.
