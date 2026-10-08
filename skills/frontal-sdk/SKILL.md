---
name: frontal-sdk-go
description: Guidance for building integrations with the unified Frontal Go SDK.
---

# Frontal Go SDK

Import the root package `github.com/frontal-labs/sdk-go` as `frontal` and create one context-first client with `frontal.New(opts ...frontal.Option)`. Service fields cover every namespace in `contracts/sdk-endpoints.json`.

Check the committed inventory before selecting an operation. Use `client.<Service>.Endpoint(method, path)` or `.Endpoints()` and dispatch with `.Call(ctx, resources.Request, out)`. Keep response types local to the caller when the contract has no declared SDK model; do not invent schemas.

Test requests with `net/http/httptest.Server`. Use `FetchPage[T]` for page metadata, `PollUntil[T]` for polling, and `Watch[T]` for cancellable JSON SSE streams. `*frontal.APIError` carries status, code, request ID, and retryability.

Go does not load `.env` files. The client reads `FRONTAL_API_KEY`, `FRONTAL_API_URL`, `FRONTAL_ENV`, `FRONTAL_DEBUG`, and `FRONTAL_TIMEOUT`; explicit functional options override environment defaults.
