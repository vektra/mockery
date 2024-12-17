package main

import (
	"os"

	"github.com/vektra/mockery/tools/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
