# Releasing the Frontal Go SDK

Merge Conventional Commit changes to `main`. Release Please opens or updates a release PR with the next semver version and generated `CHANGELOG.md` entry. A `feat` change creates a minor release candidate, a `fix` creates a patch candidate, and a breaking change creates a major candidate. Merge the release PR only after its required checks pass; this creates the version tag and GitHub Release. The tag workflow revalidates the source and API compatibility before uploading the source archive, SBOM, and attestations.

The canonical v1 module path is `github.com/frontal-labs/sdk-go`. Future major releases use a matching `/vN` module path suffix. Go's module proxy indexes public tags automatically, so there is no registry upload step or publishing credential. Protect the `release` GitHub Environment with required reviewers. Configure tag and branch rules as described in [`.github/BRANCH_PROTECTION.md`](./.github/BRANCH_PROTECTION.md).
