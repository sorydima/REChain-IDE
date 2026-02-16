# API

This document summarizes the HTTP APIs across services.

## Orchestrator (8081)
- `GET /health`
- `GET /drivers`
- `GET /models`
- `GET /models/health`
- `GET /models/cost-profile?budget_usd=...`
- `GET /dashboard/summary`
- `GET /ping-metrics`
- `GET /metrics`
- `GET /queue-depth`
- `POST /tasks`
- `GET /tasks/{id}`
- `POST /tasks/{id}/cancel`
- `GET /tasks/{id}/result`
- `GET /tasks/{id}/artifacts`
- `GET /tasks/{id}/trace`
- `GET /tasks/latest/trace`
- `GET /tasks/recent?limit=...`
- `POST /tasks/{id}/replay`
- `POST /tasks/{id}/replay/batch`
- `GET /tasks/{id}/replay-chain`
- `GET /tasks/{id}/debug`
- `POST /quality-score`

Notes:
- `/drivers` returns driver IDs and metadata (cost, capabilities).
- `/models` returns model registry entries (driver-backed + HF primary/fallback model IDs).
- `/models/health` returns model availability from ping cache (`ok|fail|stale|unknown`).
- `/models/cost-profile` returns models sorted by cost and optional budget-based selection.
- `/dashboard/summary` returns orchestrator queue/tasks snapshot, models health summary, and key downstream metrics from kernel/rag/quantum/agent-compiler.
- `/dashboard/summary?format=prom` (or `Accept: text/plain`) returns the same summary as Prometheus-compatible metrics for Grafana/Prometheus scrape.
- `/tasks` can include routing weights and fallback models via constraints.
- `/tasks/{id}/trace` returns execution trace: selected models, per-model metrics, merge source, and final merge payload.
- `/tasks/latest/trace` returns the most recent trace (by finished/start timestamp), useful for dashboards.
- `/tasks/recent` returns recent task summaries with state, merge source, and quality score.
- `/tasks/{id}/replay` enqueues a copy of a previous task and links trace via `parent_task_id`.
- `/tasks/{id}/replay?mode=force-agent|force-policy|force-agent-soft` controls replay merge strategy.
- `/tasks/{id}/replay/batch` accepts `{ "modes": ["force-policy","force-agent-soft",...] }` and enqueues multiple replay tasks.
- `/tasks/{id}/replay-chain` returns lineage (ancestors) and descendants for replay debugging.
- `/tasks/{id}/debug` returns consolidated task payload: status, trace, replay-chain, artifacts, and merge metrics.
- `/tasks/{id}/debug?format=prom` (or `Accept: text/plain`) returns Prometheus-compatible task debug gauges for direct Grafana/Prometheus scrape.
- `/tasks/{id}/debug?format=prom&scope=task|global|all` controls whether global merge-choice metrics are included (`all` default).
- `/metrics` now includes `rechain_task_trace_total` with labels:
  - `state="queued|running|completed|failed|canceled|unknown"`
  - `merge_source="agent_compiler|policy_merge|none"`
- `/metrics` also includes replay metrics:
  - `rechain_task_replay_total`
  - `rechain_task_trace_parent_links_total`
- `/metrics` includes merge strategy counters:
  - `rechain_merge_choice_total{source="policy_merge|agent_compiler|..."}`
- `/metrics` includes:
  - `rechain_forced_agent_fallback_total` for `force-agent-soft` fallbacks.
  - `rechain_task_replay_mode_total{mode="..."}`
- `/dashboard/summary` includes replay counters in JSON and Prom format.
- `/dashboard/summary` JSON includes `orchestrator.trace.by_state` and `orchestrator.trace.by_merge_source`.
- `/dashboard/summary` JSON includes `orchestrator.merge_choice` map.
- `/dashboard/summary` JSON includes `orchestrator.replay_modes` map.
- `/dashboard/summary?format=prom` includes `rechain_dashboard_task_replay_mode_total{mode="..."}`.
- `/dashboard/summary?format=prom` includes Web6 downstream gauges:
  - `rechain_dashboard_downstream_up{service="web6"}`
  - `rechain_dashboard_web6_proxy_alert_level`
  - `rechain_dashboard_web6_proxy_json_stale`
  - `rechain_dashboard_web6_proxy_prom_stale`
- `/dashboard/summary` includes `rechain_dashboard_forced_agent_fallback_total` in Prom format.
- `/metrics` includes queue depth, routing counts, latency histogram, and cache metrics.
- `/metrics` also includes routing-by-model counters and per-model latency histograms.

## Kernel (8082)
- `GET /health`
- `POST /run`
- `GET /metrics`

## RAG (8083)
- `GET /health`
- `POST /index`
- `GET /search?q=...&k=...&mode=...`
- `GET /search/lexical?q=...&k=...`
- `GET /search/semantic?q=...&k=...`
- `GET /search/hybrid?q=...&k=...`
- `GET /search/hybrid-tune`
- `POST /search/hybrid-tune`
- `PATCH /search/hybrid-tune`
- `POST /search/hybrid-tune/reset`
- `GET /search/hybrid-tune/history?limit=...`
- `GET /search/hybrid-tune/export`
- `POST /search/hybrid-tune/import`
- `POST /embed`
- `GET /metrics`
- `GET /cache-metrics`

Notes:
- `/metrics` includes embedding latency histogram and cache metrics.
- `/metrics` includes current lexical/semantic weights and temperature.
- `/search/hybrid-tune` updates lexical/semantic weights and temperature at runtime; weights auto-normalize to sum `1.0`.
- `/search/hybrid-tune` response includes `version` and `updated_at`.
- `/metrics` includes `rechain_rag_hybrid_tune_updates_total`, `rechain_rag_hybrid_tune_config_version`, `rechain_rag_hybrid_tune_updated_unix`.
- Runtime hybrid tune config persists in BoltDB (`RAG_CONFIG_PATH`) across restarts.
- `/metrics` includes `rechain_rag_hybrid_tune_import_total` and `rechain_rag_hybrid_tune_export_total`.

## Web6-3D (8084)
- `GET /health`
- `GET /graph`
- `GET /` (viewer)
- `GET /metrics`
- `GET /queue-depth`
- `GET /last-task-trace`
- `GET /tasks/recent?limit=...`
- `GET /task-trace?id=...`
- `GET /dashboard-summary`
- `GET /dashboard-summary?format=prom` (Prometheus proxy for orchestrator `/dashboard/summary?format=prom`)
- `GET /dashboard-web6`
- `GET /dashboard-web6?format=prom`
- `GET /dashboard-web6/alerts`
- `GET /dashboard-web6/alerts?format=prom`
- `GET /dashboard-web6/summary`
- `GET /dashboard-web6/summary?format=prom`
- `GET /dashboard-web6/history`
- `GET /dashboard-web6/history?format=prom`
- `POST /dashboard-web6/history` (reset in-memory history)
- `GET /replay-chain?id=...`
- `POST /task-replay?id=...&mode=force-policy|force-agent|force-agent-soft`
- `GET /task-debug-prom?id=...&scope=task|global|all`
- `GET /debug-compare?id1=...&id2=...`
- `GET /debug-compare?format=prom&id1=...&id2=...`
- `GET /proxy-counters`
- `GET /proxy-counters?format=prom`
- `GET /proxy-counters/history`
- `GET /proxy-counters/history?format=prom`
- `POST /proxy-counters/history` (reset in-memory history)
- `GET /proxy-counters/health`
- `GET /proxy-counters/health?format=prom`
- `GET /proxy-counters/alerts`
- `GET /proxy-counters/alerts?format=prom`
- `POST /cache/clear`
- `GET /cache/metrics`

Notes:
- `/graph` accepts optional `q` filter, `depth` (context hop), and `type` (file/dir/pkg/unknown). Graph format documented in docs/web6-graph.md.
- `/cache/metrics` JSON includes `filter_type` and `filter_type_active` map.
- `/debug-compare?format=prom` exports `rechain_web6_debug_compare_quality_delta`, `rechain_web6_debug_compare_confidence_delta`, and merge-source pair labels.
- `/proxy-counters` JSON includes: `debug_compare_total`, `debug_compare_prom_total`, `proxy_counters_total`, `proxy_counters_prom_total`, `proxy_alerts_total`, `proxy_alerts_prom_total`, `proxy_last_json_unix`, `proxy_last_prom_unix`, `dashboard_summary_total`, `dashboard_summary_prom_total`.
- `/proxy-counters/history` accepts `limit` and `format_filter` (`all|json|prom`) and returns recent proxy counter snapshots.
- `/proxy-counters/health` includes freshness fields: `proxy_json_age_sec`, `proxy_prom_age_sec`, `proxy_json_stale`, `proxy_prom_stale`, `stale_threshold_sec`.
- `/proxy-counters/alerts` includes `level` (`ok|warn|critical`), `level_score` (`0|1|2`), and `reason`.
- `/dashboard-web6/history` accepts `limit` query parameter for recent event window.
- `/dashboard-web6/history` also accepts `level` (`all|ok|warn|critical`) and `source` (`all|alerts|summary`) filters.
- `/tasks/recent` supports pass-through filters:
  - `state=queued|running|completed|failed|canceled|all`
  - `merge_source=policy_merge|agent_compiler|all`
  - `has_parent=yes|no|all`
  - `sort=updated_desc|updated_asc|quality_desc|quality_asc`

## Quantum (8085)
- `GET /health`
- `POST /optimize`
- `GET /metrics`

## Agent Compiler (8086)
- `GET /health`
- `POST /compile`
- `GET /metrics`

## Request IDs
- Clients may send X-Request-Id header.
- Services respond with X-Request-Id.

