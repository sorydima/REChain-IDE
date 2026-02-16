# ADR-0002: API Schema Versioning

## Status
Accepted

## Date
2026-02-09

## Context
As the system evolves, payloads will change. Without a versioning policy, integrations will break or require tight coupling between services.

## Decision
Adopt semantic versioning for API schemas and embed a top-level "schema_version" field in all payloads. Increment rules:
- MAJOR: breaking changes to required fields or data types.
- MINOR: backward-compatible additions.
- PATCH: clarifications and non-functional changes.

## Consequences
- Early payloads should include a "schema_version" even if it is "0.1.0".
- Services must validate the schema_version and reject unsupported versions.
