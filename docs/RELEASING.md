# Go release checklist

Tag the repository with `vMAJOR.MINOR.PATCH`. Keep the module import path stable, verify `go list -m` resolves the tagged version, and publish release notes. The Go proxy indexes public tags; no registry upload step is required.

Before publishing, run the Go CI checks, update the changelog and package metadata, review `contracts/reports/migration-matrix.md`, and verify the artifact contents.
