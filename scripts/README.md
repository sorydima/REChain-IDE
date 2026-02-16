# Scripts

## dev
Starts all services in separate PowerShell windows.

## stop
Attempts to close service windows started by dev.

## status
Checks health endpoints for all services.

## demo
Runs a quick end-to-end flow.

## e2e
Runs a fuller scenario:
- waits for orchestrator/rag/quantum/agent-compiler health
- indexes repo data in RAG
- submits a task via VS Code client CLI
- fetches merged result and quality score
- queries orchestrator model registry APIs
- runs a quantum optimize demo and prints metrics URLs

## check-go
Verifies that Go is installed and in PATH.

## doctor
Runs local diagnostics:
- checks Go installation
- checks for UTF-8 BOM in `go.work` and `go.mod`
- optional BOM fix with `-FixBom`
- optional service checks with `-CheckServices`
- optional tests with `-RunTests`

## license-check
Stub for future license checks.

