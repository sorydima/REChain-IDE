# Requirements

Non-functional requirements for MVP and beyond.

## Reliability
- Services should start reliably on developer machines.
- Basic health endpoints for monitoring.

## Security
- Default-deny policy for execution.
- Request ID propagation and logging.

## Performance
- Orchestrator responses for task submission within 200ms.
- RAG search returns within 500ms for 1k-file repo.

## Usability
- VS Code command can submit tasks with minimal config.
- Clear docs for setup and runbook.
