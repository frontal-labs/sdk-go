# Go API contracts

The JSON and OpenAPI files here are shared Frontal API inputs. `sdk-endpoints.json` lists service operations and `coverage-floor.json` holds the project's coverage baseline. Do not edit generated snapshots by hand.

`scripts/generate_endpoints.py` generates the route descriptor catalog in `pkg/resources/endpoints_generated.go`; run it with `go generate ./pkg/resources`. Generic request support does not count as an endpoint-specific typed implementation.
