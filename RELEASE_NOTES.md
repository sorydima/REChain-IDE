# REChain Quantum-CrossAI IDE Engine v1.0.0 Release Notes

## Release Date
February 16, 2026

## Overview
Initial release of REChain Quantum-CrossAI IDE Engine - a next-generation integrated development environment combining quantum computing capabilities with cross-AI model orchestration.

## What's Included

### Core Services (10 Binaries)

| Binary | Port | Description | Size |
|--------|------|-------------|------|
| rechain-orchestrator | 8081 | Task routing and model orchestration | ~9.2 MB |
| rechain-kernel | 8082 | Policy enforcement and sandbox execution | ~8.6 MB |
| rechain-rag | 8083 | Retrieval-Augmented Generation system | ~9.5 MB |
| rechain-web6 | 8084 | 3D visualization engine | ~10.1 MB |
| rechain-quantum | 8085 | Quantum computing integration | ~8.1 MB |
| rechain-agent-compiler | 8086 | AI agent compilation | ~8.1 MB |
| rechain-orchctl | - | Orchestrator CLI control tool | ~8.1 MB |
| rechain-ragctl | - | RAG CLI control tool | ~8.2 MB |
| rechain | - | Main CLI interface | ~8.1 MB |
| rechain-ide-client | - | IDE client for VS Code/Cursor | ~8.3 MB |

**Total Package Size:** ~88 MB

## Key Features

### Multi-Model AI Orchestration
- Seamless routing across multiple AI model providers
- Configurable routing weights and fallback chains
- Result merging with multiple strategies (policy_merge, agent_compiler)
- Cost-based model selection

### Retrieval-Augmented Generation (RAG)
- Semantic, lexical, and hybrid search modes
- Runtime-tunable search weights
- Persistent configuration with BoltDB
- Repository indexing with context-aware retrieval

### 3D Code Visualization (Web6)
- Interactive 3D code graph visualization
- Real-time dashboard with Prometheus metrics
- Task replay and debugging visualization
- Cross-task comparison tools

### Quantum Computing Integration
- Quantum optimization algorithms
- Simulation and hardware backend support
- Problem-specific algorithm selection

### Security & Policy
- Policy-based command sandboxing
- Kernel-level security enforcement
- Request ID tracking across services
- Health monitoring and alerting

## API Highlights

### Orchestrator API (8081)
```bash
POST /tasks              # Submit tasks
GET /tasks/{id}          # Check status
GET /models              # List models
GET /dashboard/summary   # System overview
POST /tasks/{id}/replay  # Replay with different strategies
```

### RAG API (8083)
```bash
GET /search?q=...        # Auto mode search
GET /search/hybrid?q=... # Hybrid search
POST /search/hybrid-tune # Adjust search weights
POST /index              # Index repository
```

### Web6-3D API (8084)
```bash
GET /graph?q=...         # Code graph data
GET /dashboard-summary   # Dashboard metrics
GET /debug-compare       # Compare task executions
```

## Quick Start

### 1. Start Core Services
```powershell
# Start services in background
Start-Process .\rechain-orchestrator.exe
Start-Process .\rechain-kernel.exe
Start-Process .\rechain-rag.exe
Start-Process .\rechain-web6.exe
Start-Process .\rechain-quantum.exe
Start-Process .\rechain-agent-compiler.exe
```

### 2. Verify Health
```bash
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
```

### 3. Submit First Task
```bash
curl -X POST http://localhost:8081/tasks \
  -H "Content-Type: application/json" \
  -d '{"prompt":"Hello world in Python"}'
```

## Platform Support

| Platform | Architecture | Status |
|----------|--------------|--------|
| Windows | amd64 | Included |
| Linux | amd64 | Build scripts provided |
| macOS | amd64 | Build scripts provided |

## Build from Source

```bash
# Clone repository
git clone https://github.com/rechain-ai/rechain-ide.git
cd rechain-ide

# Install dependencies
make install

# Build all modules
make build

# Run tests
make test

# Package release
make release
```

## Documentation

- **WIKI.md** - Complete project wiki
- **docs/index.md** - Documentation index
- **API_DOCS.md** - API reference
- **TECHNICAL_DOCS.md** - Technical documentation

## Configuration

### Environment Variables
```bash
ORCHESTRATOR_PORT=8081
KERNEL_PORT=8082
RAG_PORT=8083
WEB6_PORT=8084
QUANTUM_PORT=8085
AGENT_COMPILER_PORT=8086
```

### Service Dependencies
```
orchestrator -> kernel, rag
web6-3d -> orchestrator
(all services independent for startup)
```

## Metrics & Monitoring

All services expose Prometheus-compatible metrics at `/metrics`:

| Metric | Description |
|--------|-------------|
| rechain_task_trace_total | Task execution counts |
| rechain_task_replay_total | Replay operation counts |
| rechain_merge_choice_total | Merge strategy usage |
| rechain_rag_hybrid_tune_updates_total | RAG tuning updates |

## Known Limitations

1. **Model Drivers**: Stub implementations included; real model backends require API key configuration
2. **Quantum Backend**: Simulation mode only by default; hardware integration requires additional setup
3. **Sandbox**: Basic policy enforcement; advanced sandboxing in development

## Security Considerations

- API keys not included; configure in environment
- Review policy.yaml before production use
- Enable sandbox mode for untrusted code execution
- Use request IDs for audit trails

## Checksums

SHA-256 checksums for all binaries included in `checksums.sha256`.

Verify:
```bash
sha256sum -c checksums.sha256
```

## Support

- Documentation: https://docs.rechain.ai
- Issues: https://github.com/rechain-ai/rechain-ide/issues
- Community: https://community.rechain.ai
- Email: support@rechain.ai

## License

MIT License - see LICENSE file

## Changelog

### v1.0.0 (2026-02-16)
- Initial release
- 6 core services with full API support
- VS Code and Cursor integration scaffold
- CLI tools for management
- Prometheus metrics integration
- Task replay and debugging capabilities
- Hybrid RAG search with tunable weights
- 3D code visualization engine

---

**Release URL:** https://github.com/rechain-ai/rechain-ide/releases/tag/v1.0.0
