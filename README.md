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

## Functions

The Functions client shares the root client's Bearer authentication, base URL, timeout, and retry settings. Create functions directly from a `functions.FunctionDefinition`, then invoke them with a JSON object:

```go
package main

import (
	"context"
	"log"
	"os"
	"time"

	frontal "github.com/frontal-labs/sdk-go/v2"
	"github.com/frontal-labs/sdk-go/v2/functions"
)

func main() {
	client, err := frontal.New(
		frontal.WithAPIKey(os.Getenv("FRONTAL_API_KEY")),
		frontal.WithBaseURL("https://api.frontal.dev/v1"),
		frontal.WithTimeout(30*time.Second),
		frontal.WithMaxRetries(3),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	created, err := client.Functions.Create(ctx, functions.FunctionDefinition{
		Name:       "hello",
		Runtime:    functions.RuntimeNodeJS22,
		Entrypoint: "index.handler",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{"name": map[string]any{"type": "string"}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.Functions.Invoke(ctx, functions.FunctionInvocationInput{
		FunctionID: created.ID,
		Input:      map[string]any{"name": "Ada"},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("execution %s result: %#v", result.ExecutionID, result.Result)
}
```

The shared client defaults to a 30-second timeout and three retries for safe GET requests. Configure a different base URL, timeout, or retry count with `WithBaseURL`, `WithTimeout`, and `WithMaxRetries`.

## Client and services

`frontal.New(opts ...frontal.Option)` reads `FRONTAL_API_KEY`, `FRONTAL_API_URL`, `FRONTAL_ENV`, `FRONTAL_DEBUG`, and `FRONTAL_TIMEOUT`. Go does not load `.env` files automatically. `FRONTAL_ENV` defaults to `development`. Options include `WithAPIKey`, `WithBaseURL`, `WithHTTPClient`, `WithTimeout`, `WithMaxRetries`, `WithUserAgent`, `WithLogger`, and `WithMaxResponseBytes`.

The client exposes one service client for each supported API group. Inventory-backed service packages own generated endpoint catalogs and provide `Endpoints`, `Endpoint`, `Call`, and `Stream`; the Functions package also provides typed operations. `client.Request` supports direct HTTP access outside the inventory.

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
