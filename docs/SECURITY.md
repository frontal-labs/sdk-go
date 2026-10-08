# Go security notes

- Load credentials from the process environment or a secret manager; do not commit `.env` values.
- Use TLS for API traffic and avoid logging authorization headers or payloads containing customer data.
- Keep Go dependency and compiler/runtime versions current within the documented support range.
- Use synthetic data in fixtures and deterministic local HTTP mocks in tests.

Report vulnerabilities using the process in the repository root `SECURITY.md`.
