# orchestrator

## Scope
- Task routing across multiple models.
- Merge strategy for model outputs.
- Artifact generation and task tracking.

## Responsibilities
- Accept TaskSpec and emit TaskStatus updates.
- Route tasks to model drivers based on constraints.
- Merge results into deterministic diffs.

## Deliverables (MVP)
- Router for 2 models.
- Merge with rationale and confidence score.
- Artifact store for diffs and logs.
