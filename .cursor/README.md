# Cursor project configuration

This directory uses Cursor Project Rules in `.cursor/rules/*.mdc` and project commands in `.cursor/commands/*.md`. The shared rule applies to every task; the language rule is auto-attached to `**/*.go` files. The commands provide reusable review and verification workflows. root `.cursorignore` excludes generated output, dependency caches, and local secrets from Cursor context.

Cursor Agent also reads root `AGENTS.md` and `CLAUDE.md`; Claude Code reads `CLAUDE.md`, which imports `AGENTS.md`. Project instruction sources are therefore shared without relying on the obsolete `.cursor.json` format.
