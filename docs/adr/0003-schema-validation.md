# ADR-0003: Schema Validation Script

## Status
Accepted

## Date
2026-02-09

## Context
Schema examples can drift from documented requirements. A simple validator helps prevent accidental omissions of required fields such as schema_version.

## Decision
Add a lightweight PowerShell validator at schemas/validate.ps1 and run it in CI.

## Consequences
- Early detection of schema example regressions.
- Validator should stay minimal and fast.
