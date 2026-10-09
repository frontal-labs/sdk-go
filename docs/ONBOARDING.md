# Go onboarding

1. Install Go 1.22 or 1.23.
2. Clone this repository and inspect the root module.
3. Set `FRONTAL_API_KEY` and optional `FRONTAL_API_URL`, `FRONTAL_ENV`, `FRONTAL_DEBUG`, and `FRONTAL_TIMEOUT`; Go does not load `.env` files.
4. Import `github.com/frontal-labs/sdk-go/v2` as `frontal` and create a client with `frontal.New()`.
5. Import the service package (for example, `agents`), then inspect `client.Agents.Endpoints()` or find an operation with `client.Agents.Endpoint("GET", "/agents/health")`.
6. Call the endpoint with that package's request type, such as `client.Agents.Call(ctx, agents.Request{Endpoint: endpoint}, &result)`.

For deterministic tests, use `WithBaseURL` with an `httptest.Server` and pass its client with `WithHTTPClient`. See [`example_test.go`](../example_test.go).
