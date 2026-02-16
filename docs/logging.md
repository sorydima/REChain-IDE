# Logging

All services emit structured request logs with `X-Request-Id` propagation.

## Request ID
- Client may provide `X-Request-Id` header.
- If missing, the service generates one.
- Responses include `X-Request-Id`.

## Log format
- `rid=<id> method=<method> path=<path> status=<code> dur=<duration>`
