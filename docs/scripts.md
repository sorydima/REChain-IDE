# Scripts

## Available
- scripts/dev.ps1: start all services
- scripts/stop.ps1: attempt to close service windows
- scripts/status.ps1: check service health
- scripts/demo.ps1: run a demo flow
- scripts/e2e.ps1: run end-to-end scenario with replay modes, replay-batch, replay-chain, debug-compare, RAG tune reset/import/export, and metrics assertions
- scripts/check-go.ps1: verify Go installation

- scripts/license-check.ps1: license check stub

## Quick Runbook
- Start stack: `./scripts/dev.ps1`
- Check health: `./scripts/status.ps1`
- Execute E2E: `./scripts/e2e.ps1`
- Stop stack: `./scripts/stop.ps1`

## API Smoke Examples
- Force soft replay:
  - `Invoke-RestMethod -Method Post -Uri "http://localhost:8081/tasks/<task_id>/replay?mode=force-agent-soft" -ContentType "application/json" -Body "{}"`
- Unified debug payload:
  - `Invoke-RestMethod -Uri "http://localhost:8081/tasks/<task_id>/debug"`
- RAG tune reset + history:
  - `Invoke-RestMethod -Method Post -Uri "http://localhost:8083/search/hybrid-tune/reset" -ContentType "application/json" -Body "{}"`
  - `Invoke-RestMethod -Uri "http://localhost:8083/search/hybrid-tune/history?limit=5"`

