---
paths:
  - "**/*.go"
---

# Go changes

Use context.Context for request cancellation, wrap errors with %w, keep credentials out of URLs/logs/errors, and use the committed endpoint inventory and OpenAPI snapshots as contract authority.

Follow the repository's `AGENTS.md` and `CLAUDE.md`. Use the relevant installed language skill under `.claude/skills/`. Add deterministic, offline tests for behavior changes, and run the repository's relevant format, lint, type/build, and contract checks.
