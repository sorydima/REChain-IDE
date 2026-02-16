# Local diagnostics for REChain IDE workspace.

param(
  [switch]$FixBom,
  [switch]$CheckServices,
  [switch]$RunTests
)

$ErrorActionPreference = "Stop"

$scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$repo = Resolve-Path (Join-Path $scriptRoot "..")
Set-Location $repo

function Write-Ok($msg) {
  Write-Host $msg -ForegroundColor Green
}

function Write-Warn($msg) {
  Write-Host $msg -ForegroundColor Yellow
}

function Get-BomFiles {
  $files = @("go.work")
  $files += Get-ChildItem -Path "rechain-ide" -Filter "go.mod" -Recurse | ForEach-Object { $_.FullName }
  $bom = @()
  foreach ($f in $files) {
    $path = Resolve-Path $f
    $bytes = [System.IO.File]::ReadAllBytes($path)
    if ($bytes.Length -ge 3 -and $bytes[0] -eq 0xEF -and $bytes[1] -eq 0xBB -and $bytes[2] -eq 0xBF) {
      $bom += $path
    }
  }
  return $bom
}

function Fix-BomFiles($paths) {
  $enc = New-Object System.Text.UTF8Encoding($false)
  foreach ($p in $paths) {
    $txt = [System.IO.File]::ReadAllText($p)
    [System.IO.File]::WriteAllText($p, $txt, $enc)
  }
}

Write-Host "== doctor =="

try {
  $goVersion = & go version
  Write-Ok $goVersion
} catch {
  Write-Host "Go is not installed or not in PATH." -ForegroundColor Red
  exit 1
}

$bomFiles = Get-BomFiles
if ($bomFiles.Count -eq 0) {
  Write-Ok "Encoding check: no UTF-8 BOM in go.work/go.mod files."
} else {
  Write-Warn ("Encoding check: BOM found in {0} file(s)." -f $bomFiles.Count)
  $bomFiles | ForEach-Object { Write-Host (" - " + $_) }
  if ($FixBom) {
    Fix-BomFiles $bomFiles
    Write-Ok "BOM removed from listed files."
  } else {
    Write-Warn "Run: ./scripts/doctor.ps1 -FixBom"
  }
}

if ($RunTests) {
  Write-Host "Running go tests across workspace modules..."
  $env:GOWORK = Join-Path $repo "go.work"
  $mods = Get-ChildItem -Path "rechain-ide" -Filter "go.mod" -Recurse | ForEach-Object { Split-Path $_.FullName -Parent }
  foreach ($m in $mods) {
    Write-Host ("== test " + $m + " ==")
    Push-Location $m
    & go test ./...
    $code = $LASTEXITCODE
    Pop-Location
    if ($code -ne 0) {
      exit $code
    }
  }
}

if ($CheckServices) {
  Write-Host "Checking local service health..."
  & "$scriptRoot\\status.ps1"
}

Write-Ok "doctor finished."
