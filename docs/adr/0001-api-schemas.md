# ADR-0001: Canonical API Schemas in JSON

## Status
Accepted

## Date
2026-02-09

## Context
The system requires stable, interop-friendly contracts between IDE surfaces, the orchestrator, the kernel, and model drivers. Early integration work will move faster if all components align on a single canonical payload format.

## Decision
Use JSON as the canonical schema representation for TaskSpec, TaskStatus, ModelResult, MergeResult, and Artifact. These schemas live in docs/ARCHITECTURE.md and are the source of truth for early integrations.

## Alternatives considered
- Protobuf-first schemas: rejected for early phase due to heavier tooling and slower iteration.
- Ad hoc payloads per component: rejected due to integration friction and inconsistent fields.

## Consequences
- Faster integration between components during MVP.
- Later migration to protobuf or OpenAPI is possible if required for performance or tooling.
