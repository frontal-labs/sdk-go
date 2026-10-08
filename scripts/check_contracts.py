#!/usr/bin/env python3
"""Validate that committed OpenAPI and endpoint contract snapshots are readable."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
FILES = [
    ROOT / "contracts/sdk-endpoints.json",
    ROOT / "contracts/coverage-floor.json",
    ROOT / "contracts/openapi/api.openapi.json",
    ROOT / "contracts/openapi/ai.openapi.generated.json",
    ROOT / "contracts/openapi/manifest.json",
]
for path in FILES:
    if not path.is_file():
        raise SystemExit(f"missing contract file: {path.relative_to(ROOT)}")
    with path.open(encoding="utf-8") as stream:
        json.load(stream)
print(f"validated {len(FILES)} contract snapshots")
