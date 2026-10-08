# Contributing to the Frontal Go SDK

## Set up

Install Go 1.22 or 1.23 and the format, lint, and hook tools listed in [`docs/ONBOARDING.md`](./docs/ONBOARDING.md). Node.js and npm are only needed for the local commit-message hook.

## Make a change

- Keep the public unified client in the root `frontal` package and shared transport behavior in `pkg/resources`. Keep auth and HTTP helpers in their matching support packages.
- Use `resources.Endpoint` descriptors generated from `contracts/sdk-endpoints.json`; do not guess endpoint paths or schemas.
- Keep tests beside the source they exercise in `*_test.go` files.
- Use gofumpt, goimports, golangci-lint (govet, staticcheck, errcheck, revive), `go build`, and `go test -race`.
- Keep contracts and generated reports synchronized when public endpoint coverage changes.
- Add API documentation and runnable Go examples for public README usage snippets.
- Record user-visible changes in `CHANGELOG.md` and use `type(scope): summary` commit subjects.
- Install the Git hooks with `lefthook install`; commit subjects follow Conventional Commits.
- Use Conventional Commit subjects (`feat:`, `fix:`, `docs:`, `chore:`, etc.). `feat` creates a minor release candidate and `fix` creates a patch release candidate; mark breaking changes with `!` or a `BREAKING CHANGE:` footer.

## Before opening a pull request

Run the commands in the root README and report any contract or public API changes. Do not include live credentials in tests or examples.
