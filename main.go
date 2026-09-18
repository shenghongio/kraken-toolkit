package main

import (
	"github.com/kraken-pedestal/cmd"
	"github.com/kraken-pedestal/pkg/logger"
	"os"
	"runtime/debug"
)

func main() {
	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			logger.Default().Error("程序发生 Panic", "panic", r, "stack", string(debug.Stack()))
			os.Exit(1)
		}
	}()
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
