# CLI Tools

## orchctl
Indexes files into the RAG service.

```powershell
cd rechain-ide/orchestrator/cmd/orchctl
./orchctl -rag http://localhost:8083 -repo demo -files "a.go,b.go"
```

## ragctl
Watches a folder and periodically indexes files.

```powershell
cd rechain-ide/rag/cmd/ragctl
./ragctl -rag http://localhost:8083 -repo demo -root . -interval 10
```

## rechain (orchestrator helper)
Basic CLI for health, submit, status, and metrics.

```powershell
go run rechain-ide/cli/cmd/rechain/main.go -cmd health
go run rechain-ide/cli/cmd/rechain/main.go -cmd submit -input "add logging"
go run rechain-ide/cli/cmd/rechain/main.go -cmd status -task task_123
go run rechain-ide/cli/cmd/rechain/main.go -cmd metrics
```
