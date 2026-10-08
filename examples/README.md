# Go examples

The executable examples in [`example_test.go`](../example_test.go) show how to create the unified client and call agent and AI operations. Run them with `go test -run '^Example' ./...`; they use `net/http/httptest.Server`, so they need no API key or live backend. For real requests, set `FRONTAL_API_KEY` and call `frontal.New()`; Go does not load `.env` files automatically.

The SDK also supports deterministic application tests: pass `frontal.WithBaseURL(server.URL)` and `frontal.WithHTTPClient(server.Client())` when creating the client, and have the test server return the response your code expects. See the [SDK guide](./SDKS_GUIDE.md) for request, error, pagination, and stream examples.
