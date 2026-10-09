# Go API contracts

The JSON and OpenAPI files are shared Frontal API inputs. `sdk-endpoints.json` lists service operations and `coverage-floor.json` holds the coverage baseline. Do not edit generated snapshots by hand.

`go generate ./internal/core` generates the internal route catalog and top-level service endpoint packages. Operation-specific request and response types are added only when authoritative contract schemas define their wire shape; generic service `Call` methods remain the fallback for other routes.
