# Stop all REChain dev service windows

Get-Process | Where-Object { $_.ProcessName -eq "powershell" } | ForEach-Object {
  try {
    if ($_.MainWindowTitle -match "orchestrator|kernel|rag|web6-3d") {
      $_.CloseMainWindow() | Out-Null
    }
  } catch {}
}
