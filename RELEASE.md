# Releasing the Frontal Go SDK

Tag the repository with `vMAJOR.MINOR.PATCH`. Keep the module import path stable, verify `go list -m` resolves the tagged version, and publish release notes. The Go proxy indexes public tags; no registry upload step is required.

Before release, update `CHANGELOG.md`, confirm the supported Go version range, check the generated contract matrix, and verify package metadata. Publishing automation is not enabled while this repository is a scaffold.
