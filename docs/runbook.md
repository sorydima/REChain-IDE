# Runbook

## Local dev

Start services:

```powershell
./scripts/dev.ps1
```

Use import graph mode:
```powershell
./scripts/dev.ps1 -ImportGraph
```

Use Go list graph mode:
```powershell
./scripts/dev.ps1 -GoListGraph
```
Requires Go installed and available in PATH.

Ports:
- Orchestrator: 8081
- Kernel: 8082
- RAG: 8083
- Web6-3D: 8084
- Quantum: 8085
- Agent Compiler: 8086

Environment:
- RAG_URL: Orchestrator uses this to enrich context (e.g., http://localhost:8083)
- ORCH_URL: Web6-3D uses this to fetch queue depth (e.g., http://localhost:8081)
- AGENT_COMPILER_URL: Orchestrator can use this for multi-agent compile (e.g., http://localhost:8086)
- QUANTUM_URL: Orchestrator can use this for quantum routing (e.g., http://localhost:8085)
- WEB6_GRAPH_PATH: Web6-3D graph JSON path (optional)
- WEB6_ROOT: root path for on-the-fly graph generation (default ".")
- WEB6_MAX_NODES: max nodes for on-the-fly graph (default 2000)
- WEB6_GRAPH_TTL_SEC: cache TTL for graph generation (default 5)
- WEB6_GRAPH_MODE: graph mode "tree" or "imports" (default "tree")
- WEB6_GRAPH_MODE supports "imports_go_list" for Go dependency graph
- Go list graph cache hits/misses exposed in `/metrics`
- WEB6_GRAPH_DEPTH_DEFAULT: default context depth (default 1)
- WEB6_COLOR_UNKNOWN: color for unknown type (default #999)
- WEB6_COLOR_MATCH_DIR: match color for dir nodes (default #f5a623)
- WEB6_COLOR_MATCH_FILE: match color for file nodes (default #f08c6a)
- WEB6_COLOR_MATCH_PKG: match color for pkg nodes (default #f1b54a)
- WEB6_COLOR_MATCH_UNKNOWN: match color for unknown nodes (default #f6c08a)
- WEB6_COLOR_MATCH_DEFAULT: fallback match color (default #f39c12)
- QUANTUM_WEIGHT_COST: weight for cost in quantum optimizer (default 0.3)
- QUANTUM_WEIGHT_LATENCY: weight for latency in quantum optimizer (default 0.5)
- QUANTUM_WEIGHT_QUALITY: weight for quality in quantum optimizer (default 0.2)
- AGENT_SCORE_WEIGHT_SIZE: weight for diff size in agent scorer (default 0.4)
- AGENT_SCORE_WEIGHT_CHURN: weight for diff churn in agent scorer (default 0.3)
- AGENT_SCORE_WEIGHT_ERRORS: weight for error tokens in agent scorer (default 0.3)
- ORCH_WORKERS: number of worker goroutines (default 4)
- ORCH_QUEUE_SIZE: queue size per priority (default 200)
- RAG_EMBED_INDEX: enable embedding-based chunk index (default false)
- RAG_EMBED_MAX_CHUNKS: max chunks to embed per index (default 500)
- RAG search mode: `mode=lexical|semantic|hybrid` (default hybrid)
- RAG_WEIGHT_LEXICAL: lexical score weight (default 0.6)
- RAG_WEIGHT_SEMANTIC: semantic score weight (default 0.4)
- RAG_TEMPERATURE: score temperature for hybrid blending (default 1.0)
- RAG weights/temperature exported in RAG `/metrics`
- RAG weights are auto-normalized to sum to 1.0 (fallback 0.5/0.5 if zero)
- Web6-3D filter metrics exported in `/metrics` (matches/edges)
- Web6-3D depth usage exported in `/metrics` with `depth` label
- HF_MODEL_ID: HuggingFace model id for driver
- HF_API_URL: HuggingFace base URL (default https://router.huggingface.co/hf-inference/models)
- HF_TOKEN: HuggingFace API token
- HF_WAIT_FOR_MODEL: true to wait for model loading
- HF_USE_CACHE: false to disable cache
- HF_TIMEOUT_MS: timeout in ms for HF requests (default 8000)
- HF_FALLBACK_MODELS: comma-separated fallback model IDs
- HF_PING_TIMEOUT_MS: model availability ping timeout in ms (default 1500)
- HF_PING_TTL_MS: ping cache TTL in ms (default 15000)
- HF_PING_BACKOFF_MS: initial backoff in ms (default 1000)
- HF_PING_BACKOFF_MAX_MS: max backoff in ms (default 10000)
- HF_PING_INTERVAL_MS: background ping interval (default 60000)
- RAG_CACHE_METRICS_URL: base URL for cache metrics (e.g., http://localhost:8083)
- KERNEL_ALLOWLIST: comma-separated commands (default: echo)
- KERNEL_DENYLIST: comma-separated commands to deny
- KERNEL_MAX_TIMEOUT_MS: max timeout in ms (default: 5000)
- RAG_EMBEDDING_MODEL: model id for embeddings (default sentence-transformers/all-MiniLM-L6-v2)
- RAG_EMBEDDING_URL: embedding base URL (default https://router.huggingface.co/hf-inference/models)
- RAG_EMBEDDING_TOKEN: token for embedding endpoint
- RAG_EMBEDDING_TIMEOUT_MS: embedding timeout in ms (default 8000)
- RAG_CACHE_PATH: embedding cache sqlite/bolt path (default .rag-cache/embeddings.db)
- RAG_CACHE_TTL_SEC: embedding cache TTL in seconds (default 86400)
- RAG_CACHE_PURGE_INTERVAL_SEC: cache purge interval in seconds (default 300)
- RAG_CACHE_MAX_ENTRIES: max cache entries (default 0 = unlimited)
- RAG_CACHE_MAX_BYTES: max cache size in bytes (default 0 = unlimited)
- RAG_MAX_FILE_BYTES: max file size for chunking (default 200000)
- RAG_CHUNK_LINES: lines per chunk (default 40)

## Quick check
- GET http://localhost:8081/health
- GET http://localhost:8082/health
- GET http://localhost:8083/health
- GET http://localhost:8084/health

## Web6 graph
- Run `./scripts/gen-graph.ps1 -Root . -Out .web6-graph.json`
- Set `WEB6_GRAPH_PATH=.web6-graph.json` before starting Web6-3D

## Requirements
- Go 1.21+ is required to run the services.


## Troubleshooting
- If a service fails to start, check that Go is installed and ports 8081-8084 are free.
- Run ./scripts/status.ps1 to verify health endpoints.

