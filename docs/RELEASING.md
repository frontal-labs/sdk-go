# Go release checklist

This module uses an unchanged import path, so release tags use `vMAJOR.MINOR.PATCH` (no `/v2` suffix before major version 2). Pushing a tag triggers GoReleaser to publish the GitHub release; the Go module proxy indexes the public tag automatically.

Before a release, update `CHANGELOG.md`, pass CI on Go 1.22 and 1.23, and check the public API against the latest release with `apidiff`. `v1.0.0` is the first release and has no prior Go release tag to compare.

```bash
git tag -s v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

The tag is the publication point for Go consumers. Run `go list -m github.com/frontal-labs/sdk-go@v1.0.0` after the proxy indexes it. See [PUBLISHING](./PUBLISHING.md) for repository permissions and release ownership.
