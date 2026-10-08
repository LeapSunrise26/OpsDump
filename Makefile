ROOT_DIR    = $(shell pwd)
NAMESPACE   = "default"
DEPLOY_NAME = "template-single"
DOCKER_NAME = "template-single"

include ./hack/hack-cli.mk
include ./hack/hack.mk

# Build self-contained Windows sidecar binary for the Electron desktop app.
# Embeds resource/ (templates + static + vendor) via gf pack, then cross-compiles.
# Requires gf CLI on PATH (see `make cli`). Logic lives in hack/build-windows.ps1
# because the Windows make shell is cmd (no POSIX sh for env-var prefixing).
.PHONY: build-windows
build-windows:
	@powershell -NoProfile -ExecutionPolicy Bypass -File hack/build-windows.ps1