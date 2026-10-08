---
name: frontal-sdk-go
description: Guidance for adding or reviewing integrations using the Frontal Go SDK.
---

# Frontal Go SDK

The repository provides a generic Go client and generated endpoint descriptors. Check the committed contracts before using an operation; descriptors provide route shapes, while endpoint-specific convenience methods and typed response models are not generated yet. Follow the idioms and toolchain documented in this repository.

## Configuration

The client reads `FRONTAL_API_KEY` and optional `FRONTAL_API_URL` and `FRONTAL_TIMEOUT` settings. The default URL is `https://api.frontal.dev/v1`; this language does not load `.env` files automatically.

## Modules

See [the architecture guide](../../docs/ARCHITECTURE.md) and [the SDK overview](../../docs/OVERVIEW.md).
