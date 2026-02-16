# Check service health

$services = @(
  @{name="orchestrator"; url="http://localhost:8081/health"},
  @{name="kernel"; url="http://localhost:8082/health"},
  @{name="rag"; url="http://localhost:8083/health"},
  @{name="web6-3d"; url="http://localhost:8084/health"},
  @{name="quantum"; url="http://localhost:8085/health"},
  @{name="agent-compiler"; url="http://localhost:8086/health"}
)

foreach ($s in $services) {
  try {
    $resp = Invoke-RestMethod -Uri $s.url -TimeoutSec 2
    if ($resp -eq "ok") {
      Write-Host "$($s.name): 200" -ForegroundColor Green
    } else {
      Write-Host "$($s.name): up" -ForegroundColor Green
    }
  } catch {
    Write-Host "$($s.name): down" -ForegroundColor Red
  }
}
