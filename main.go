package main

import (
	"fmt"
	"github.com/kraken-toolkit/cmd"
	"github.com/kraken-toolkit/pkg/logger"
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
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
