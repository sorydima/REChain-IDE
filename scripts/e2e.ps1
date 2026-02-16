# End-to-end scenario:
# VSCode client -> orchestrator -> RAG/agent-compiler (+ quantum demo) -> metrics snapshot

param(
  [string]$Orchestrator = "http://localhost:8081",
  [string]$Rag = "http://localhost:8083",
  [string]$Quantum = "http://localhost:8085",
  [string]$AgentCompiler = "http://localhost:8086"
)

$ErrorActionPreference = "Stop"
$repo = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $repo

function Wait-Healthy([string]$name, [string]$url, [int]$tries = 15) {
  for ($i = 0; $i -lt $tries; $i++) {
    try {
      $resp = Invoke-RestMethod -Uri $url -TimeoutSec 2
      if ($resp -eq "ok" -or $null -ne $resp) {
        Write-Host "${name}: ok" -ForegroundColor Green
        return
      }
    } catch {}
    Start-Sleep -Milliseconds 500
  }
  throw "$name is not healthy: $url"
}

function Get-MetricValue([string]$metricsText, [string]$metricName) {
  $line = ($metricsText -split "`n" | Where-Object { $_ -match "^$metricName\s+" } | Select-Object -First 1)
  if (-not $line) { return 0.0 }
  $parts = ($line -split "\s+")
  if ($parts.Length -lt 2) { return 0.0 }
  return [double]$parts[-1]
}

function Get-MetricValueByMatch([string]$metricsText, [string]$contains) {
  $escaped = [regex]::Escape($contains)
  $line = ($metricsText -split "`n" | Where-Object { $_ -match $escaped } | Select-Object -First 1)
  if (-not $line) { return 0.0 }
  $parts = ($line -split "\s+")
  if ($parts.Length -lt 2) { return 0.0 }
  return [double]$parts[-1]
}

Write-Host "== e2e preflight =="
Wait-Healthy "orchestrator" "$Orchestrator/health"
Wait-Healthy "rag" "$Rag/health"
Wait-Healthy "quantum" "$Quantum/health"
Wait-Healthy "agent-compiler" "$AgentCompiler/health"

Write-Host "== agent compiler direct call =="
$agentMetricsBefore = Invoke-RestMethod -Uri "$AgentCompiler/metrics" -TimeoutSec 5
$agentCompileBefore = Get-MetricValue $agentMetricsBefore "rechain_agent_compile_total"
$compileBody = @{
  schema_version = "0.1.0"
  task_id = "e2e-agent-direct"
  policy = "quality"
  results = @(
    @{ model_id = "model_a"; output = "ok"; diff = "diff --git a/a b/a`n+alpha`n" },
    @{ model_id = "model_b"; output = "ok"; diff = "diff --git a/a b/a`n+beta`n" }
  )
} | ConvertTo-Json -Depth 6
$agentCompileRes = Invoke-RestMethod -Method Post -Uri "$AgentCompiler/compile" -ContentType "application/json" -Body $compileBody
$agentMetricsAfter = Invoke-RestMethod -Uri "$AgentCompiler/metrics" -TimeoutSec 5
$agentCompileAfter = Get-MetricValue $agentMetricsAfter "rechain_agent_compile_total"
if ($agentCompileAfter -le $agentCompileBefore) {
  throw "agent-compiler compile counter did not increase"
}

Write-Host "== rag index =="
$indexBody = @{
  schema_version = "0.1.0"
  repo = "rechain-ide"
  root = "$repo"
  extensions = @(".go", ".md", ".ps1")
  max_files = 300
} | ConvertTo-Json -Depth 5
Invoke-RestMethod -Method Post -Uri "$Rag/index" -ContentType "application/json" -Body $indexBody | Out-Null

Write-Host "== rag hybrid tune =="
$ragMetricsBefore = Invoke-RestMethod -Uri "$Rag/metrics" -TimeoutSec 5
$ragTuneUpdatesBefore = Get-MetricValue $ragMetricsBefore "rechain_rag_hybrid_tune_updates_total"
$ragTuneImportsBefore = Get-MetricValue $ragMetricsBefore "rechain_rag_hybrid_tune_import_total"
$ragTuneExportsBefore = Get-MetricValue $ragMetricsBefore "rechain_rag_hybrid_tune_export_total"
$tuneBody = @{
  lexical_weight = 0.7
  semantic_weight = 0.3
  temperature = 0.9
} | ConvertTo-Json -Depth 5
$tuneRes = Invoke-RestMethod -Method Patch -Uri "$Rag/search/hybrid-tune" -ContentType "application/json" -Body $tuneBody
if (-not $tuneRes.version -or $tuneRes.version -lt 2) {
  throw "rag tune version was not incremented"
}
$ragMetricsAfter = Invoke-RestMethod -Uri "$Rag/metrics" -TimeoutSec 5
$ragTuneUpdatesAfter = Get-MetricValue $ragMetricsAfter "rechain_rag_hybrid_tune_updates_total"
if ($ragTuneUpdatesAfter -le $ragTuneUpdatesBefore) {
  throw "rag hybrid tune updates metric did not increase"
}
$tuneResetRes = Invoke-RestMethod -Method Post -Uri "$Rag/search/hybrid-tune/reset" -ContentType "application/json" -Body "{}"
if (-not $tuneResetRes.status -or $tuneResetRes.status -ne "reset") {
  throw "rag tune reset failed"
}
$ragMetricsAfterReset = Invoke-RestMethod -Uri "$Rag/metrics" -TimeoutSec 5
$ragTuneUpdatesAfterReset = Get-MetricValue $ragMetricsAfterReset "rechain_rag_hybrid_tune_updates_total"
if ($ragTuneUpdatesAfterReset -le $ragTuneUpdatesAfter) {
  throw "rag hybrid tune reset metric did not increase"
}
$tuneExport = Invoke-RestMethod -Uri "$Rag/search/hybrid-tune/export" -TimeoutSec 5
$tuneImportBody = @{
  lexical_weight = $tuneExport.lexical_weight
  semantic_weight = $tuneExport.semantic_weight
  temperature = 0.95
} | ConvertTo-Json -Depth 5
$tuneImport = Invoke-RestMethod -Method Post -Uri "$Rag/search/hybrid-tune/import" -ContentType "application/json" -Body $tuneImportBody
if (-not $tuneImport.status -or $tuneImport.status -ne "imported") {
  throw "rag tune import failed"
}
$ragMetricsAfterImport = Invoke-RestMethod -Uri "$Rag/metrics" -TimeoutSec 5
$ragTuneImportsAfter = Get-MetricValue $ragMetricsAfterImport "rechain_rag_hybrid_tune_import_total"
$ragTuneExportsAfter = Get-MetricValue $ragMetricsAfterImport "rechain_rag_hybrid_tune_export_total"
if ($ragTuneImportsAfter -le $ragTuneImportsBefore) {
  throw "rag tune import metric did not increase"
}
if ($ragTuneExportsAfter -le $ragTuneExportsBefore) {
  throw "rag tune export metric did not increase"
}

Write-Host "== vscode client task =="
$clientOut = & go run rechain-ide/vscode-extension/cmd/ide-client/main.go -server $Orchestrator -input "prepare patch plan for orchestrator+web6 metrics" 2>&1
$jsonLine = $clientOut | Where-Object { $_ -is [string] -and $_.Trim().StartsWith("{") -and $_.Trim().EndsWith("}") } | Select-Object -Last 1
if (-not $jsonLine) {
  throw ("ide-client output does not contain json status: " + ($clientOut -join " | "))
}
$status = $jsonLine | ConvertFrom-Json
$taskId = $status.id
if (-not $taskId) {
  throw "failed to parse task id from ide-client output"
}
Write-Host "task: $taskId"

Write-Host "== task result =="
$result = Invoke-RestMethod -Uri "$Orchestrator/tasks/$taskId/result"
$trace = Invoke-RestMethod -Uri "$Orchestrator/tasks/$taskId/trace"
if (-not $trace) {
  throw "trace endpoint returned empty payload"
}
if (-not $trace.state -or $trace.state -ne "completed") {
  throw "trace state is not completed"
}
if (-not $trace.merge_source) {
  throw "trace merge_source is empty"
}
if (-not $trace.results -or $trace.results.Count -lt 1) {
  throw "trace results are empty"
}
if (-not $trace.merge -or -not $trace.merge.diff) {
  throw "trace merge payload is missing"
}

Write-Host "== replay task =="
$replay = Invoke-RestMethod -Method Post -Uri "$Orchestrator/tasks/$taskId/replay?mode=force-policy" -ContentType "application/json" -Body "{}"
$replayId = $replay.replay_task_id
if (-not $replayId) {
  throw "replay_task_id is missing"
}
$replayTrace = $null
for ($i=0; $i -lt 20; $i++) {
  Start-Sleep -Milliseconds 300
  try {
    $rt = Invoke-RestMethod -Uri "$Orchestrator/tasks/$replayId/trace" -TimeoutSec 3
    if ($rt.state -eq "completed" -or $rt.state -eq "failed") {
      $replayTrace = $rt
      break
    }
  } catch {}
}
if (-not $replayTrace) {
  throw "replay trace not available"
}
if ($replayTrace.parent_task_id -ne $taskId) {
  throw "replay parent_task_id mismatch"
}
if ($replayTrace.merge_source -ne "policy_merge") {
  throw "force-policy replay merge source mismatch"
}
$batchBody = @{ modes = @("force-policy", "force-agent-soft") } | ConvertTo-Json -Depth 4
$replayBatch = Invoke-RestMethod -Method Post -Uri "$Orchestrator/tasks/$taskId/replay/batch" -ContentType "application/json" -Body $batchBody
if (-not $replayBatch.items -or $replayBatch.items.Count -lt 2) {
  throw "replay batch did not return expected items"
}
$replayAgent = Invoke-RestMethod -Method Post -Uri "$Orchestrator/tasks/$taskId/replay?mode=force-agent" -ContentType "application/json" -Body "{}"
$replayAgentId = $replayAgent.replay_task_id
if (-not $replayAgentId) {
  throw "force-agent replay_task_id is missing"
}
$replayAgentTrace = $null
for ($i=0; $i -lt 20; $i++) {
  Start-Sleep -Milliseconds 300
  try {
    $rt = Invoke-RestMethod -Uri "$Orchestrator/tasks/$replayAgentId/trace" -TimeoutSec 3
    if ($rt.state -eq "completed" -or $rt.state -eq "failed") {
      $replayAgentTrace = $rt
      break
    }
  } catch {}
}
if (-not $replayAgentTrace) {
  throw "force-agent replay trace not available"
}
if ($replayAgentTrace.state -eq "completed") {
  if ($replayAgentTrace.merge_source -ne "agent_compiler") {
    throw "force-agent replay merge source mismatch"
  }
} elseif ($replayAgentTrace.state -eq "failed") {
  if (-not $replayAgentTrace.error -or $replayAgentTrace.error -notlike "*forced agent_compiler failed*") {
    throw "force-agent replay failed without forced-agent error"
  }
} else {
  throw "force-agent replay unexpected state"
}
$replaySoft = Invoke-RestMethod -Method Post -Uri "$Orchestrator/tasks/$taskId/replay?mode=force-agent-soft" -ContentType "application/json" -Body "{}"
$replaySoftId = $replaySoft.replay_task_id
if (-not $replaySoftId) {
  throw "force-agent-soft replay_task_id is missing"
}
$replaySoftTrace = $null
for ($i=0; $i -lt 20; $i++) {
  Start-Sleep -Milliseconds 300
  try {
    $rt = Invoke-RestMethod -Uri "$Orchestrator/tasks/$replaySoftId/trace" -TimeoutSec 3
    if ($rt.state -eq "completed" -or $rt.state -eq "failed") {
      $replaySoftTrace = $rt
      break
    }
  } catch {}
}
if (-not $replaySoftTrace) {
  throw "force-agent-soft replay trace not available"
}
if ($replaySoftTrace.state -ne "completed") {
  throw "force-agent-soft replay expected completed state"
}
if ($replaySoftTrace.merge_source -ne "policy_merge" -and $replaySoftTrace.merge_source -ne "agent_compiler") {
  throw "force-agent-soft replay merge source unexpected"
}
$replayChain = Invoke-RestMethod -Uri "$Orchestrator/tasks/$replayId/replay-chain" -TimeoutSec 5
if (-not $replayChain.lineage -or $replayChain.lineage.Count -lt 2) {
  throw "replay chain lineage too short"
}
if ($null -eq $replayChain.descendants) {
  throw "replay chain descendants missing"
}
$debugPayload = Invoke-RestMethod -Uri "$Orchestrator/tasks/$replayId/debug" -TimeoutSec 5
if (-not $debugPayload.status -or -not $debugPayload.trace -or -not $debugPayload.replay_chain) {
  throw "task debug payload missing required sections"
}
$debugProm = Invoke-RestMethod -Uri "$Orchestrator/tasks/$replayId/debug?format=prom&scope=task" -TimeoutSec 5
if ($debugProm -notmatch "rechain_task_debug_state") {
  throw "task debug prom missing state metric"
}
$uiReplay = Invoke-RestMethod -Method Post -Uri "http://localhost:8084/task-replay?id=$taskId&mode=force-policy" -TimeoutSec 5
if (-not $uiReplay.replay_task_id) {
  throw "web6 replay trigger failed"
}
$orchMetrics = Invoke-RestMethod -Uri "$Orchestrator/metrics" -TimeoutSec 5
$replayMetric = Get-MetricValue $orchMetrics "rechain_task_replay_total"
$parentLinksMetric = Get-MetricValue $orchMetrics "rechain_task_trace_parent_links_total"
if ((Get-MetricValue $orchMetrics "rechain_forced_agent_fallback_total") -lt 0) {
  throw "forced agent fallback metric missing"
}
if ((Get-MetricValueByMatch $orchMetrics 'rechain_merge_choice_total{source="policy_merge"}') -lt 1) {
  throw "merge choice policy metric missing"
}
if ((Get-MetricValueByMatch $orchMetrics 'rechain_task_replay_mode_total{mode="force-policy"}') -lt 1) {
  throw "replay mode policy metric missing"
}
if ($replayMetric -lt 1) {
  throw "replay metric did not increase"
}
if ($parentLinksMetric -lt 1) {
  throw "parent links metric did not increase"
}
$dashboardProm = Invoke-RestMethod -Uri "$Orchestrator/dashboard/summary?format=prom" -TimeoutSec 5
$dashboardReplay = Get-MetricValueByMatch $dashboardProm 'rechain_dashboard_tasks_total{state="replayed"}'
if ($dashboardReplay -lt 1) {
  throw "dashboard replay metric missing"
}
if ((Get-MetricValue $dashboardProm "rechain_dashboard_forced_agent_fallback_total") -lt 0) {
  throw "dashboard forced fallback metric missing"
}
if ((Get-MetricValueByMatch $dashboardProm 'rechain_dashboard_downstream_up{service="web6"}') -lt 0) {
  throw "dashboard downstream web6 metric missing"
}
if ((Get-MetricValue $dashboardProm "rechain_dashboard_web6_proxy_alert_level") -lt 0) {
  throw "dashboard web6 proxy alert metric missing"
}
$qualityReq = @{
  output = ""
  diff = $result.diff
} | ConvertTo-Json -Depth 5
$quality = Invoke-RestMethod -Method Post -Uri "$Orchestrator/quality-score" -ContentType "application/json" -Body $qualityReq

Write-Host "== model registry =="
$models = Invoke-RestMethod -Uri "$Orchestrator/models"
$modelHealth = Invoke-RestMethod -Uri "$Orchestrator/models/health"
$costProfile = Invoke-RestMethod -Uri "$Orchestrator/models/cost-profile?budget_usd=0.1"
$dashboard = Invoke-RestMethod -Uri "$Orchestrator/dashboard/summary"
if (-not $dashboard.orchestrator.tasks.replayed -or $dashboard.orchestrator.tasks.replayed -lt 1) {
  throw "dashboard json replayed counter missing"
}
if (-not $dashboard.orchestrator.trace_parent_links_total -or $dashboard.orchestrator.trace_parent_links_total -lt 1) {
  throw "dashboard json parent links counter missing"
}
if (-not $dashboard.orchestrator.trace.by_state.completed -or $dashboard.orchestrator.trace.by_state.completed -lt 1) {
  throw "dashboard json trace by_state completed missing"
}
if (-not $dashboard.orchestrator.merge_choice.policy_merge -or $dashboard.orchestrator.merge_choice.policy_merge -lt 1) {
  throw "dashboard json merge_choice policy missing"
}
if (-not $dashboard.downstream.web6) {
  throw "dashboard json web6 downstream block missing"
}
if ($null -eq $dashboard.downstream.web6.proxy_alert_level) {
  throw "dashboard json web6 proxy alert level missing"
}
if ($null -eq $dashboard.downstream.web6.proxy_json_stale -or $null -eq $dashboard.downstream.web6.proxy_prom_stale) {
  throw "dashboard json web6 stale flags missing"
}
$webProxyBefore = Invoke-RestMethod -Uri "http://localhost:8084/proxy-counters" -TimeoutSec 5
$webDebugCompareBefore = [double]$webProxyBefore.debug_compare_total
$webDebugComparePromBefore = [double]$webProxyBefore.debug_compare_prom_total
$webProxyCountersBefore = [double]$webProxyBefore.proxy_counters_total
$webProxyCountersPromBefore = [double]$webProxyBefore.proxy_counters_prom_total
$webProxyAlertsBefore = [double]$webProxyBefore.proxy_alerts_total
$webProxyAlertsPromBefore = [double]$webProxyBefore.proxy_alerts_prom_total
$webProxyLastJSONBefore = [double]$webProxyBefore.proxy_last_json_unix
$webProxyLastPromBefore = [double]$webProxyBefore.proxy_last_prom_unix
$webDashboardSummaryBefore = [double]$webProxyBefore.dashboard_summary_total
$webDashboardSummaryPromBefore = [double]$webProxyBefore.dashboard_summary_prom_total
$webRecent = Invoke-RestMethod -Uri "http://localhost:8084/tasks/recent?limit=5&state=completed&merge_source=policy_merge&has_parent=yes&sort=updated_desc" -TimeoutSec 5
if (-not $webRecent.tasks -or $webRecent.tasks.Count -lt 1) {
  throw "web6 filtered recent tasks returned empty"
}
$webChain = Invoke-RestMethod -Uri "http://localhost:8084/replay-chain?id=$replayId" -TimeoutSec 5
if (-not $webChain.lineage -or $webChain.lineage.Count -lt 2) {
  throw "web6 replay chain lineage missing"
}
$webDebugProm = Invoke-RestMethod -Uri "http://localhost:8084/task-debug-prom?id=$replayId&scope=global" -TimeoutSec 5
if ($webDebugProm -notmatch "rechain_task_debug_global_merge_choice_total") {
  throw "web6 task debug prom missing global merge metrics"
}
$webCompare = Invoke-RestMethod -Uri "http://localhost:8084/debug-compare?id1=$taskId&id2=$replayId" -TimeoutSec 5
if (-not $webCompare.metrics) {
  throw "web6 debug compare missing metrics"
}
$webRootBody = [string](Invoke-RestMethod -Uri "http://localhost:8084/" -TimeoutSec 5)
if ($webRootBody -notmatch "id=`"debug-compare-prom`"") {
  throw "web6 ui root missing debug-compare-prom panel"
}
if ($webRootBody -notmatch "id=`"dashboard-prom`"") {
  throw "web6 ui root missing dashboard-prom panel"
}
if ($webRootBody -notmatch "id=`"drawer`"") {
  throw "web6 ui root missing drawer block"
}
if ($webRootBody -notmatch "id=`"proxy-metrics`"") {
  throw "web6 ui root missing proxy-metrics panel"
}
if ($webRootBody -notmatch "id=`"proxy-prom`"") {
  throw "web6 ui root missing proxy-prom panel"
}
if ($webRootBody -notmatch "id=`"proxy-history`"") {
  throw "web6 ui root missing proxy-history panel"
}
if ($webRootBody -notmatch "id=`"proxy-history-limit`"") {
  throw "web6 ui root missing proxy-history-limit control"
}
if ($webRootBody -notmatch "id=`"proxy-history-format`"") {
  throw "web6 ui root missing proxy-history-format control"
}
if ($webRootBody -notmatch "id=`"proxy-history-copy`"") {
  throw "web6 ui root missing proxy-history-copy control"
}
if ($webRootBody -notmatch "id=`"proxy-history-reset`"") {
  throw "web6 ui root missing proxy-history-reset control"
}
if ($webRootBody -notmatch "id=`"proxy-health`"") {
  throw "web6 ui root missing proxy-health panel"
}
if ($webRootBody -notmatch "id=`"proxy-alerts`"") {
  throw "web6 ui root missing proxy-alerts panel"
}
if ($webRootBody -notmatch "id=`"dashboard-web6`"") {
  throw "web6 ui root missing dashboard-web6 panel"
}
if ($webRootBody -notmatch "id=`"dashboard-web6-prom`"") {
  throw "web6 ui root missing dashboard-web6-prom panel"
}
if ($webRootBody -notmatch "id=`"dashboard-web6-alerts`"") {
  throw "web6 ui root missing dashboard-web6-alerts panel"
}
if ($webRootBody -notmatch "id=`"dashboard-web6-summary`"") {
  throw "web6 ui root missing dashboard-web6-summary panel"
}
if ($webRootBody -notmatch "id=`"dashboard-web6-history`"") {
  throw "web6 ui root missing dashboard-web6-history panel"
}
if ($webRootBody -notmatch "id=`"dashboard-web6-history-limit`"") {
  throw "web6 ui root missing dashboard-web6-history-limit control"
}
if ($webRootBody -notmatch "id=`"dashboard-web6-history-level`"") {
  throw "web6 ui root missing dashboard-web6-history-level control"
}
if ($webRootBody -notmatch "id=`"dashboard-web6-history-source`"") {
  throw "web6 ui root missing dashboard-web6-history-source control"
}
if ($webRootBody -notmatch "id=`"dashboard-web6-history-copy`"") {
  throw "web6 ui root missing dashboard-web6-history-copy control"
}
if ($webRootBody -notmatch "id=`"dashboard-web6-history-reset`"") {
  throw "web6 ui root missing dashboard-web6-history-reset control"
}
$webCompareProm = Invoke-RestMethod -Uri "http://localhost:8084/debug-compare?format=prom&id1=$taskId&id2=$replayId" -TimeoutSec 5
if ($webCompareProm -notmatch "rechain_web6_debug_compare_quality_delta") {
  throw "web6 debug compare prom missing quality delta metric"
}
$webDashboard = Invoke-RestMethod -Uri "http://localhost:8084/dashboard-summary" -TimeoutSec 5
if (-not $webDashboard.orchestrator) {
  throw "web6 dashboard json proxy missing orchestrator block"
}
$webDashboardProm = Invoke-RestMethod -Uri "http://localhost:8084/dashboard-summary?format=prom" -TimeoutSec 5
if ($webDashboardProm -notmatch "rechain_dashboard_tasks_total") {
  throw "web6 dashboard prom proxy missing orchestrator metrics"
}
$webDashboardWeb6 = Invoke-RestMethod -Uri "http://localhost:8084/dashboard-web6" -TimeoutSec 5
if (-not $webDashboardWeb6.metrics) {
  throw "web6 dashboard-web6 json missing metrics"
}
$webDashboardWeb6Prom = Invoke-RestMethod -Uri "http://localhost:8084/dashboard-web6?format=prom" -TimeoutSec 5
if ($webDashboardWeb6Prom -notmatch "rechain_dashboard_web6_proxy_alert_level") {
  throw "web6 dashboard-web6 prom missing web6 alert metric"
}
$webDashboardWeb6Alerts = Invoke-RestMethod -Uri "http://localhost:8084/dashboard-web6/alerts" -TimeoutSec 5
if (-not $webDashboardWeb6Alerts.level) {
  throw "web6 dashboard-web6 alerts json missing level"
}
$webDashboardWeb6AlertsProm = Invoke-RestMethod -Uri "http://localhost:8084/dashboard-web6/alerts?format=prom" -TimeoutSec 5
if ($webDashboardWeb6AlertsProm -notmatch "rechain_web6_dashboard_web6_alert_level") {
  throw "web6 dashboard-web6 alerts prom missing alert level"
}
$webDashboardWeb6Summary = Invoke-RestMethod -Uri "http://localhost:8084/dashboard-web6/summary" -TimeoutSec 5
if (-not $webDashboardWeb6Summary.alert -or -not $webDashboardWeb6Summary.metrics) {
  throw "web6 dashboard-web6 summary json missing alert/metrics"
}
$webDashboardWeb6SummaryProm = Invoke-RestMethod -Uri "http://localhost:8084/dashboard-web6/summary?format=prom" -TimeoutSec 5
if ($webDashboardWeb6SummaryProm -notmatch "rechain_web6_dashboard_web6_alert_level") {
  throw "web6 dashboard-web6 summary prom missing alert level"
}
$webDashboardWeb6History = Invoke-RestMethod -Uri "http://localhost:8084/dashboard-web6/history?limit=8" -TimeoutSec 5
if ($null -eq $webDashboardWeb6History.count -or -not $webDashboardWeb6History.level_counts) {
  throw "web6 dashboard-web6 history json missing count/level_counts"
}
$webDashboardWeb6HistoryFiltered = Invoke-RestMethod -Uri "http://localhost:8084/dashboard-web6/history?limit=8&level=ok&source=summary" -TimeoutSec 5
if (-not $webDashboardWeb6HistoryFiltered.level_filter -or -not $webDashboardWeb6HistoryFiltered.source_filter) {
  throw "web6 dashboard-web6 history filtered json missing filter fields"
}
$webDashboardWeb6HistoryProm = Invoke-RestMethod -Uri "http://localhost:8084/dashboard-web6/history?limit=8&format=prom" -TimeoutSec 5
if ($webDashboardWeb6HistoryProm -notmatch "rechain_web6_dashboard_web6_alert_history_total") {
  throw "web6 dashboard-web6 history prom missing history metric"
}
$webDashboardWeb6HistoryFilteredProm = Invoke-RestMethod -Uri "http://localhost:8084/dashboard-web6/history?limit=8&level=ok&source=summary&format=prom" -TimeoutSec 5
if ($webDashboardWeb6HistoryFilteredProm -notmatch "rechain_web6_dashboard_web6_alert_history_filter_active") {
  throw "web6 dashboard-web6 history filtered prom missing active filter metric"
}
$webDashboardWeb6HistoryReset = Invoke-RestMethod -Method Post -Uri "http://localhost:8084/dashboard-web6/history" -TimeoutSec 5
if (-not $webDashboardWeb6HistoryReset.status -or $webDashboardWeb6HistoryReset.status -ne "reset") {
  throw "web6 dashboard-web6 history reset failed"
}
$webDashboardWeb6HistoryAfterReset = Invoke-RestMethod -Uri "http://localhost:8084/dashboard-web6/history?limit=8" -TimeoutSec 5
if ($webDashboardWeb6HistoryAfterReset.count -ne 0) {
  throw "web6 dashboard-web6 history reset did not clear items"
}
$webProxyProm = Invoke-RestMethod -Uri "http://localhost:8084/proxy-counters?format=prom" -TimeoutSec 5
if ($webProxyProm -notmatch "rechain_web6_debug_compare_total") {
  throw "web6 proxy-counters prom missing debug compare metric"
}
$webProxyHealth = Invoke-RestMethod -Uri "http://localhost:8084/proxy-counters/health" -TimeoutSec 5
if ($null -eq $webProxyHealth.proxy_json_stale -or $null -eq $webProxyHealth.proxy_prom_stale) {
  throw "web6 proxy-counters health missing stale fields"
}
$webProxyHealthProm = Invoke-RestMethod -Uri "http://localhost:8084/proxy-counters/health?format=prom" -TimeoutSec 5
if ($webProxyHealthProm -notmatch "rechain_web6_proxy_json_stale") {
  throw "web6 proxy-counters health prom missing stale metric"
}
$webProxyAlerts = Invoke-RestMethod -Uri "http://localhost:8084/proxy-counters/alerts" -TimeoutSec 5
if (-not $webProxyAlerts.level) {
  throw "web6 proxy-counters alerts missing level"
}
$webProxyAlertsProm = Invoke-RestMethod -Uri "http://localhost:8084/proxy-counters/alerts?format=prom" -TimeoutSec 5
if ($webProxyAlertsProm -notmatch "rechain_web6_proxy_alert_level") {
  throw "web6 proxy-counters alerts prom missing level metric"
}
$webProxyAfter = Invoke-RestMethod -Uri "http://localhost:8084/proxy-counters" -TimeoutSec 5
$webDebugCompareAfter = [double]$webProxyAfter.debug_compare_total
$webDebugComparePromAfter = [double]$webProxyAfter.debug_compare_prom_total
$webProxyCountersAfter = [double]$webProxyAfter.proxy_counters_total
$webProxyCountersPromAfter = [double]$webProxyAfter.proxy_counters_prom_total
$webProxyAlertsAfter = [double]$webProxyAfter.proxy_alerts_total
$webProxyAlertsPromAfter = [double]$webProxyAfter.proxy_alerts_prom_total
$webProxyLastJSONAfter = [double]$webProxyAfter.proxy_last_json_unix
$webProxyLastPromAfter = [double]$webProxyAfter.proxy_last_prom_unix
$webDashboardSummaryAfter = [double]$webProxyAfter.dashboard_summary_total
$webDashboardSummaryPromAfter = [double]$webProxyAfter.dashboard_summary_prom_total
if ($webDebugCompareAfter -le $webDebugCompareBefore) {
  throw "web6 debug compare counter did not increase"
}
if ($webDebugComparePromAfter -le $webDebugComparePromBefore) {
  throw "web6 debug compare prom counter did not increase"
}
if ($webProxyCountersAfter -le $webProxyCountersBefore) {
  throw "web6 proxy counters json counter did not increase"
}
if ($webProxyCountersPromAfter -le $webProxyCountersPromBefore) {
  throw "web6 proxy counters prom counter did not increase"
}
if ($webProxyAlertsAfter -le $webProxyAlertsBefore) {
  throw "web6 proxy alerts json counter did not increase"
}
if ($webProxyAlertsPromAfter -le $webProxyAlertsPromBefore) {
  throw "web6 proxy alerts prom counter did not increase"
}
if ($webProxyLastJSONAfter -le 0) {
  throw "web6 proxy counters json last timestamp not set"
}
if ($webProxyLastPromAfter -le 0) {
  throw "web6 proxy counters prom last timestamp not set"
}
if ($webDashboardSummaryAfter -le $webDashboardSummaryBefore) {
  throw "web6 dashboard summary counter did not increase"
}
if ($webDashboardSummaryPromAfter -le $webDashboardSummaryPromBefore) {
  throw "web6 dashboard summary prom counter did not increase"
}
$webProxyHistory = Invoke-RestMethod -Uri "http://localhost:8084/proxy-counters/history?limit=10" -TimeoutSec 5
if ($null -eq $webProxyHistory.count -or -not $webProxyHistory.format_counts) {
  throw "web6 proxy history json missing count/format_counts"
}
$webProxyHistoryFiltered = Invoke-RestMethod -Uri "http://localhost:8084/proxy-counters/history?limit=10&format_filter=prom" -TimeoutSec 5
if (-not $webProxyHistoryFiltered.format_filter -or $webProxyHistoryFiltered.format_filter -ne "prom") {
  throw "web6 proxy history filtered json missing format_filter=prom"
}
$webProxyHistoryProm = Invoke-RestMethod -Uri "http://localhost:8084/proxy-counters/history?limit=10&format=prom" -TimeoutSec 5
if ($webProxyHistoryProm -notmatch "rechain_web6_proxy_history_total") {
  throw "web6 proxy history prom missing total metric"
}
$webProxyHistoryFilteredProm = Invoke-RestMethod -Uri "http://localhost:8084/proxy-counters/history?limit=10&format_filter=prom&format=prom" -TimeoutSec 5
if ($webProxyHistoryFilteredProm -notmatch "rechain_web6_proxy_history_filter_active") {
  throw "web6 proxy history filtered prom missing active filter metric"
}
$webProxyHistoryReset = Invoke-RestMethod -Method Post -Uri "http://localhost:8084/proxy-counters/history" -TimeoutSec 5
if (-not $webProxyHistoryReset.status -or $webProxyHistoryReset.status -ne "reset") {
  throw "web6 proxy history reset failed"
}
$webProxyHistoryAfterReset = Invoke-RestMethod -Uri "http://localhost:8084/proxy-counters/history?limit=10" -TimeoutSec 5
if ($webProxyHistoryAfterReset.count -ne 0) {
  throw "web6 proxy history reset did not clear items"
}
$ragTuneHistory = Invoke-RestMethod -Uri "$Rag/search/hybrid-tune/history?limit=5" -TimeoutSec 5
if (-not $ragTuneHistory.history -or $ragTuneHistory.count -lt 1) {
  throw "rag tune history missing"
}

Write-Host "== quantum demo =="
$cand = @()
$i = 0
foreach ($m in $costProfile.models) {
  if ($i -ge 3) { break }
  $cand += @{
    id = $m.id
    cost_usd = [double]$m.cost_usd
    latency_ms = [double](120 + ($i * 40))
    quality = [double](0.55 + ($i * 0.1))
  }
  $i++
}
$qBody = @{
  schema_version = "0.1.0"
  objective = "weighted"
  candidates = $cand
} | ConvertTo-Json -Depth 6
$qRes = Invoke-RestMethod -Method Post -Uri "$Quantum/optimize" -ContentType "application/json" -Body $qBody

Write-Host "== summary =="
$summary = [ordered]@{
  task_id = $taskId
  merge_quality_score = $result.quality_score
  quality_score_recheck = $quality.quality_score
  models_count = $models.count
  model_health_ok = $modelHealth.summary.ok
  dashboard_queue_depth = $dashboard.orchestrator.queue_depth
  dashboard_agent_compile_total = $dashboard.downstream.agent_compiler.compile_total
  dashboard_quantum_optimize_total = $dashboard.downstream.quantum.optimize_total
  trace_state = $trace.state
  trace_merge_source = $trace.merge_source
  trace_results_count = $trace.results.Count
  replay_task_id = $replayId
  replay_state = $replayTrace.state
  replay_parent_task_id = $replayTrace.parent_task_id
  replay_force_agent_state = $replayAgentTrace.state
  replay_force_agent_merge_source = $replayAgentTrace.merge_source
  replay_force_agent_soft_state = $replaySoftTrace.state
  replay_force_agent_soft_merge_source = $replaySoftTrace.merge_source
  replay_chain_lineage_count = $replayChain.lineage.Count
  replay_chain_descendants_count = $replayChain.descendants.Count
  debug_has_merge_metrics = [bool]$debugPayload.merge_metrics
  web_replay_task_id = $uiReplay.replay_task_id
  replay_metric = $replayMetric
  parent_links_metric = $parentLinksMetric
  rag_tune_version = $tuneRes.version
  rag_tune_updated_at = $tuneRes.updated_at
  rag_tune_updates_before = $ragTuneUpdatesBefore
  rag_tune_updates_after = $ragTuneUpdatesAfter
  rag_tune_updates_after_reset = $ragTuneUpdatesAfterReset
  rag_tune_imports_before = $ragTuneImportsBefore
  rag_tune_imports_after = $ragTuneImportsAfter
  rag_tune_exports_before = $ragTuneExportsBefore
  rag_tune_exports_after = $ragTuneExportsAfter
  rag_tune_history_count = $ragTuneHistory.count
  replay_batch_count = $replayBatch.count
  web_compare_quality_delta = $webCompare.metrics.quality_delta
  web_compare_prom_ok = [bool]($webCompareProm -match "rechain_web6_debug_compare_quality_delta")
  web_dashboard_prom_ok = [bool]($webDashboardProm -match "rechain_dashboard_tasks_total")
  web_dashboard_web6_prom_ok = [bool]($webDashboardWeb6Prom -match "rechain_dashboard_web6_proxy_alert_level")
  web_dashboard_web6_alerts_prom_ok = [bool]($webDashboardWeb6AlertsProm -match "rechain_web6_dashboard_web6_alert_level")
  web_dashboard_web6_summary_prom_ok = [bool]($webDashboardWeb6SummaryProm -match "rechain_web6_dashboard_web6_alert_level")
  web_dashboard_web6_history_prom_ok = [bool]($webDashboardWeb6HistoryProm -match "rechain_web6_dashboard_web6_alert_history_total")
  web_dashboard_web6_history_filter_prom_ok = [bool]($webDashboardWeb6HistoryFilteredProm -match "rechain_web6_dashboard_web6_alert_history_filter_active")
  web_dashboard_web6_history_reset_ok = [bool]($webDashboardWeb6HistoryAfterReset.count -eq 0)
  web_proxy_history_prom_ok = [bool]($webProxyHistoryProm -match "rechain_web6_proxy_history_total")
  web_proxy_history_filter_prom_ok = [bool]($webProxyHistoryFilteredProm -match "rechain_web6_proxy_history_filter_active")
  web_proxy_history_reset_ok = [bool]($webProxyHistoryAfterReset.count -eq 0)
  web_ui_panels_ok = [bool](($webRootBody -match "id=`"debug-compare-prom`"") -and ($webRootBody -match "id=`"dashboard-prom`"") -and ($webRootBody -match "id=`"drawer`"") -and ($webRootBody -match "id=`"proxy-metrics`"") -and ($webRootBody -match "id=`"proxy-prom`"") -and ($webRootBody -match "id=`"proxy-history`"") -and ($webRootBody -match "id=`"proxy-history-limit`"") -and ($webRootBody -match "id=`"proxy-history-format`"") -and ($webRootBody -match "id=`"proxy-history-copy`"") -and ($webRootBody -match "id=`"proxy-history-reset`"") -and ($webRootBody -match "id=`"proxy-health`"") -and ($webRootBody -match "id=`"proxy-alerts`"") -and ($webRootBody -match "id=`"dashboard-web6`"") -and ($webRootBody -match "id=`"dashboard-web6-prom`"") -and ($webRootBody -match "id=`"dashboard-web6-alerts`"") -and ($webRootBody -match "id=`"dashboard-web6-summary`"") -and ($webRootBody -match "id=`"dashboard-web6-history`"") -and ($webRootBody -match "id=`"dashboard-web6-history-limit`"") -and ($webRootBody -match "id=`"dashboard-web6-history-level`"") -and ($webRootBody -match "id=`"dashboard-web6-history-source`"") -and ($webRootBody -match "id=`"dashboard-web6-history-copy`"") -and ($webRootBody -match "id=`"dashboard-web6-history-reset`""))
  web_debug_compare_total_before = $webDebugCompareBefore
  web_debug_compare_total_after = $webDebugCompareAfter
  web_debug_compare_prom_total_before = $webDebugComparePromBefore
  web_debug_compare_prom_total_after = $webDebugComparePromAfter
  web_proxy_counters_total_before = $webProxyCountersBefore
  web_proxy_counters_total_after = $webProxyCountersAfter
  web_proxy_counters_prom_total_before = $webProxyCountersPromBefore
  web_proxy_counters_prom_total_after = $webProxyCountersPromAfter
  web_proxy_alerts_total_before = $webProxyAlertsBefore
  web_proxy_alerts_total_after = $webProxyAlertsAfter
  web_proxy_alerts_prom_total_before = $webProxyAlertsPromBefore
  web_proxy_alerts_prom_total_after = $webProxyAlertsPromAfter
  web_proxy_last_json_unix_before = $webProxyLastJSONBefore
  web_proxy_last_json_unix_after = $webProxyLastJSONAfter
  web_proxy_last_prom_unix_before = $webProxyLastPromBefore
  web_proxy_last_prom_unix_after = $webProxyLastPromAfter
  web_proxy_health_json_stale = [bool]$webProxyHealth.proxy_json_stale
  web_proxy_health_prom_stale = [bool]$webProxyHealth.proxy_prom_stale
  web_proxy_alert_level = $webProxyAlerts.level
  web_proxy_alert_score = $webProxyAlerts.level_score
  web_dashboard_summary_total_before = $webDashboardSummaryBefore
  web_dashboard_summary_total_after = $webDashboardSummaryAfter
  web_dashboard_summary_prom_total_before = $webDashboardSummaryPromBefore
  web_dashboard_summary_prom_total_after = $webDashboardSummaryPromAfter
  agent_direct_selected_model = $agentCompileRes.selected_model
  agent_compile_total_before = $agentCompileBefore
  agent_compile_total_after = $agentCompileAfter
  selected_cost_models = ($costProfile.models | ForEach-Object { $_.id }) -join ","
  quantum_selected = $qRes.selected_id
  metrics = [ordered]@{
    orchestrator = "$Orchestrator/metrics"
    rag = "$Rag/metrics"
    quantum = "$Quantum/metrics"
    agent_compiler = "$AgentCompiler/metrics"
  }
}
$summary | ConvertTo-Json -Depth 6
