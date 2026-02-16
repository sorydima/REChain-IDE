# Constraints

Explicit limitations for the MVP.

- Model drivers are stubbed.
- Diff merging is deterministic but simplistic.
- Kernel sandbox only allows `echo` by default.
- Web6-3D viewer is a stub and renders a static graph.
- Web6-3D can load a graph from file but does not parse repo structure yet.
- RAG is string match based by default; embeddings can be enabled via env.
- VS Code extension applies diffs with `git apply` only.
- HF driver requires an available Inference Provider for the selected model.
