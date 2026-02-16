# Web6 Graph Format

The Web6-3D viewer reads a JSON graph from `WEB6_GRAPH_PATH`.

## Schema (v0)
Top-level object:
- `nodes`: array of nodes
- `edges`: array of edges

Node:
- `id`: unique string (recommended: repo-relative path)
- `label`: display label (recommended: base name)
- `type`: optional (`file`, `dir`, `pkg`, `unknown`)

Edge:
- `from`: node id
- `to`: node id

Example:
```json
{
  "nodes": [
    { "id": ".", "label": "repo" },
    { "id": "rechain-ide/orchestrator", "label": "orchestrator" }
  ],
  "edges": [
    { "from": ".", "to": "rechain-ide/orchestrator" }
  ]
}
```

Notes:
- Extra fields are ignored by the current viewer.
- Use `scripts/gen-graph.ps1` to generate a starter graph.
- Set `WEB6_GRAPH_MODE=imports` to build a live import graph (Go/JS/TS) without a file.
- Set `WEB6_GRAPH_MODE=imports_go_list` to build a Go dependency graph via `go list`.
