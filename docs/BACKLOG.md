# Backlog

## Epics
- E1: Kernel runtime and policy enforcement
- E2: Multi-model orchestration and merge
- E3: IDE surfaces (VS Code and Cursor)
- E4: RAG indexing and retrieval
- E5: 3D spatial graph viewer
- E6: Distributed cache

## Stories
- S1: Define TaskSpec and ModelResult schemas. (Owner: Orchestrator)
- S2: Build minimal model router with 2 drivers. (Owner: Orchestrator)
- S3: Implement patch merge (3-way diff + scoring). (Owner: Orchestrator)
- S4: VS Code extension can submit TaskSpec and apply diff. (Owner: IDE Surfaces)
- S5: Cursor integration can submit TaskSpec and apply diff. (Owner: IDE Surfaces)
- S6: RAG index for a repo with incremental updates. (Owner: Orchestrator)
- S7: Quantum-simulated embeddings for code search. (Owner: Quantum)
- S8: 3D viewer loads repo dependency graph. (Owner: Web6-3D)
- S9: Kernel sandbox isolates execution and logs outputs. (Owner: Kernel)
- S10: Distributed cache with TTL and invalidation. (Owner: Infra)

## Stories
- S11: Schema validation script for payload examples. (Owner: Orchestrator)

## Acceptance criteria (MVP)
- A single user can request a patch and receive a deterministic diff.
- At least 2 models are used with a merge decision and rationale.
- VS Code and Cursor surfaces can apply diffs without manual edits.
- RAG index updates on file change.
- 3D viewer renders a graph for a medium repo (1k+ files).
- Sandbox logs are captured and attached to artifacts.

## Links
- docs/roadmap.md
- docs/ARCHITECTURE.md

