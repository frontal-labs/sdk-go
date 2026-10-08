# Releasing the Frontal Go SDK

Tag the repository with `vMAJOR.MINOR.PATCH`. Keep the module import path stable, verify `go list -m` resolves the tagged version, and publish release notes. The Go proxy indexes public tags; no registry upload step is required.

Before release, update `CHANGELOG.md`, confirm the supported Go version range, regenerate and check the contract catalog, build the module, and verify package metadata. Protected release automation still needs to be configured.
