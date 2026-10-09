# Frontal Go SDK

![Frontal Banner](./banner.png)

[![skills.sh](https://skills.sh/b/frontal-labs/sdk-go)](https://skills.sh/frontal-labs/sdk-go)

One context-first Go client for Frontal AI, agents, workflows, and every other service in the committed endpoint inventory.

## Quickstart

With `FRONTAL_API_KEY` set, create one client and call a service operation:

```go
package main

import (
	"context"
	"log"

	frontal "github.com/frontal-labs/sdk-go/v2"
	"github.com/frontal-labs/sdk-go/v2/agents"
)

func main() {
	client, err := frontal.New()
	if err != nil {
		log.Fatal(err)
	}
	endpoint, ok := client.Agents.Endpoint("GET", "/agents/health")
	if !ok {
		log.Fatal("agents health endpoint is unavailable")
	}
	var health struct {
		Status string `json:"status"`
	}
	if err := client.Agents.Call(context.Background(), agents.Request{Endpoint: endpoint}, &health); err != nil {
		log.Fatal(err)
	}
	log.Println(health.Status)
}
```

Import `github.com/frontal-labs/sdk-go/v2/agents` for the service request type. The example in [`example_test.go`](./example_test.go) runs against an `httptest.Server`.

## Client and services

`frontal.New(opts ...frontal.Option)` reads `FRONTAL_API_KEY`, `FRONTAL_API_URL`, `FRONTAL_ENV`, `FRONTAL_DEBUG`, and `FRONTAL_TIMEOUT`. Go does not load `.env` files automatically. `FRONTAL_ENV` defaults to `development`. Options include `WithAPIKey`, `WithBaseURL`, `WithHTTPClient`, `WithTimeout`, `WithMaxRetries`, `WithUserAgent`, `WithLogger`, and `WithMaxResponseBytes`.

The client exposes one service client for each contract group. Each service package owns its endpoint catalog and provides `Endpoints`, `Endpoint`, `Call`, and `Stream`. Use service-scoped calls as the contract-checked API while operation-specific types are added from stable schemas. `client.Request` supports direct HTTP access outside the inventory.

`frontal.BindOperation[Request, Response]` binds a known route to caller-defined types. `frontal.Field[T]`, `frontal.F`, and `frontal.Null[T]()` represent omitted, explicit, and null request fields. `frontal.ExtraFields` preserves response properties not represented in a local struct.

`frontal.FetchPage[T]` decodes collection and cursor metadata. `frontal.Watch[T]` streams JSON SSE events; cancel the context to stop the stream. `frontal.PollUntil[T]` polls with a caller-supplied fetch and completion check.

API failures are `*frontal.APIError` values discoverable through `errors.As`, with status, code, request ID, and retryability. Use `IsAuthError`, `IsRateLimitError`, `IsValidationError`, `IsServerError`, and `IsNetworkError` to classify failures.

## Package layout

| Path | Purpose |
| --- | --- |
| `client.go`, `service.go` | Unified client, generic request binding, pagination, polling, and streams |
| Top-level service packages | Service clients and generated endpoint catalogs |
| `internal/` | Private transport, authentication, response handling, headers, and URL helpers |
| `contracts/` | OpenAPI snapshots and endpoint inventory |

## Development

The module supports Go 1.22 and 1.23. CI uses gofumpt, goimports, golangci-lint (govet, staticcheck, errcheck, revive), and race tests.

```bash
gofumpt -w .
goimports -w .
go generate ./internal/core
go build ./...
golangci-lint run ./...
go test -race ./...
go test -run '^Example' ./...
python3 scripts/check_contracts.py
```

See [`CONTRIBUTING.md`](./CONTRIBUTING.md), [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md), and [`docs/RELEASING.md`](./docs/RELEASING.md).

## Agent skills

Install this repository's Go-specific agent skills with the [skills CLI](https://skills.sh/docs/cli):

```bash
npx skills add frontal-labs/sdk-go
```

## License

Apache-2.0. See [`LICENSE.md`](./LICENSE.md).
