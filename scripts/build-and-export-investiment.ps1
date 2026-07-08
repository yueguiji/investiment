param(
    [Parameter(Mandatory=$true)]
    [string]$ProjectDir,

    [Parameter(Mandatory=$true)]
    [string]$ExportDir,

    [string]$WailsVersion = "v2.12.0"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path -LiteralPath $ProjectDir)) {
    throw "ProjectDir does not exist: $ProjectDir"
}

$wailsConfig = Join-Path $ProjectDir "wails.json"
if (-not (Test-Path -LiteralPath $wailsConfig)) {
    throw "wails.json not found in: $ProjectDir"
}

$configRaw = [System.IO.File]::ReadAllText($wailsConfig, [System.Text.Encoding]::UTF8)
$config = $configRaw | ConvertFrom-Json
$outputFileName = $config.outputfilename
if (-not $outputFileName) {
    throw "outputfilename is missing in wails.json"
}

$cacheRoot = Join-Path $ProjectDir ".buildcache"
$tempDir = Join-Path $cacheRoot "temp"
$goCache = Join-Path $cacheRoot "gocache"

@($ExportDir, $tempDir, $goCache) | ForEach-Object {
    New-Item -ItemType Directory -Force -Path $_ | Out-Null
}

$env:TEMP = $tempDir
$env:TMP = $tempDir
$env:GOCACHE = $goCache

$dbPatterns = @("*.db", "*.db-wal", "*.db-shm", "*.sqlite", "*.sqlite3")
$dbBackupFiles = @()
$dbProtected = $false

foreach ($pattern in $dbPatterns) {
    $dbFiles = Get-ChildItem -LiteralPath $ExportDir -Recurse -File -Filter $pattern -ErrorAction SilentlyContinue
    foreach ($dbFile in $dbFiles) {
        $relativePath = $dbFile.FullName.Substring($ExportDir.Length).TrimStart('\', '/')
        $dbBackup = Join-Path $env:TEMP ("db-backup\" + $relativePath)
        $dbBackupDir = Split-Path -Parent $dbBackup
        New-Item -ItemType Directory -Force -Path $dbBackupDir | Out-Null
        Copy-Item -LiteralPath $dbFile.FullName -Destination $dbBackup -Force
        $dbBackupFiles += @{Original=$dbFile.FullName; Backup=$dbBackup}
        $dbProtected = $true
        Write-Host "[PROTECT] Backed up existing database file: $($dbFile.FullName) -> $dbBackup"
    }
}

$frontendDist = Join-Path $ProjectDir "frontend\dist"
if (Test-Path -LiteralPath $frontendDist) {
    Write-Host "[SKIP] Frontend already built, skipping frontend build"
} else {
    Write-Host "Building frontend in $ProjectDir"
    Push-Location (Join-Path $ProjectDir "frontend")
    try {
        npm install
        npm run build
    } finally {
        Pop-Location
    }
}

$buildCommand = @()
if (Get-Command wails -ErrorAction SilentlyContinue) {
    $buildCommand = @("wails", "build")
} else {
    $buildCommand = @("go", "run", "github.com/wailsapp/wails/v2/cmd/wails@$WailsVersion", "build")
}

Write-Host "Building Wails app in $ProjectDir"
Push-Location $ProjectDir
try {
    & $buildCommand[0] $buildCommand[1..($buildCommand.Length - 1)]
    if ($LASTEXITCODE -ne 0) {
        throw "Wails build failed with exit code $LASTEXITCODE"
    }
} finally {
    Pop-Location
}

$sourceExe = Join-Path $ProjectDir ("build\bin\" + $outputFileName + ".exe")
if (-not (Test-Path -LiteralPath $sourceExe)) {
    throw "Build completed but exe not found: $sourceExe"
}

$targetExe = Join-Path $ExportDir ([IO.Path]::GetFileName($sourceExe))
try {
    Copy-Item -LiteralPath $sourceExe -Destination $targetExe -Force
} catch {
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $fallbackExe = Join-Path $ExportDir ($outputFileName + "-" + $timestamp + ".exe")
    Copy-Item -LiteralPath $sourceExe -Destination $fallbackExe -Force
    $targetExe = $fallbackExe
}

if ($dbProtected) {
    foreach ($pair in $dbBackupFiles) {
        Copy-Item -LiteralPath $pair.Backup -Destination $pair.Original -Force
        Remove-Item -LiteralPath $pair.Backup -Force
        Write-Host "[PROTECT] Restored database after build: $($pair.Original)"
    }
}

Write-Host ""
Write-Host "=== Build & Export Complete ==="
Write-Host "Project exe: $sourceExe"
Write-Host "Exported exe: $targetExe"
Write-Host "Data dir:     $ExportDir"
if ($dbProtected) {
    Write-Host "Database:     PRESERVED (not overwritten)"
} else {
    Write-Host "Database:     Will be created on first run"
}
