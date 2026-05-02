package main

import (
	"fmt"
	"github.com/sysint/internal/cmd"
	"github.com/sysint/internal/logger"
	"os"
)

func main() {

	// 优雅处理panic
	defer func() {
		if r := recover(); r != nil {
			logger.Error("程序发生panic", "panic", r)
			fmt.Fprintf(os.Stderr, "致命错误: %v\n", r)
			os.Exit(1)
		}
	}()

	rootCmd := cmd.NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		// 错误已经在命令中记录，这里只输出到stderr
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
