param(
  [string]$SchemaDir = "schemas"
)

$files = Get-ChildItem -Path $SchemaDir -Filter *.json -File
$errors = @()

foreach ($f in $files) {
  try {
    $json = Get-Content -Path $f.FullName -Raw | ConvertFrom-Json
    if (-not $json.PSObject.Properties.Name.Contains('schema_version')) {
      $errors += "$($f.Name): missing schema_version"
    }
  } catch {
    $errors += "$($f.Name): invalid JSON"
  }
}

if ($errors.Count -gt 0) {
  Write-Host "Schema validation failed:" -ForegroundColor Red
  $errors | ForEach-Object { Write-Host "- $_" }
  exit 1
}

Write-Host "Schema validation passed." -ForegroundColor Green
