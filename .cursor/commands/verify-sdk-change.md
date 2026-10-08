Review the current diff and identify affected modules, public behavior, and relevant contract entries. Run only checks relevant to the changes, choosing from the repository commands below. Do not edit code unless I explicitly ask you to fix a failure. Report checks run and their results.

```sh
gofumpt -w .
goimports -w .
go build ./...
go test -race ./...
golangci-lint run ./...
python3 scripts/check_contracts.py
```
