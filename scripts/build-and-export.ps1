param(
    [string]$OutputDir,
    [string]$ShortcutName = "rubin_investment.lnk",
    [switch]$SkipShortcut
)

$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot
$exeName = "investment-platform.exe"

if ([string]::IsNullOrWhiteSpace($OutputDir)) {
    $OutputDir = Join-Path $projectRoot "build\\export"
}

$desktopDir = [Environment]::GetFolderPath("Desktop")
$shortcutPath = Join-Path $desktopDir $ShortcutName

Set-Location $projectRoot

$wailsCmd = Get-Command wails -ErrorAction SilentlyContinue
if (-not $wailsCmd) {
    $goBin = go env GOPATH
    if ($LASTEXITCODE -eq 0 -and $goBin) {
        $candidate = Join-Path $goBin.Trim() "bin\wails.exe"
        if (Test-Path $candidate) {
            $wailsCmd = Get-Item $candidate
        }
    }
}
if (-not $wailsCmd) {
    throw "Wails CLI is not installed or not in PATH."
}

Write-Host "Building Wails app..."
$wailsExecutable = if ($wailsCmd.Source) { $wailsCmd.Source } elseif ($wailsCmd.Path) { $wailsCmd.Path } elseif ($wailsCmd.FullName) { $wailsCmd.FullName } else { $null }
if (-not $wailsExecutable) {
    throw "Unable to resolve Wails CLI executable path."
}
& $wailsExecutable build
if ($LASTEXITCODE -ne 0) {
    throw "wails build failed."
}

$builtExe = Join-Path $projectRoot "build\bin\$exeName"
if (-not (Test-Path $builtExe)) {
    throw "Built executable not found: $builtExe"
}

New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null
$installedExe = Join-Path $OutputDir $exeName
Copy-Item -Path $builtExe -Destination $installedExe -Force

if (-not $SkipShortcut) {
    $shell = New-Object -ComObject WScript.Shell
    $shortcut = $shell.CreateShortcut($shortcutPath)
    $shortcut.TargetPath = $installedExe
    $shortcut.WorkingDirectory = $OutputDir
    $shortcut.IconLocation = $installedExe
    $shortcut.Description = "Rubin Investment"
    $shortcut.Save()
}

Write-Host "Installed exe: $installedExe"
if (-not $SkipShortcut) {
    Write-Host "Desktop shortcut: $shortcutPath"
}
