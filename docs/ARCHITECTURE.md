# Go SDK architecture

The repository is one Go module. `pkg/resources` is the public SDK package. Authentication, HTTP handling, headers, and URL/retry utilities remain separate packages so the client can reuse them without coupling the transport to resource types.

## Package layout

- `pkg/resources` owns client configuration, endpoint/request models, the generated endpoint catalog, and private HTTP transport.
- `pkg/authentication` validates API keys and applies Bearer authentication.
- `pkg/headers` defines common HTTP header names and defaults.
- `pkg/handlers` builds JSON requests, decodes bounded JSON responses, converts API errors, and parses JSON Server-Sent Events.
- `pkg/utils` validates and joins URLs, expands path parameters, parses timeouts, and reads `Retry-After` values.
- `contracts/` is the source of truth for route inventory and OpenAPI shapes.

`scripts/generate_endpoints.py` turns `contracts/sdk-endpoints.json` into `pkg/resources/endpoints_generated.go`. Callers find descriptors with `resources.FindEndpoint`, list them with `resources.EndpointsFor`, and execute them through `Client.Call` or `resources.DoJSON[T]`. Endpoint-specific request and response types are not generated because committed schemas remain generic.

## Request flow

`Application → resources.Client → authentication + headers → private HTTP transport → Frontal API`

`Client.Request` accepts a method and path directly. `Client.Call` accepts a contract endpoint and positional path parameters. Both support query values and JSON request bodies; `Client.NewRequest` and `Client.Do` support raw bodies. `Client.Stream` returns an open response for streaming endpoints; `resources.StreamEvents[T]` decodes JSON event data into `handlers.Event[T]`.

The default base URL is `https://api.frontal.dev/v1`. Paths from the endpoint inventory omit `/v1`; paths copied from OpenAPI may include it. URL joining handles both forms without duplicating the version prefix.

## Safety and lifecycle

API keys are sent only in the Authorization header and requests are restricted to the configured API origin. The default client uses a 30-second timeout for regular requests and no client-wide timeout for streams; stream contexts still control cancellation. JSON responses are limited to 32 MiB, error bodies to 1 MiB, and only GET and HEAD requests are retried on transient failures. `Client.Request` consumes and closes response bodies; `Client.Do` and `Client.Stream` return bodies that callers must close.
