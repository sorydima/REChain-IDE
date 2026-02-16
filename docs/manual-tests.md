# Manual Test Steps

## Orchestrator
1. Start services: `./scripts/dev.ps1`
2. Submit a task:
   - Use requests/orchestrator.http
3. Check status:
   - GET /tasks/<id>
4. Fetch result:
   - GET /tasks/<id>/result

## Kernel
1. POST /run with an allowed command (echo)
2. Confirm response includes `exit_code` and `output`

## RAG
1. POST /index with file list
2. GET /search?q=... and check matches

## Web6-3D
1. Open http://localhost:8084
2. Verify /graph returns JSON

## VS Code Extension
1. Build: `npm install` and `npm run compile`
2. Run command `REChain: Submit Task`
3. Confirm diff applies or opens in editor
