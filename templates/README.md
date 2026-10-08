# Go project templates

These Go `text/template` starters use `pkg/resources` and cover common project setups:

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

`templates.Available()` lists the template names. `Render` creates a new destination directory and writes `go.mod`, `main.go`, and `README.md`. The generated `go.mod` uses a local `replace` directive so the project builds against this checkout before the SDK is published. Replace it with a released SDK version when moving the project elsewhere.
