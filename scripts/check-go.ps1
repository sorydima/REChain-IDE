# Check Go installation

try {
  $version = & go version
  Write-Host $version -ForegroundColor Green
} catch {
  Write-Host "Go is not installed or not in PATH." -ForegroundColor Red
  exit 1
}
