# Examples

This walkthrough demonstrates an end-to-end flow.

## 1) Start services

```powershell
./scripts/dev.ps1
```

## 2) Index a repo with RAG

```powershell
cd rechain-ide/rag/cmd/ragctl
./ragctl -rag http://localhost:8083 -repo demo -root . -interval 30
```

## 3) Submit a task

```powershell
curl -X POST http://localhost:8081/tasks \
  -H "Content-Type: application/json" \
  -d '{"schema_version":"0.1.0","type":"patch","input":"add logging","context":[],"constraints":[],"metadata":{"requester":"cli","priority":"normal"}}'
```

Copy the returned task id, then:

```powershell
curl http://localhost:8081/tasks/<task_id>
curl http://localhost:8081/tasks/<task_id>/result
```

## 4) View 3D graph stub

Open http://localhost:8084 in a browser.
