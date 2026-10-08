# Build the Windows desktop deliverable end to end:
#   1. build the Go sidecar (hack/build-windows.ps1)
#   2. run electron-builder to produce the NSIS installer
#
# Electron binaries and electron-builder's helper binaries (winCodeSign, nsis)
# are fetched from GitHub by default, which is unreliable from CN networks.
# Default both mirrors to npmmirror unless the caller already set them.
#
# Usage (from desktop/): powershell -NoProfile -ExecutionPolicy Bypass -File ../hack/build-desktop.ps1

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$desktop = Join-Path $root "desktop"

if (-not $env:ELECTRON_MIRROR) {
    $env:ELECTRON_MIRROR = "https://npmmirror.com/mirrors/electron/"
}
if (-not $env:ELECTRON_BUILDER_BINARIES_MIRROR) {
    $env:ELECTRON_BUILDER_BINARIES_MIRROR = "https://npmmirror.com/mirrors/electron-builder-binaries/"
}

# 1) sidecar (self-contained ops-dump.exe with embedded resource/)
& powershell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot "build-windows.ps1")
if ($LASTEXITCODE -ne 0) { throw "sidecar build failed (exit $LASTEXITCODE)" }

# 2) installer
& (Join-Path $desktop "node_modules\.bin\electron-builder.cmd") --win
if ($LASTEXITCODE -ne 0) { throw "electron-builder failed (exit $LASTEXITCODE)" }

Write-Host "=> dist/OpsDump Setup <version>.exe"
