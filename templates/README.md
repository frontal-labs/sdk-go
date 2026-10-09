# Go project templates

These Go `text/template` starters use the unified `frontal.Client`, its contract-backed service fields, and `pkg/resources.Request` values. They cover common project setups:

- `cli` — one-shot command using a generated endpoint descriptor.
- `http-service` — Go HTTP service backed by the Frontal API.
- `worker` — cancellable background polling loop.
- `stream-consumer` — cancellable Server-Sent Events consumer.

Render a project from Go with the `templates` package:

```go
err := templates.Render(
	"http-service",
	"example.com/acme/service",
	"../service",
	"../sdk-go", // path from the generated project to this SDK checkout
)
```

`templates.Available()` lists the template names. `Render` creates a new destination directory and writes `go.mod`, `main.go`, and `README.md`. The `sdkPath` argument must be a relative path from the generated project directory to this SDK checkout. The generated `go.mod` uses a local `replace` directive so the project can build against this checkout. Remove that directive and pin the released SDK version (for example, `v1.0.0`) when moving the project elsewhere.
