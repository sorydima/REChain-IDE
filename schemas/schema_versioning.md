# Schema Versioning

All schema payloads include a top-level "schema_version" field using semantic versioning.

## Rules
- MAJOR: breaking changes to required fields or data types.
- MINOR: backward-compatible additions.
- PATCH: clarifications and non-functional changes.

## Current version
- 0.1.0

## Related ADR
- docs/adr/0002-api-versioning.md
