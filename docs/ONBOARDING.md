# Go onboarding

1. Install Go 1.22 or later.
2. Clone this repository and use the Go toolchain to inspect or build the root module.
3. Follow the development commands in the root README.
4. Review `AGENTS.md`, `docs/ARCHITECTURE.md`, and `contracts/README.md`.
5. Find an endpoint in `contracts/sdk-endpoints.json`; use `resources.FindEndpoint` and `Client.Call` to invoke it.
6. Add endpoint-specific convenience methods or response models only when they can be grounded in the committed contracts.

Export `FRONTAL_API_KEY` and optional `FRONTAL_API_URL` and `FRONTAL_TIMEOUT` settings in your shell; Go does not load `.env` files by default. `resources.NewFromEnvironment` reads these values from the process environment. Import the client from `github.com/frontal-labs/sdk-go/pkg/resources`.
