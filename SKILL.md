---
name: frontal-sdk-go
description: Guidance for building integrations with the unified Frontal Go SDK.
---

# Frontal Go SDK

Use the root `frontal` package to create one client with `frontal.New(opts ...frontal.Option)`. The client exposes services such as `AI`, `Agents`, and `Workflows`; calls take `context.Context` first.

```go
client, err := frontal.New(frontal.WithAPIKey(apiKey))
```

Use `client.<Service>.Endpoints()` and `Endpoint` to inspect the committed service inventory. Send a contract operation with `Service.Call(ctx, resources.Request, out)`. Do not invent endpoint paths or response schemas; check `contracts/sdk-endpoints.json` and the OpenAPI snapshots.

For isolated tests, use `net/http/httptest.Server` with `WithBaseURL` and `WithHTTPClient`. Use `FetchPage[T]` for collection/cursor metadata, `PollUntil[T]` for polling, and `Watch[T]` for context-cancelled SSE channels. API failures are typed `*frontal.APIError` values; classify them with the exported error helpers.

Go does not load `.env` files. Configuration comes from `FRONTAL_API_KEY`, `FRONTAL_API_URL`, `FRONTAL_ENV`, `FRONTAL_DEBUG`, and `FRONTAL_TIMEOUT`, or functional options.
