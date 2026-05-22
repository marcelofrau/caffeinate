# build.ps1 — Build Caffeinate for Windows amd64
# Run this script from the repository root on a Windows machine with Go installed.
#
# Requirements:
#   - Go 1.21+
#   - Git (for go get)
#
# Optional (to embed manifest and suppress console window properly):
#   go install github.com/akavel/rsrc@latest
#   rsrc -manifest cmd/caffeinate/caffeinate.manifest -o cmd/caffeinate/rsrc.syso

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Write-Host "==> Downloading dependencies..." -ForegroundColor Cyan
go mod tidy

# Embed manifest and icon if rsrc is available
$rsrc = Get-Command rsrc -ErrorAction SilentlyContinue
if ($rsrc) {
    Write-Host "==> Embedding manifest and icon with rsrc..." -ForegroundColor Cyan
    rsrc -manifest cmd/caffeinate/caffeinate.manifest -ico icon/app_icon.ico -o cmd/caffeinate/rsrc.syso
}
else {
    Write-Host "    rsrc not found — skipping resource embedding (console may flash, no exe icon)." -ForegroundColor Yellow
    Write-Host "    Install with: go install github.com/akavel/rsrc@latest" -ForegroundColor Yellow
}

Write-Host "==> Building caffeinate.exe..." -ForegroundColor Cyan
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

# Ensure dist directory exists and place the built binary there
$dist = Join-Path $PSScriptRoot 'dist'
if (-not (Test-Path $dist)) {
    New-Item -ItemType Directory -Path $dist | Out-Null
}

$output = Join-Path $dist 'caffeinate.exe'

go build -ldflags="-H windowsgui -s -w" -o $output ./cmd/caffeinate

Write-Host "==> Done: $output" -ForegroundColor Green
