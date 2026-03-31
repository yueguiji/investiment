param(
    [switch]$OpenBrowser
)

$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot
$logDir = Join-Path $projectRoot "build\\dev-logs"
$stdout = Join-Path $logDir "wails-go-stock.out.log"
$stderr = Join-Path $logDir "wails-go-stock.err.log"

function Ensure-Directory {
    param([string]$PathValue)
    if (-not (Test-Path -LiteralPath $PathValue)) {
        New-Item -ItemType Directory -Path $PathValue | Out-Null
    }
}

function Resolve-WailsExecutable {
    $wailsCmd = Get-Command wails -ErrorAction SilentlyContinue
    if ($wailsCmd) {
        return if ($wailsCmd.Source) { $wailsCmd.Source } elseif ($wailsCmd.Path) { $wailsCmd.Path } else { $wailsCmd.Definition }
    }

    $goBin = go env GOBIN 2>$null
    if ($LASTEXITCODE -eq 0 -and $goBin) {
        $candidate = Join-Path $goBin.Trim() "wails.exe"
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    $goPath = go env GOPATH 2>$null
    if ($LASTEXITCODE -eq 0 -and $goPath) {
        $candidate = Join-Path $goPath.Trim() "bin\\wails.exe"
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    throw "Wails CLI is not installed or not in PATH. Install it with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
}

if (Test-Path $stdout) { Remove-Item $stdout -Force }
if (Test-Path $stderr) { Remove-Item $stderr -Force }
Ensure-Directory $logDir

Set-Location $projectRoot

$arguments = @("dev")
if ($OpenBrowser) {
    $arguments += "-browser"
}

$wailsExecutable = Resolve-WailsExecutable
Start-Process -FilePath $wailsExecutable `
  -ArgumentList $arguments `
  -WorkingDirectory $projectRoot `
  -RedirectStandardOutput $stdout `
  -RedirectStandardError $stderr
