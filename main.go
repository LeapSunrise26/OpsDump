package main

import (
	"os"

	"ops-dump/internal/cmd"
	_ "ops-dump/internal/packed"

	"github.com/gogf/gf/v2/os/gres"
)

func main() {
	// gres.Dump lists every embedded resource; only useful when debugging
	// the gf pack payload, and it floods the desktop sidecar log otherwise.
	if os.Getenv("OPSDUMP_DEBUG") != "" {
		gres.Dump()
	}

	cmd.Main()
}
