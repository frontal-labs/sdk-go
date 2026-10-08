# Contributing to the Frontal Go SDK

## Set up

Install Go 1.22 or 1.23, gofumpt, goimports, golangci-lint, lefthook, Node.js, and npm. Follow [`docs/ONBOARDING.md`](./docs/ONBOARDING.md) to prepare the local environment.

## Make a change

- Keep the public unified client in the root `frontal` package and shared transport behavior in `pkg/resources`. Keep auth and HTTP helpers in their matching support packages.
- Use `resources.Endpoint` descriptors generated from `contracts/sdk-endpoints.json`; do not guess endpoint paths or schemas.
- Keep tests beside the source they exercise in `*_test.go` files.
- Use gofumpt, goimports, golangci-lint (govet, staticcheck, errcheck, revive), `go build`, and `go test -race`.
- Keep contracts and generated reports synchronized when public endpoint coverage changes.
- Add API documentation and runnable Go examples for public README usage snippets.
- Record user-visible changes in `CHANGELOG.md` and use `type(scope): summary` commit subjects.
- Install the Git hooks with `lefthook install`; commit subjects follow Conventional Commits.

## Before opening a pull request

Run the commands in the root README and report any contract or public API changes. Do not include live credentials in tests or examples.
