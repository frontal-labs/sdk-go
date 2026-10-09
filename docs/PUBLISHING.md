# Publishing a Go module release

Push a `vMAJOR.MINOR.PATCH` tag to trigger `.github/workflows/publish.yml`. The workflow validates the tag and module path, checks API compatibility against the latest earlier release in the same major version, builds and tests the tagged source, checks the endpoint contract, and uses GoReleaser to publish a GitHub release. The Go module proxy indexes public tags automatically; no separate module upload is required.

The canonical v2 module path is `github.com/frontal-labs/sdk-go/v2`; the release workflow enforces a `/vN` suffix for future major versions. See [RELEASING](./RELEASING.md) for the release checklist and commands.
