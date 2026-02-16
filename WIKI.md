# REChain Quantum-CrossAI IDE Engine - Complete Wiki

## Table of Contents

1. [Introduction](#introduction)
2. [Quick Reference](#quick-reference)
3. [Project Overview](#project-overview)
4. [Architecture](#architecture)
5. [Getting Started](#getting-started)
6. [Module Reference](#module-reference)
7. [API Reference](#api-reference)
8. [Development Guide](#development-guide)
9. [Operations & Deployment](#operations--deployment)
10. [Security & Compliance](#security--compliance)
11. [Troubleshooting](#troubleshooting)
12. [Glossary](#glossary)

---

## Introduction

**REChain Quantum-CrossAI IDE Engine** is a next-generation integrated development environment that combines quantum computing capabilities with cross-AI model orchestration to provide an unparalleled development experience.

### Key Features

- **Multi-Model AI Orchestration**: Seamlessly routes tasks across multiple AI models
- **Quantum Computing Integration**: Leverages quantum algorithms for optimization
- **Retrieval-Augmented Generation (RAG)**: Context-aware code assistance
- **3D Code Visualization**: Web6-powered 3D graph visualization
- **Multi-IDE Support**: VS Code and Cursor integrations
- **Distributed Architecture**: Microservices-based scalable design

---

## Quick Reference

### Essential Commands

| Command | Description |
|---------|-------------|
| `make install` | Install all dependencies |
| `make build` | Build the entire project |
| `make test` | Run all tests |
| `make dev` | Start development environment |
| `make docker-up` | Start with Docker Compose |
| `make lint` | Run linting |
| `make fmt` | Format code |

### Service Ports

| Service | Port | Purpose |
|---------|------|---------|
| Orchestrator | 8081 | Task routing and model orchestration |
| Kernel | 8082 | Policy enforcement and sandbox |
| RAG | 8083 | Retrieval-Augmented Generation |
| Web6-3D | 8084 | 3D visualization engine |
| Quantum | 8085 | Quantum computing interface |
| Agent Compiler | 8086 | AI agent compilation |

### Health Check Endpoints

```bash
curl http://localhost:8081/health  # Orchestrator
curl http://localhost:8082/health  # Kernel
curl http://localhost:8083/health  # RAG
curl http://localhost:8084/health  # Web6-3D
curl http://localhost:8085/health  # Quantum
curl http://localhost:8086/health  # Agent Compiler
```

---

## Project Overview

### What is REChain?

REChain is an open architecture for a multi-model, distributed IDE runtime. It coordinates:
- Model drivers for various AI providers
- Policy-controlled code execution
- RAG-based context retrieval
- IDE surface integrations (VS Code, Cursor)

### Technology Stack

| Layer | Technologies |
|-------|--------------|
| Language | Go 1.21+, TypeScript/JavaScript |
| Runtime | Docker, Kubernetes |
| AI Models | OpenAI, Anthropic, Hugging Face |
| Quantum | Quantum simulators and hardware APIs |
| Storage | BoltDB, distributed cache |
| Monitoring | Prometheus, Grafana |

### Project Structure

```
REChain-IDE/
├── .github/              # GitHub workflows and templates
├── docs/                 # Comprehensive documentation
├── rechain-ide/          # Core modules
│   ├── agents/           # AI agents for code assistance
│   ├── cli/              # Command-line interface
│   ├── cursor-integration/# Cursor editor integration
│   ├── distributed-cache/# High-performance caching
│   ├── kernel/           # Core IDE kernel
│   ├── local-models/     # On-premises AI models
│   ├── orchestrator/     # Cross-AI model orchestrator
│   ├── quantum/          # Quantum computing integration
│   ├── rag/              # RAG system
│   ├── shared/           # Shared libraries
│   ├── vscode-extension/ # VS Code extension
│   ├── web6-3d/          # 3D visualization engine
│   └── windsraf-api/     # API gateway
├── requests/             # HTTP request examples
├── schemas/              # JSON schemas
└── scripts/              # Utility scripts
```

---

## Architecture

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                      IDE Surface Layer                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ VS Code Ext │  │   Cursor    │  │      CLI Interface      │  │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘  │
└─────────┼────────────────┼──────────────────────┼─────────────────┘
          │                │                      │
          └────────────────┴──────────┬───────────┘
                                      │
┌─────────────────────────────────────┼───────────────────────────┐
│                   Orchestrator (8081)│                           │
│  ┌───────────────────────────────────┴─────────────────────┐      │
│  │ Task Queue → Router → Model Drivers → Result Merger   │      │
│  └─────────────────────────────────────────────────────────┘      │
└───────────────────────────┬─────────────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          │               │               │
┌─────────┴────────┐ ┌────┴────┐ ┌───────┴───────┐
│      RAG (8083)   │ │Quantum  │ │  Web6-3D (8084)│
│  Index / Search   │ │(8085)   │ │  3D Graph Vis  │
└─────────┬─────────┘ └────┬────┘ └───────┬───────┘
          │               │               │
┌─────────┴─────────┐ ┌────┴────┐ ┌───────┴────────┐
│   Kernel (8082)   │ │  AI     │ │ Agent Compiler │
│  Policy/Sandbox   │ │ Models  │ │    (8086)      │
└───────────────────┘ └─────────┘ └────────────────┘
```

### Core Components

#### 1. Orchestrator (Port 8081)

The brain of the system. Responsibilities:
- Task routing and queue management
- Multi-model dispatch
- Result merging strategies
- Cost optimization
- Health monitoring

**Key Endpoints:**
- `POST /tasks` - Submit new tasks
- `GET /tasks/{id}` - Check task status
- `GET /tasks/{id}/result` - Get task results
- `GET /models` - List available models
- `GET /dashboard/summary` - System overview

#### 2. Kernel (Port 8082)

Policy enforcement and sandbox execution:
- Security policy validation
- Sandboxed code execution
- Command filtering
- Resource limits

**Key Endpoints:**
- `POST /run` - Execute guarded commands
- `GET /metrics` - Performance metrics

#### 3. RAG - Retrieval-Augmented Generation (Port 8083)

Context-aware intelligence:
- Code repository indexing
- Semantic search
- Lexical search
- Hybrid search with tunable weights
- Embedding generation

**Key Endpoints:**
- `POST /index` - Index repository
- `GET /search?q=...` - Search (auto mode)
- `GET /search/semantic?q=...` - Semantic search
- `GET /search/lexical?q=...` - Lexical search
- `GET /search/hybrid?q=...` - Hybrid search
- `POST /search/hybrid-tune` - Adjust search weights

#### 4. Web6-3D (Port 8084)

3D code visualization:
- Code graph visualization
- Interactive 3D scenes
- Dashboard proxy
- Prometheus metrics

**Key Endpoints:**
- `GET /graph` - Code graph data
- `GET /dashboard-summary` - Dashboard data
- `GET /` - 3D viewer interface

#### 5. Quantum (Port 8085)

Quantum computing integration:
- Quantum algorithm execution
- Optimization problems
- Simulation mode
- Hardware integration

**Key Endpoints:**
- `POST /optimize` - Run quantum optimization
- `GET /metrics` - Quantum metrics

#### 6. Agent Compiler (Port 8086)

AI agent management:
- Agent compilation
- Agent deployment
- Version management

**Key Endpoints:**
- `POST /compile` - Compile agent
- `GET /metrics` - Compiler metrics

### Data Flow

```
1. User submits TaskSpec to Orchestrator
2. Orchestrator queries RAG for context (optional)
3. Orchestrator dispatches to appropriate model drivers
4. Results are merged using configured strategy
5. Kernel executes any guarded commands
6. Artifacts and diffs returned to IDE
```

---

## Getting Started

### Prerequisites

- **Go** 1.21 or higher
- **Node.js** 18 or higher
- **Docker** (optional, for containerized development)
- **Make** (for build automation)

### Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd REChain-IDE
   ```

2. **Install dependencies:**
   ```bash
   make install
   ```

3. **Build the project:**
   ```bash
   make build
   ```

4. **Run tests:**
   ```bash
   make test
   ```

### Development Environment

**Option 1: Local Development**
```bash
make dev
```

**Option 2: Docker Compose**
```bash
make docker-up
```

This starts:
- Orchestrator on port 8081
- Kernel on port 8082
- RAG on port 8083
- Web6-3D on port 8084
- Quantum on port 8085
- Agent Compiler on port 8086

### Verification

Check all services are healthy:
```bash
make health-check
# or manually:
for port in 8081 8082 8083 8084 8085 8086; do
  curl -s http://localhost:$port/health && echo " :$port OK"
done
```

---

## Module Reference

### Orchestrator Module (`rechain-ide/orchestrator/`)

Central task routing and model orchestration.

**Files:**
- `main.go` - Service entry point
- `router.go` - Task routing logic
- `merger.go` - Result merging strategies
- `drivers/` - Model driver implementations
- `models.yaml` - Model registry

**Environment Variables:**
```bash
ORCHESTRATOR_PORT=8081
MODEL_REGISTRY_PATH=/config/models.yaml
RAG_ENDPOINT=http://localhost:8083
KERNEL_ENDPOINT=http://localhost:8082
```

### RAG Module (`rechain-ide/rag/`)

Retrieval-Augmented Generation system.

**Files:**
- `main.go` - Service entry point
- `indexer.go` - Repository indexing
- `search.go` - Search implementations
- `embed.go` - Embedding generation
- `hybrid_tune.go` - Search weight tuning

**Environment Variables:**
```bash
RAG_PORT=8083
RAG_CONFIG_PATH=/data/rag.db
EMBEDDING_MODEL=text-embedding-3-small
```

**Search Modes:**
- `semantic` - Vector similarity search
- `lexical` - Keyword-based search
- `hybrid` - Combined approach with tunable weights

### Kernel Module (`rechain-ide/kernel/`)

Policy enforcement and sandbox execution.

**Files:**
- `main.go` - Service entry point
- `policy.go` - Policy definitions
- `sandbox.go` - Sandbox execution
- `guard.go` - Command filtering

**Environment Variables:**
```bash
KERNEL_PORT=8082
POLICY_PATH=/config/policy.yaml
SANDBOX_ENABLED=true
```

### Web6-3D Module (`rechain-ide/web6-3d/`)

3D visualization engine.

**Files:**
- `main.go` - Service entry point
- `graph.go` - Code graph generation
- `viewer.go` - 3D viewer interface
- `dashboard.go` - Dashboard proxy
- `proxy.go` - Prometheus proxy

**Environment Variables:**
```bash
WEB6_PORT=8084
ORCHESTRATOR_ENDPOINT=http://localhost:8081
GRAPH_DEPTH_DEFAULT=3
```

### Quantum Module (`rechain-ide/quantum/`)

Quantum computing integration.

**Files:**
- `main.go` - Service entry point
- `optimizer.go` - Quantum optimization
- `algorithms.go` - Algorithm implementations

**Environment Variables:**
```bash
QUANTUM_PORT=8085
QUANTUM_SIMULATOR=true
QUANTUM_HARDWARE_ENDPOINT=
```

### VS Code Extension (`rechain-ide/vscode-extension/`)

Visual Studio Code integration.

**Files:**
- `package.json` - Extension manifest
- `extension.ts` - Main extension code
- `commands/` - Command implementations
- `providers/` - Language providers

### CLI Module (`rechain-ide/cli/`)

Command-line interface.

**Commands:**
- `rechain init` - Initialize project
- `rechain build` - Build project
- `rechain test` - Run tests
- `rechain deploy` - Deploy application
- `rechain status` - Check service status

---

## API Reference

### Orchestrator API (Port 8081)

#### Task Management

**Submit Task**
```http
POST /tasks
Content-Type: application/json

{
  "prompt": "Generate a Python function to calculate fibonacci",
  "constraints": {
    "models": ["gpt-4", "claude-3"],
    "routing_weights": {"gpt-4": 0.7, "claude-3": 0.3}
  }
}
```

**Get Task Status**
```http
GET /tasks/{id}
```

**Get Task Result**
```http
GET /tasks/{id}/result
```

**Cancel Task**
```http
POST /tasks/{id}/cancel
```

#### Model Management

**List Models**
```http
GET /models
```

**Get Model Health**
```http
GET /models/health
```

**Get Cost Profile**
```http
GET /models/cost-profile?budget_usd=10.00
```

#### Replay and Debugging

**Replay Task**
```http
POST /tasks/{id}/replay?mode=force-policy
```

Modes:
- `force-policy` - Use policy-based merge
- `force-agent` - Use agent compiler
- `force-agent-soft` - Try agent, fallback to policy

**Get Debug Info**
```http
GET /tasks/{id}/debug
GET /tasks/{id}/debug?format=prom
```

**Get Replay Chain**
```http
GET /tasks/{id}/replay-chain
```

#### Dashboard

**Get Summary**
```http
GET /dashboard/summary
GET /dashboard/summary?format=prom
```

**Get Metrics**
```http
GET /metrics
```

### RAG API (Port 8083)

#### Search

**Semantic Search**
```http
GET /search/semantic?q=authentication middleware&k=10
```

**Lexical Search**
```http
GET /search/lexical?q=user login&k=10
```

**Hybrid Search**
```http
GET /search/hybrid?q=database connection&k=10
```

**Auto Mode Search**
```http
GET /search?q=function to sort array&k=10
```

#### Hybrid Tuning

**Get Current Weights**
```http
GET /search/hybrid-tune
```

**Update Weights**
```http
POST /search/hybrid-tune
Content-Type: application/json

{
  "lexical_weight": 0.3,
  "semantic_weight": 0.7,
  "temperature": 0.5
}
```

**Reset to Defaults**
```http
POST /search/hybrid-tune/reset
```

#### Indexing

**Index Repository**
```http
POST /index
Content-Type: application/json

{
  "repo_path": "/path/to/repo",
  "exclude_patterns": ["*.test.js", "node_modules/"]
}
```

### Web6-3D API (Port 8084)

**Get Code Graph**
```http
GET /graph?q=main&type=file&depth=3
```

Parameters:
- `q` - Filter query
- `type` - Filter type: `file`, `dir`, `pkg`, `unknown`
- `depth` - Context hop depth

**Get Dashboard Summary**
```http
GET /dashboard-summary
GET /dashboard-summary?format=prom
```

**Get Recent Tasks**
```http
GET /tasks/recent?limit=20
GET /tasks/recent?state=completed&merge_source=agent_compiler
```

**Compare Tasks**
```http
GET /debug-compare?id1=task-123&id2=task-456
```

### Kernel API (Port 8082)

**Execute Command**
```http
POST /run
Content-Type: application/json

{
  "command": "go test ./...",
  "working_dir": "/workspace",
  "env": {"GOOS": "linux"}
}
```

### Quantum API (Port 8085)

**Run Optimization**
```http
POST /optimize
Content-Type: application/json

{
  "algorithm": "qaoa",
  "problem": "max-cut",
  "graph": {
    "nodes": [0, 1, 2, 3],
    "edges": [[0,1], [1,2], [2,3], [3,0]]
  },
  "simulation": true
}
```

### Agent Compiler API (Port 8086)

**Compile Agent**
```http
POST /compile
Content-Type: application/json

{
  "agent_id": "code-reviewer",
  "source": "...",
  "version": "1.2.0"
}
```

---

## Development Guide

### Setting Up Development Environment

1. **Install Prerequisites**
   ```bash
   # Go
   winget install Go.Go

   # Node.js
   winget install OpenJS.NodeJS

   # Docker Desktop
   winget install Docker.DockerDesktop
   ```

2. **Configure Git Hooks**
   ```bash
   git config core.hooksPath .githooks
   chmod +x .githooks/*  # On Unix systems
   ```

3. **Set Up Environment Variables**
   Create `.env` file:
   ```bash
   ORCHESTRATOR_PORT=8081
   KERNEL_PORT=8082
   RAG_PORT=8083
   WEB6_PORT=8084
   QUANTUM_PORT=8085
   AGENT_COMPILER_PORT=8086

   # API Keys (for production)
   OPENAI_API_KEY=sk-...
   ANTHROPIC_API_KEY=sk-ant-...
   ```

### Coding Standards

**Go Code:**
- Follow Effective Go guidelines
- Use `gofmt` for formatting
- Add comments for exported functions
- Write unit tests with >80% coverage

**TypeScript/JavaScript:**
- Use ESLint and Prettier
- Follow Angular commit message format
- Use strict TypeScript mode

### Testing

**Run All Tests**
```bash
make test
```

**Run Specific Module Tests**
```bash
cd rechain-ide/orchestrator
go test ./...
```

**Integration Tests**
```bash
make test-integration
```

**Load Tests**
```bash
make test-load
```

### Debugging

**View Service Logs**
```bash
# Docker logs
docker-compose logs -f orchestrator

# Or local logs
./scripts/logs.sh orchestrator
```

**Enable Debug Mode**
```bash
export DEBUG=1
export LOG_LEVEL=debug
make dev
```

### Common Development Tasks

**Add a New Model Driver**

1. Create driver file:
   ```go
   // rechain-ide/orchestrator/drivers/mymodel.go
   package drivers

   type MyModelDriver struct {
       apiKey string
   }

   func (d *MyModelDriver) Execute(ctx context.Context, task TaskSpec) (Result, error) {
       // Implementation
   }
   ```

2. Register in `models.yaml`:
   ```yaml
   models:
     - id: my-model
       driver: mymodel
       capabilities: [code_generation, chat]
       cost_per_1k_tokens: 0.002
   ```

**Add New RAG Index Field**

1. Update `indexer.go`:
   ```go
   type CodeDocument struct {
       // Existing fields
       NewField string `json:"new_field"`
   }
   ```

2. Reindex:
   ```bash
   curl -X POST http://localhost:8083/index \
     -H "Content-Type: application/json" \
     -d '{"repo_path": "/path/to/repo"}'
   ```

---

## Operations & Deployment

### Docker Deployment

**Build Images**
```bash
make docker-build
```

**Start Services**
```bash
make docker-up
```

**View Logs**
```bash
docker-compose logs -f
```

**Scale Services**
```bash
docker-compose up -d --scale orchestrator=3
```

### Kubernetes Deployment

**Apply Manifests**
```bash
kubectl apply -f k8s/
```

**Check Status**
```bash
kubectl get pods -l app=rechain
```

**View Logs**
```bash
kubectl logs -l app=rechain-orchestrator
```

### Monitoring

**Prometheus Metrics**

All services expose `/metrics` endpoint:

```bash
# Orchestrator metrics
curl http://localhost:8081/metrics

# Key metrics:
# - rechain_task_trace_total
# - rechain_task_replay_total
# - rechain_merge_choice_total
# - rechain_forced_agent_fallback_total
```

**Grafana Dashboard**

Import dashboard from `grafana/dashboards/rechain.json`

Key panels:
- Task queue depth
- Model health status
- Response latency
- Merge strategy distribution
- Replay statistics

### Alerting

**Web6 Alert Levels**

| Level | Score | Condition |
|-------|-------|-----------|
| OK | 0 | All systems operational |
| Warn | 1 | Stale data (>60s old) |
| Critical | 2 | Multiple services down |

Check alerts:
```bash
curl http://localhost:8084/dashboard-web6/alerts
```

### Backup and Recovery

**RAG Data Backup**
```bash
# Export hybrid tune config
curl http://localhost:8083/search/hybrid-tune/export > rag-config.json

# Import
curl -X POST http://localhost:8083/search/hybrid-tune/import \
  -H "Content-Type: application/json" \
  -d @rag-config.json
```

**Database Backup**
```bash
# Backup BoltDB
cp /data/rag.db /backups/rag-$(date +%Y%m%d).db
```

### Scaling

**Horizontal Scaling**

Scale stateless services:
```bash
docker-compose up -d --scale orchestrator=5
```

**Vertical Scaling**

Adjust resource limits:
```yaml
services:
  orchestrator:
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
```

---

## Security & Compliance

### Authentication

**API Key Authentication**
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     http://localhost:8081/tasks
```

**JWT Authentication**
```bash
curl -H "Authorization: Bearer JWT_TOKEN" \
     http://localhost:8081/tasks
```

### Request ID Tracking

Include `X-Request-Id` header for request tracing:
```bash
curl -H "X-Request-Id: req-12345" \
     http://localhost:8081/tasks/abc123
```

Response includes the same header for correlation.

### Policy Enforcement

Kernel policies control allowed operations:

```yaml
# policy.yaml
commands:
  allowed:
    - go test ./...
    - go build .
    - npm test
  blocked:
    - rm -rf /
    - sudo *
resources:
  max_cpu_percent: 80
  max_memory_mb: 4096
```

### Security Scanning

```bash
# Run security scan
make security-scan

# Check dependencies
go list -json -m all | nancy sleuth
```

### Compliance

- GDPR data handling in `DATA_MANAGEMENT.md`
- Accessibility in `ACCESSIBILITY_GUIDELINES.md`
- Privacy policy in `docs/privacy.md`

---

## Troubleshooting

### Common Issues

#### Service Won't Start

**Symptom:** `make dev` fails with port already in use

**Solution:**
```bash
# Find and kill process
lsof -i :8081
kill -9 <PID>

# Or use different ports
export ORCHESTRATOR_PORT=9081
make dev
```

#### RAG Search Returns No Results

**Symptom:** Empty search results

**Solution:**
```bash
# Check if indexed
curl http://localhost:8083/health

# Reindex repository
curl -X POST http://localhost:8083/index \
  -d '{"repo_path": "/path/to/repo"}'

# Check search weights
curl http://localhost:8083/search/hybrid-tune
```

#### Model Health Shows Fail

**Symptom:** `GET /models/health` shows `fail` status

**Solution:**
```bash
# Check model configuration
cat rechain-ide/orchestrator/models.yaml

# Verify API keys
env | grep API_KEY

# Check network connectivity
curl -I https://api.openai.com
```

#### High Task Latency

**Symptom:** Tasks taking too long

**Solution:**
```bash
# Check queue depth
curl http://localhost:8081/queue-depth

# View metrics
curl http://localhost:8081/metrics | grep rechain_task_latency

# Adjust routing weights
curl -X POST http://localhost:8081/tasks \
  -d '{"constraints": {"models": ["faster-model"]}}'
```

#### Web6 Dashboard Not Loading

**Symptom:** 3D viewer shows blank page

**Solution:**
```bash
# Clear cache
curl -X POST http://localhost:8084/cache/clear

# Check graph data
curl http://localhost:8084/graph | head

# Verify Web6 health
curl http://localhost:8084/health
```

### Log Locations

| Service | Log Location |
|---------|--------------|
| Orchestrator | `/var/log/rechain/orchestrator.log` |
| Kernel | `/var/log/rechain/kernel.log` |
| RAG | `/var/log/rechain/rag.log` |
| Web6-3D | `/var/log/rechain/web6.log` |
| Quantum | `/var/log/rechain/quantum.log` |

### Debug Mode

Enable verbose logging:
```bash
export DEBUG=1
export LOG_LEVEL=debug
export TRACE_ENABLED=true
```

### Support Contacts

| Issue Type | Contact |
|------------|---------|
| Technical | support@rechain.ai |
| Security | security@rechain.ai |
| Documentation | docs@rechain.ai |

---

## Glossary

| Term | Definition |
|------|------------|
| **Agent Compiler** | Service that compiles and deploys AI agents |
| **Hybrid Search** | Combined lexical and semantic search |
| **Kernel** | Core service for policy enforcement and sandbox execution |
| **Merge Strategy** | Algorithm for combining results from multiple models |
| **Orchestrator** | Central service for task routing and model coordination |
| **RAG** | Retrieval-Augmented Generation - context-aware AI |
| **Replay** | Re-executing a task with different merge strategies |
| **Routing Weights** | Priority values for model selection |
| **TaskSpec** | Specification for a task to be executed |
| **Trace** | Detailed execution log of a task |
| **Web6** | Next-generation web platform with 3D and decentralized features |
| **Quantum Optimization** | Using quantum algorithms for optimization problems |

---

## Additional Resources

### Documentation Index

- `docs/index.md` - Full documentation index
- `docs/ARCHITECTURE.md` - Architecture details
- `docs/api.md` - API reference
- `docs/deployment.md` - Deployment guide
- `docs/SECURITY.md` - Security documentation

### External Links

- Project Repository: https://github.com/rechain/REChain-IDE
- Documentation Site: https://docs.rechain.ai
- API Reference: https://api.rechain.ai/docs
- Community Forum: https://community.rechain.ai

### Contributing

See `CONTRIBUTING.md` for:
- Code of conduct
- Pull request process
- Development workflow
- Commit message format

---

*Last Updated: February 16, 2026*
*Version: 1.0.0*
