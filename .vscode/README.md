# VS Code workspace setup

This folder contains portable workspace settings, extension recommendations, and manually invoked tasks. Open the repository root as a folder in VS Code; no machine-specific paths are configured. Tasks do not run automatically.

Go formatting uses gopls with gofumpt enabled; import organization is delegated to the Go extension. The lint task requires golangci-lint installed locally.

The repository's `.editorconfig` remains authoritative for whitespace and line endings. See `AGENTS.md` for the full development workflow.
