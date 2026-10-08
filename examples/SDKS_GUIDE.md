# Frontal Go SDK guide

This guide uses Go terms and layout. Client construction and service operations are not available yet, so it intentionally contains no guessed method calls.

## Toolchain

Go 1.22 or later. See the root README for direct build and quality commands.

## Configuration

Set `FRONTAL_API_KEY` to a key beginning with `frt_`. `FRONTAL_API_URL` defaults to `https://api.frontal.dev/v1`. Go does not read `.env` files automatically. Export the variables in your shell or inject them through your deployment environment.

## Package map

- `pkg/sdk` will provide the unified client.
- `internal/core` will hold shared transport, config, errors, retries, and response handling.
- `pkg/testutil` will provide transport-level mocks and fixtures.
- Service packages live under `pkg/`, including `agents`, `ai`, `audit`, `auth`, `billing`, `blob`, `connectors`, `data`, `governance`, `lineage`, `observability`, `ontology`, `pipelines`, `sandbox`, `schedules`, `webhooks`, and `workflows`.

Each package README describes its current status. Public examples will be added when the matching Go methods are implemented.
