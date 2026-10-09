# Go SDK architecture

The v2 module exposes one root `frontal.Client` and top-level packages for each API service. Service packages own their endpoint catalog and scoped client. Shared transport, authentication, response decoding, headers, and URL behavior live in private `internal/` packages.

## Package layout

- `client.go` configures one shared client and wires its service clients.
- `service.go` provides raw contract calls, typed operation binding, pagination, polling, and SSE helpers.
- Top-level service packages expose `Client`, `Endpoint`, and `Request` types generated from `contracts/sdk-endpoints.json`.
- `internal/core` owns the HTTP client, request/endpoint types, retry behavior, and generated endpoint catalog.
- `internal/authentication`, `internal/handlers`, `internal/headers`, and `internal/utils` own shared implementation details.
- `contracts/` holds committed endpoint inventory and OpenAPI snapshots.

The service generator creates an isolated package for every contract group. Each package validates its calls against its own catalog and delegates to the root client's shared transport. No service package imports the root package, so there are no import cycles.

## Request flow

`Application → frontal.Client → top-level service client → internal/core → authentication + headers → net/http → Frontal API`

Use service package `Endpoint` and `Request` types for scoped contract calls. Use `frontal.Client.Call` for a generic catalog-backed call and `frontal.Client.Request` for direct HTTP access. The service package `Stream` method and root `Watch[T]` helper return cancellable SSE streams.

## Configuration and errors

`frontal.New` reads `FRONTAL_API_KEY`, `FRONTAL_API_URL`, `FRONTAL_ENV`, `FRONTAL_DEBUG`, and `FRONTAL_TIMEOUT`. API keys are sent only as Bearer authorization. Debug logs contain request metadata, not credentials or bodies. JSON responses are bounded; request IDs remain stable across safe retries.

Non-success responses become `*frontal.APIError` values with status, code, request ID, and retryability. Only safe GET and HEAD calls retry transient failures. Streams use the request context for cancellation and do not use a client-wide timeout.

## Typed API coverage

Endpoint descriptors are generated from `contracts/sdk-endpoints.json`. Operation-specific request and response types are generated only when committed contracts define their wire schema. Calls without those schemas remain available through scoped `Call` or root `Request` methods until authoritative schemas are added.
