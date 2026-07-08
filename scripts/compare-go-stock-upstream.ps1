param(
    [string]$ProjectDir = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path,
    [string]$UpstreamDir = (Join-Path $env:TEMP "upstream-go-stock-dev"),
    [string]$UpstreamRemote = "https://gitcode.com/gh_mirrors/go/go-stock.git",
    [string]$UpstreamBranch = "dev",
    [switch]$Refresh
)

$ErrorActionPreference = "Stop"

$localGoStock = Join-Path $ProjectDir "go-stock"
if (-not (Test-Path -LiteralPath $localGoStock)) {
    throw "Local go-stock directory not found: $localGoStock"
}

if ($Refresh -or -not (Test-Path -LiteralPath $UpstreamDir)) {
    if (Test-Path -LiteralPath $UpstreamDir) {
        Remove-Item -LiteralPath $UpstreamDir -Recurse -Force
    }
    git clone --depth 1 --branch $UpstreamBranch $UpstreamRemote $UpstreamDir
} else {
    git -C $UpstreamDir fetch --depth 1 origin $UpstreamBranch
    git -C $UpstreamDir checkout $UpstreamBranch | Out-Null
    git -C $UpstreamDir reset --hard "origin/$UpstreamBranch" | Out-Null
}

$ignoreRegex = "(^|/)(\.git|logs|node_modules|dist)(/|$)" +
    "|\.(png|jpg|jpeg|gif|ico|icns|db|exe|dll|docx|pack|idx|rev)$"

function Get-RelativeFiles([string]$Root) {
    Get-ChildItem -LiteralPath $Root -Recurse -File -Force |
        ForEach-Object { $_.FullName.Substring($Root.Length).TrimStart("\") -replace "\\", "/" } |
        Where-Object { $_ -notmatch $ignoreRegex } |
        Sort-Object -Unique
}

$upstreamFiles = @(Get-RelativeFiles $UpstreamDir)
$localFiles = @(Get-RelativeFiles $localGoStock)

$upstreamSet = @{}
$upstreamFiles | ForEach-Object { $upstreamSet[$_] = $true }

$localSet = @{}
$localFiles | ForEach-Object { $localSet[$_] = $true }

$upstreamOnly = @($upstreamFiles | Where-Object { -not $localSet.ContainsKey($_) })
$localOnly = @($localFiles | Where-Object { -not $upstreamSet.ContainsKey($_) })
$common = @($upstreamFiles | Where-Object { $localSet.ContainsKey($_) })

$changed = New-Object System.Collections.Generic.List[string]
foreach ($path in $common) {
    $upstreamPath = Join-Path $UpstreamDir ($path -replace "/", "\")
    $localPath = Join-Path $localGoStock ($path -replace "/", "\")
    $upstreamHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $upstreamPath).Hash
    $localHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $localPath).Hash
    if ($upstreamHash -ne $localHash) {
        $changed.Add($path)
    }
}

$upstreamCommit = git -C $UpstreamDir log --oneline -1
$upstreamTags = git -C $UpstreamDir tag --points-at HEAD

Write-Host "Upstream: $upstreamCommit"
if ($upstreamTags) {
    Write-Host "Tags: $($upstreamTags -join ', ')"
}
Write-Host "Local:    $localGoStock"
Write-Host ""
Write-Host "Counts"
Write-Host "  upstream files : $($upstreamFiles.Count)"
Write-Host "  local files    : $($localFiles.Count)"
Write-Host "  common files   : $($common.Count)"
Write-Host "  changed common : $($changed.Count)"
Write-Host "  upstream only  : $($upstreamOnly.Count)"
Write-Host "  local only     : $($localOnly.Count)"
Write-Host ""

Write-Host "Upstream-only sample"
$upstreamOnly | Select-Object -First 40 | ForEach-Object { Write-Host "  $_" }
Write-Host ""

Write-Host "Local-only sample"
$localOnly | Select-Object -First 40 | ForEach-Object { Write-Host "  $_" }
Write-Host ""

Write-Host "Changed common sample"
$changed | Select-Object -First 40 | ForEach-Object { Write-Host "  $_" }
