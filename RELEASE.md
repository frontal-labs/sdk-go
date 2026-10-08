# Releasing the Frontal Go SDK

Tag releases with `vMAJOR.MINOR.PATCH`. The v1 module path remains `github.com/frontal-labs/sdk-go`; major versions v2 and later require a matching `/vN` suffix. Pushing a valid tag runs compatibility checks and GoReleaser. The Go module proxy indexes public tags, so no registry upload step is required.

Before release, update `CHANGELOG.md`, pass CI on Go 1.22 and 1.23, run `apidiff` against the latest release in the same major version, check the contract catalog, build the module, and verify package metadata. The compatibility check is skipped across major versions. The first v1.0.0 release has no prior Go tag to compare.
