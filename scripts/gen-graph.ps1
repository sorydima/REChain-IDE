param(
  [string]$Root = ".",
  [string]$Out = ".web6-graph.json",
  [int]$MaxNodes = 2000
)

$rootPath = (Resolve-Path $Root).Path
$rootName = Split-Path $rootPath -Leaf

$nodes = New-Object System.Collections.Generic.List[object]
$edges = New-Object System.Collections.Generic.List[object]

function Add-Node {
  param([string]$Id, [string]$Label)
  $nodes.Add([pscustomobject]@{ id = $Id; label = $Label })
}

function Add-Edge {
  param([string]$From, [string]$To)
  $edges.Add([pscustomobject]@{ from = $From; to = $To })
}

Add-Node "." $rootName

$items = Get-ChildItem -Path $rootPath -Recurse -Force -ErrorAction SilentlyContinue
foreach ($item in $items) {
  if ($nodes.Count -ge $MaxNodes) { break }
  if ($item.FullName -like "*\.*") { continue }
  $rel = $item.FullName.Substring($rootPath.Length).TrimStart("\","/")
  if ($rel -eq "") { continue }
  $label = $item.Name
  Add-Node $rel $label
  $parent = Split-Path $rel -Parent
  if ($parent -eq "" -or $parent -eq $rel) { $parent = "." }
  Add-Edge $parent $rel
}

$graph = [pscustomobject]@{
  nodes = $nodes
  edges = $edges
}

$json = $graph | ConvertTo-Json -Depth 6
Set-Content -Path $Out -Value $json -Encoding UTF8
Write-Host "Graph written to $Out"
