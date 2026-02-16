# Start all services for local dev

param(
  [switch]$ImportGraph,
  [switch]$GoListGraph
)

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$repo = Resolve-Path (Join-Path $root '..')

$graphMode = "tree"
if ($GoListGraph) {
  $graphMode = "imports_go_list"
} elseif ($ImportGraph) {
  $graphMode = "imports"
}

if ($graphMode -eq "tree") {
  # Generate Web6 graph on each dev start
  & "$root\\gen-graph.ps1" -Root "$repo" -Out "$repo\\.web6-graph.json" | Out-Null
}

$graphPath = ""
if ($graphMode -eq "tree") {
  $graphPath = "$repo\\.web6-graph.json"
}

Start-Process powershell -ArgumentList "-NoProfile", "-Command", "cd `"$repo`"; $env:RAG_URL='http://localhost:8083'; $env:QUANTUM_URL='http://localhost:8085'; $env:AGENT_COMPILER_URL='http://localhost:8086'; go run rechain-ide/orchestrator/cmd/orchestrator/main.go" -WindowStyle Normal
Start-Process powershell -ArgumentList "-NoProfile", "-Command", "cd `"$repo`"; go run rechain-ide/kernel/cmd/kernel/main.go" -WindowStyle Normal
Start-Process powershell -ArgumentList "-NoProfile", "-Command", "cd `"$repo`"; go run rechain-ide/rag/cmd/rag/main.go" -WindowStyle Normal
Start-Process powershell -ArgumentList "-NoProfile", "-Command", "cd `"$repo`"; $env:WEB6_GRAPH_PATH='`"$graphPath`"'; $env:WEB6_GRAPH_MODE='`"$graphMode`"'; $env:WEB6_ROOT='`"$repo`"'; go run rechain-ide/web6-3d/cmd/web6-3d/main.go" -WindowStyle Normal
Start-Process powershell -ArgumentList "-NoProfile", "-Command", "cd `"$repo`"; go run rechain-ide/quantum/cmd/quantum/main.go" -WindowStyle Normal
Start-Process powershell -ArgumentList "-NoProfile", "-Command", "cd `"$repo`"; go run rechain-ide/agents/cmd/agent-compiler/main.go" -WindowStyle Normal
