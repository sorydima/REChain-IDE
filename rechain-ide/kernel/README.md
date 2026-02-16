# kernel

## Scope
- Model kernel runtime and driver lifecycle.
- Sandboxing for code execution and tool use.
- Privacy controls and policy enforcement.

## Responsibilities
- Load and isolate model drivers.
- Enforce data-access policies per task.
- Provide a stable Kernel API for orchestrator and agents.

## Interfaces
- Kernel API in docs/ARCHITECTURE.md.

## Deliverables (MVP)
- Basic sandbox with file and network guards.
- Driver registry with 2 model backends.
- Policy engine with allow/deny rules.
