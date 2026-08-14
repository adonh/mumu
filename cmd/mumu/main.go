package main

import (
	"os"
	"runtime"

	"github.com/adonh/mumu/cmd/mumu/cmd"
)

func main() {
	runtime.LockOSThread()

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
