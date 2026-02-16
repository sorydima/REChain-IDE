# Demo flow

$ErrorActionPreference = 'Stop'

./scripts/dev.ps1
Start-Sleep -Seconds 2

# index minimal set
$files = "README.md"
$body = "{\"schema_version\":\"0.1.0\",\"repo\":\"demo\",\"files\":[\"$files\"]}"
Invoke-RestMethod -Method Post -Uri http://localhost:8083/index -ContentType application/json -Body $body | Out-Null

# submit task
$task = "{\"schema_version\":\"0.1.0\",\"type\":\"patch\",\"input\":\"add logging\",\"context\":[],\"constraints\":[],\"metadata\":{\"requester\":\"cli\",\"priority\":\"normal\"}}"
$status = Invoke-RestMethod -Method Post -Uri http://localhost:8081/tasks -ContentType application/json -Body $task
$taskId = $status.id
Write-Host "Task: $taskId"

# poll
for ($i=0; $i -lt 10; $i++) {
  $s = Invoke-RestMethod -Uri "http://localhost:8081/tasks/$taskId"
  if ($s.state -ne "running") { break }
  Start-Sleep -Milliseconds 200
}

$result = Invoke-RestMethod -Uri "http://localhost:8081/tasks/$taskId/result"
Write-Host "Result diff:"
Write-Host $result.diff
