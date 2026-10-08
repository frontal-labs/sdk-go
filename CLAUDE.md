@AGENTS.md

# Go SDK repository instructions

Read `AGENTS.md` first for repository layout, project conventions, and contract rules. Read `README.md` and the relevant design docs before changing public SDK behavior.

## Contract and compatibility

Use context.Context for request cancellation, wrap errors with %w, keep credentials out of URLs/logs/errors, and use the committed endpoint inventory and OpenAPI snapshots as contract authority.

Do not add public endpoints, wire fields, or behavior unsupported by the committed contracts. Keep credentials, tokens, and customer data out of source, logs, fixtures, and examples. Make the smallest compatible change and update documentation/examples when public behavior changes.

## Skills

Use the matching skill in `.claude/skills/` when its topic applies. Skills are also linked from `.agents/skills/`; source files live in `skills/` and selected upstream sources are recorded in `skills/SOURCES.md`.

## Verification commands

Run only the commands relevant to the change; do not claim verification unless it was run.

```sh
gofumpt -w .
goimports -w .
go build ./...
go test -race ./...
golangci-lint run ./...
python3 scripts/check_contracts.py
```
