# Go SDK architecture

The repository is one Go module. The root `frontal` package provides a unified client and service namespaces. `pkg/resources` owns the shared HTTP transport, request types, and generated route catalog; authentication, HTTP handling, headers, and URL utilities stay in focused packages.

## Package layout

- `client.go` configures the client from functional options and `FRONTAL_*` environment variables, then exposes one service field for each contract group: `AI`, `Agents`, `Audit`, `Auth`, `Billing`, `Blob`, `Connectors`, `Data`, `Governance`, `Lineage`, `Observability`, `Ontology`, `Pipelines`, `React`, `Sandbox`, `Schedules`, `Webhooks`, and `Workflows`.
- `service.go` scopes endpoint calls to each contract service and provides generic page, polling, and SSE helpers.
- `pkg/resources` owns request/endpoint types, the shared transport, retry policy, and generated endpoint catalog.
- `pkg/authentication` validates API keys and applies Bearer authentication.
- `pkg/handlers` builds requests, bounds JSON response decoding, maps API errors, and decodes Server-Sent Events.
- `pkg/headers` defines common HTTP header names and defaults.
- `pkg/utils` validates and joins URLs, expands path parameters, parses timeouts, and reads `Retry-After` values.
- `contracts/` holds the committed endpoint inventory and OpenAPI snapshots.

`contracts/sdk-endpoints.json` generates `pkg/resources/endpoints_generated.go`. Each service exposes a copied endpoint list and a `Call(ctx, resources.Request, out)` method. The call validates that the operation belongs to the selected service before dispatch.

## Request flow

`Application → frontal.Client → Service → resources.Client → authentication + headers → net/http → Frontal API`

`Service.Call` sends an inventory-backed operation with positional path parameters, query values, custom headers, and a JSON body. `Client.Request` and `Client.Core` support direct requests outside the inventory. `Service.Stream` returns an open SSE response body that the caller must close; `Watch[T]` decodes it into a receive-only channel and closes the body when the context is canceled or the stream ends.

The default base URL is `https://api.frontal.dev/v1`. Paths in the endpoint inventory omit `/v1`; URL joining accepts either form without duplicating the version prefix.

## Configuration and errors

`frontal.New` reads `FRONTAL_API_KEY`, `FRONTAL_API_URL`, `FRONTAL_ENV`, `FRONTAL_DEBUG`, and `FRONTAL_TIMEOUT`. API keys are sent only as Bearer authorization. `FRONTAL_TIMEOUT` accepts a Go duration (such as `15s`) or an integer number of milliseconds. Each HTTP attempt carries an `X-Request-ID`, `X-Frontal-Environment`, and SDK version header. Debug logs contain request metadata, not credentials or bodies.

Non-success responses become `*frontal.APIError` values with status, code, request ID, and retryability. Error category helpers classify authentication, rate-limit, validation, and server failures. Only safe GET and HEAD requests retry transient statuses or network timeouts, using exponential backoff and `Retry-After` where available.

Regular responses have a 30-second default timeout and are bounded to 32 MiB of JSON. Streams do not use a client-wide timeout; their request context controls cancellation. `Client.Request` consumes and closes the response body. `Client.Core.Do` and `Service.Stream` return bodies that callers must close.
