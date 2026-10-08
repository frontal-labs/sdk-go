---
name: frontal-sdk-go
description: Guidance for adding or reviewing integrations using the Frontal Go SDK.
---

# Frontal Go SDK

This repository is a scaffold. Check the matching Go module README and source before using an operation; an endpoint in `contracts/` does not guarantee a public method exists. Follow the idioms and toolchain documented in this repository.

## Configuration

The client configuration uses `FRONTAL_API_KEY` (`frt_...`) and `FRONTAL_API_URL` (default `https://api.frontal.dev/v1`). See the root `.env.example`; this language does not load `.env` files automatically.

## Modules

See [the architecture guide](../../docs/ARCHITECTURE.md) and package READMEs under `../../pkg/`.
