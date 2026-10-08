# Frontal Go SDK

![Frontal Banner](./banner.png)

**Frontal Go SDK library.**

Frontal's Go SDK is a single Go module. The public client and contract backed endpoint catalog live in `pkg/resources`; authentication, HTTP handling, headers, and URL helpers are separated into focused support packages.

## Package layout

| Path | Purpose |
| --- | --- |
| `pkg/resources/` | Public client, request/endpoint types, generated route catalog, and transport |
| `pkg/authentication/` | API key validation and Bearer authentication |
| `pkg/handlers/` | JSON request/response handling, API errors, and SSE decoding |
| `pkg/headers/` | HTTP header names and defaults |
| `pkg/utils/` | URL, path parameter, timeout, and retry helpers |
| `tests/` | Cross-package integration tests; unit tests remain beside their code |

## Quickstart

Set `FRONTAL_API_KEY`, then call an operation from the generated endpoint catalog:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/frontal-labs/sdk-go/pkg/resources"
)

func main() {
	client, err := resources.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	endpoint, ok := resources.FindEndpoint("agents", "GET", "/agents/health")
	if !ok {
		log.Fatal("agents health endpoint is missing from the SDK catalog")
	}

	var health map[string]any
	err = client.Call(context.Background(), resources.Request{Endpoint: endpoint}, &health)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("health: %v\n", health)
}
```

`Client.Call` supports contract backed operations with path parameters, query values, custom headers, and JSON bodies. `Client.Request` handles paths outside the catalog. Use `Client.NewRequest` and `Client.Do` for raw bodies; `Client.Stream` and `resources.StreamEvents[T]` support Server-Sent Events.

The client uses `https://api.frontal.dev/v1` by default, sends API keys as Bearer tokens, bounds JSON responses to 32 MiB, and retries transient failures only for GET and HEAD requests.

## Configuration

`resources.NewClientFromEnv` reads `FRONTAL_API_KEY`, optional `FRONTAL_API_URL`, and optional `FRONTAL_TIMEOUT`. Timeout values accept Go duration syntax or integer milliseconds. Go does not load `.env` files automatically; [`.env.example`](./.env.example) is a reference only.

Use `resources.NewClient(apiKey, options...)` to configure the client directly. The module path is `github.com/frontal-labs/sdk-go`.

## Templates

The `templates/` package renders Go starter projects for CLI applications, HTTP services, background workers, and SSE consumers. See [`templates/README.md`](./templates/README.md) for details.

## Development

Requirements: Go 1.22 or later.

```bash
gofmt -w ./pkg
go generate ./pkg/resources
go build ./...
go vet ./...
go test -race ./...
python3 scripts/check_contracts.py
```

See [`CONTRIBUTING.md`](./CONTRIBUTING.md), [`docs/ONBOARDING.md`](./docs/ONBOARDING.md), and [`AGENTS.md`](./AGENTS.md).

## License

Apache-2.0. See [`LICENSE.md`](./LICENSE.md).
