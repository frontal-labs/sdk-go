# Go onboarding

1. Install Go 1.22 or later.
2. Clone this repository and resolve its dependencies using the Go toolchain.
3. Run the format, lint, build, and test commands in the root README.
4. Review `AGENTS.md`, `docs/ARCHITECTURE.md`, and `contracts/README.md`.
5. Select a service package under `pkg/` and compare its planned operations with `contracts/sdk-endpoints.json`.
6. Add implementation, Go tests, documentation, and examples together.

Install Go 1.22+, then use the root README's native Go commands. Export `FRONTAL_API_KEY` and optional settings in your shell; Go does not load `.env` files by default. The client runtime is not implemented yet, so the scaffold does not make API requests.
