package main

import (
	"ops-dump/internal/cmd"
	_ "ops-dump/internal/packed"

	"github.com/gogf/gf/v2/os/gres"
)

func main() {
	gres.Dump()

	cmd.Main()
}
