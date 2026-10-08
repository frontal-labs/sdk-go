# Go SDK overview

This repository is the Frontal SDK for Go and contains one Go module. Import the root `frontal` package to create a unified client with service fields such as `AI`, `Agents`, and `Workflows`. Service calls use the contract inventory in `contracts/sdk-endpoints.json`; callers provide normal Go request values and decode responses into their own types.

The transport uses `net/http`, propagates request contexts, authenticates with Bearer tokens, adds request IDs and environment headers, and retries transient GET and HEAD failures. `FetchPage[T]`, `PollUntil[T]`, and `Watch[T]` cover pagination metadata, polling, and cancellable SSE consumption. All network behavior can be exercised with `httptest.Server`.

See the [architecture guide](./ARCHITECTURE.md), [developer setup](./DEVELOPERS.md), and the executable examples in [`example_test.go`](../example_test.go).
