# Frontal Go SDK guide

Create one context-first client with `frontal.New(opts ...frontal.Option)`. Its service fields (`client.AI`, `client.Agents`, and the other contract groups) expose the generated operation catalog and `Call` method. Use the endpoint inventory to choose an operation, then pass a `resources.Request` and decode its response into a Go type. The SDK does not invent endpoint-specific request or response schemas.

## Toolchain

Go 1.22 or 1.23. See the root README for build and quality commands.

## Configuration

Set `FRONTAL_API_KEY`. `FRONTAL_API_URL` defaults to `https://api.frontal.dev/v1`; `FRONTAL_ENV` defaults to `development`; `FRONTAL_TIMEOUT` accepts a Go duration or integer milliseconds; and `FRONTAL_DEBUG` accepts `true`, `false`, `1`, or `0`. Go does not read `.env` files automatically. You can override settings with `frontal.WithAPIKey`, `frontal.WithBaseURL`, `frontal.WithHTTPClient`, `frontal.WithTimeout`, and `frontal.WithMaxRetries`.

The code fragments below assume `ctx` is a `context.Context`, `client` is an initialized `*frontal.Client`, and the relevant standard library and SDK packages are imported.

## Call an operation

```go
endpoint, ok := client.Agents.Endpoint(http.MethodGet, "/agents/health")
if !ok {
	return errors.New("agents health operation is missing")
}
var health map[string]any
err := client.Agents.Call(ctx, resources.Request{Endpoint: endpoint}, &health)
```

The path and method must match a committed contract operation. Put path values in `resources.Request.PathParams` in the order shown by the operation path, and pass query values through `resources.Request.Query`.

## Handle errors

```go
var apiErr *frontal.APIError
if errors.As(err, &apiErr) {
	log.Printf("API status=%d code=%s request_id=%s", apiErr.StatusCode, apiErr.Code, apiErr.RequestID)
}
```

Use `frontal.IsAuthError`, `frontal.IsRateLimitError`, `frontal.IsValidationError`, `frontal.IsServerError`, and `frontal.IsNetworkError` to classify errors.

## Pagination and streams

`frontal.FetchPage[T]` decodes a response collection and common cursor metadata. Pass its `NextCursor` back in the request query for the endpoint's cursor parameter. `frontal.Watch[T]` returns a receive-only channel of events and terminal errors; cancel the context to stop the stream. `frontal.PollUntil[T]` performs an immediate fetch and repeats at the supplied interval until the completion check succeeds or the context ends.

## Package map

- The root package provides `frontal.Client`, service namespaces, configuration options, pagination, polling, and watch helpers.
- `pkg/resources/` provides request types, the endpoint catalog, and shared transport.
- `pkg/authentication/` validates API keys and applies Bearer authentication.
- `pkg/handlers/` handles requests, responses, API errors, and event streams.
- `pkg/headers/` and `pkg/utils/` provide shared HTTP and URL helpers.

See the runnable examples in [`example_test.go`](../example_test.go), the root README for a quickstart, and [`docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) for request flow.
