# Go SDK overview

This repository is the Frontal SDK for Go and contains one Go module. Import the root `frontal` package and call `frontal.New(opts ...frontal.Option)` to create a unified client. Its service fields cover all 18 contract groups, including `AI`, `Agents`, `React`, and `Workflows`. Service calls use the contract inventory in `contracts/sdk-endpoints.json`; callers provide ordinary Go request values and decode responses into their own types.

The transport uses `net/http`, propagates request contexts, authenticates with Bearer tokens, adds request IDs and environment headers, and retries transient GET and HEAD failures. `FetchPage[T]`, `PollUntil[T]`, and `Watch[T]` cover pagination metadata, polling, and cancellable SSE consumption. All network behavior can be exercised with `httptest.Server`; no live Frontal account is needed for tests or examples.

See the [architecture guide](./ARCHITECTURE.md), [developer setup](./DEVELOPERS.md), and the executable examples in [`example_test.go`](../example_test.go).
