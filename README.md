# Frontal Go SDK

**Frontal client library for Go.** This repository is a single Go module with an authenticated HTTP client, shared request and response helpers, and a generated endpoint catalog sourced from the committed API inventory.

## Go source layout

| Path | Purpose |
| --- | --- |
| `client.go` | Public `frontal.Client`, options, contract-backed calls, retries, and stream entry points |
| `authentication/` | Opaque API key validation and Bearer authentication |
| `headers/` | Shared HTTP header names and defaults |
| `handlers/` | JSON request/response handling and Server-Sent Events decoding |
| `models/` | API errors, request/event types, and generated endpoint descriptors |
| `utils/` | URL, path parameter, timeout, and retry-delay helpers |
| `internal/core/` | Private HTTP transport and default timeout |

## Quickstart

Set `FRONTAL_API_KEY`, then use a descriptor from the generated catalog:

```go
client, err := frontal.NewFromEnvironment()
if err != nil {
    return err
}

endpoint, ok := models.FindEndpoint("agents", "GET", "/agents/health")
if !ok {
    return errors.New("agents health endpoint is missing from the SDK catalog")
}

health, err := frontal.DoJSON[map[string]any](ctx, client, models.Request{
    Endpoint: endpoint,
})
if err != nil {
    return err
}
_ = health
```

The endpoint catalog comes from `contracts/sdk-endpoints.json`. `Client.Call` and `frontal.DoJSON[T]` support positional path parameters, query values, custom headers, and JSON bodies. `Client.Request` is available for paths not represented in the catalog. Use `Client.NewRequest` and `Client.Do` for raw request bodies or streaming responses. Service-specific convenience methods are not generated yet.

The client sends API keys as `Authorization: Bearer ...`, uses `https://api.frontal.dev/v1` by default, bounds JSON responses to 32 MiB, and retries transient failures only for GET and HEAD requests. `Client.Stream` and `frontal.StreamEvents[T]` support Server-Sent Events.

## Configuration

`frontal.NewFromEnvironment` reads `FRONTAL_API_KEY`, optional `FRONTAL_API_URL`, and optional `FRONTAL_TIMEOUT`. Timeout values accept Go duration syntax or integer milliseconds. Go does not load `.env` files automatically; [`.env.example`](./.env.example) is a reference only.

Use `frontal.NewClient(apiKey, options...)` to configure the client directly. The module path is `github.com/frontal-labs/sdk-go`; the module is not published yet.

## Development

Requirements: Go 1.22 or later.

```bash
gofmt -w client.go authentication headers handlers models utils internal
go generate ./models
go build ./...
go vet ./...
go test -race ./...
python3 scripts/check_contracts.py
```

Go tests belong beside the code they exercise in `*_test.go` files. See [`CONTRIBUTING.md`](./CONTRIBUTING.md), [`docs/ONBOARDING.md`](./docs/ONBOARDING.md), and [`AGENTS.md`](./AGENTS.md).

## License

Apache-2.0. See [`LICENSE.md`](./LICENSE.md).
