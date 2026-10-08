# Frontal Go SDK guide

The SDK provides `resources.Client`, generic JSON and raw HTTP requests, a generated endpoint catalog, and Server-Sent Events helpers. Endpoint-specific convenience methods and typed response models are not generated yet, so use the committed contracts to choose a descriptor and response type.

## Toolchain

Go 1.22 or later. See the root README for build and quality commands.

## Configuration

Set `FRONTAL_API_KEY`. `FRONTAL_API_URL` defaults to `https://api.frontal.dev/v1`, and `FRONTAL_TIMEOUT` accepts a Go duration or integer milliseconds. Go does not read `.env` files automatically.

## Package map

- `pkg/resources/` provides the client, request types, endpoint catalog, and private transport.
- `pkg/authentication/` validates API keys and applies Bearer authentication.
- `pkg/handlers/` handles requests, responses, API errors, and event streams.
- `pkg/headers/` and `pkg/utils/` provide shared HTTP and URL helpers.

See the root README for a client example and `docs/ARCHITECTURE.md` for request flow.
