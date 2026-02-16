# Architecture Summary

This is a short summary of the system. For full details, see docs/ARCHITECTURE.md.

## Components
- Orchestrator: routes tasks and merges model outputs.
- Kernel: enforces policy and executes commands in a sandbox.
- RAG: indexes and retrieves repo context.
- Web6-3D: visualizes the code graph.
- IDE surfaces: VS Code and Cursor integrations.

## Flow
1. IDE surface submits a TaskSpec to the orchestrator.
2. Orchestrator fetches context from RAG (optional).
3. Orchestrator calls model drivers and merges results.
4. Kernel executes guarded commands as needed.
5. Artifacts and diffs are returned to the IDE.
