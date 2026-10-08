# Build the self-contained Windows sidecar binary for the Electron desktop app.
#   1. gf pack  : embed resource/ (templates + static + vendor) into Go source
#   2. go build : cross-compile windows/amd64, CGO off (pure-Go sqlite)
#   3. cleanup  : remove the temporary pack file (also on failure)
#
# Usage: powershell -NoProfile -ExecutionPolicy Bypass -File hack/build-windows.ps1

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$packFile = "internal/packed/build_pack_data.go"

try {
    # gf pack prompts interactively when DST exists; remove it first.
    Remove-Item -LiteralPath $packFile -Force -ErrorAction SilentlyContinue

    gf pack resource $packFile --keepPath=true
    if ($LASTEXITCODE -ne 0) { throw "gf pack failed (exit $LASTEXITCODE)" }

    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"

    New-Item -ItemType Directory -Force -Path "dist/sidecar" | Out-Null
    go build -ldflags "-s -w" -o "dist/sidecar/ops-dump.exe" .
    if ($LASTEXITCODE -ne 0) { throw "go build failed (exit $LASTEXITCODE)" }
}
finally {
    Remove-Item -LiteralPath $packFile -Force -ErrorAction SilentlyContinue
}

Write-Host "=> dist/sidecar/ops-dump.exe"
