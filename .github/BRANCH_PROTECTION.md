# Branch protection baseline

# GitHub repository settings

In **Settings → Rules → Rulesets**, create an active branch ruleset targeting the default branch (`main`):

- Require a pull request before merging, with at least one approval.
- Require approval of the most recent push and dismiss stale approvals.
- Require all review conversations to be resolved.
- Require these status checks: `CI (ubuntu-latest, Go 1.22.x)`, `CI (ubuntu-latest, Go 1.23.x)`, `CI (macos-latest, Go 1.22.x)`, `CI (macos-latest, Go 1.23.x)`, `CI (windows-latest, Go 1.22.x)`, `CI (windows-latest, Go 1.23.x)`, `Conventional Commits`, `analyze` (CodeQL), and `Dependency Review`.
- Require linear history; block force pushes and branch deletion.
- Do not allow bypass except for a documented break-glass administrator.

Create a tag ruleset targeting `v*` and restrict creation, updates, and deletion to release maintainers. Release Please needs permission to create tags/releases and update its release PR, so allow the GitHub Actions bot for that ruleset or use a narrowly scoped release maintainer workflow.

In **Settings → Code security and analysis**, enable dependency graph, Dependabot alerts and security updates, secret scanning, and push protection. Enable private vulnerability reporting so the security address remains a backup reporting path.

In **Settings → Actions → General**, set the default `GITHUB_TOKEN` permission to read-only and require full-length SHA pins for third-party actions where available. Permit the repository's required actions. Enable artifact attestations for public releases.

Create a protected **release** environment under **Settings → Environments**. Require at least one reviewer, prevent self-review, and restrict deployments to `main` and `v*` tags. The current workflow does not publish to an external package registry, so no registry credential is needed.
