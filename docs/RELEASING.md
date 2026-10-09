# Go release process

Release Please tracks Conventional Commits on `main` and opens a release pull request that updates `CHANGELOG.md`. Merge that PR after required checks pass to create the `vMAJOR.MINOR.PATCH` tag and GitHub Release. The tag workflow validates the module and API compatibility, creates a source archive and SPDX SBOM, and publishes attestations. Go's module proxy indexes the public tag automatically; there is no separate registry upload.

This major release uses the Go module path `github.com/frontal-labs/sdk-go/v2` and version tag `v2.0.0`.  Future major versions use a matching `/vN` suffix. The compatibility check is skipped across major versions. Use `feat` for minor changes, `fix` for patches, and a breaking marker (`!` or `BREAKING CHANGE:`) for major changes.

Protect the `release` environment with required reviewers. See [the repository settings checklist](../.github/BRANCH_PROTECTION.md) for environment, branch, tag, and security settings.
