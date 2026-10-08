# Contributing to the Frontal Go SDK

## Set up

Install Go 1.22 or later. Follow [`docs/ONBOARDING.md`](./docs/ONBOARDING.md) to prepare the Go toolchain and local environment.

## Make a change

- Keep public client behavior in `pkg/resources` and shared models and helpers in `pkg/authentication`, `pkg/handlers`, `pkg/headers`, and `pkg/utils`. Keep cross-package integration tests under `tests/` and unit tests beside their source.
- Use `resources.Endpoint` descriptors generated from `contracts/sdk-endpoints.json`; do not guess endpoint paths or schemas.
- Keep tests beside the source they exercise in `*_test.go` files.
- Use the native Go formatter, linter, build tool, and test runner.
- Keep contracts and generated reports synchronized when public endpoint coverage changes.
- Add API documentation and a runnable Go example for each public operation.
- Record user-visible changes in `CHANGELOG.md` and use `type(scope): summary` commit subjects.

## Before opening a pull request

Run the commands in the root README and report any contract or public API changes. Do not include live credentials in tests or examples.
