# Contributing to the Frontal Go SDK

## Set up

Install Go 1.22 or later. Follow [`docs/ONBOARDING.md`](./docs/ONBOARDING.md) to prepare the Go toolchain and local environment.

## Make a change

- Put shared transport and error behavior in `internal/core`.
- Put the unified client in `pkg/sdk` and each service's models and operations in its matching `pkg/<service>` package.
- Keep tests beside the source they exercise in `*_test.go` files.
- Use the native Go formatter, linter, build tool, and test runner.
- Keep contracts and generated reports synchronized when public endpoint coverage changes.
- Add API documentation and a runnable Go example for each public operation.
- Record user-visible changes in `CHANGELOG.md` and use `type(scope): summary` commit subjects.

## Before opening a pull request

Run the commands in the root README and report any contract or public API changes. Do not include live credentials in tests or examples.
