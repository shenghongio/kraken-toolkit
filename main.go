package main

import (
	"fmt"
	"github.com/kraken-pedestal/cmd"
	"github.com/kraken-pedestal/pkg/logger"
	"os"
	"strings"
)

func main() {
	
	// 1. 初始化后备日志（仅用于 PreRun 之前的错误，如未知命令、标志解析失败）
	// 注意：必须提供有效的级别和至少一个输出目标（控制台）
	if err := logger.Init(logger.Config{
		Level:         logger.LevelInfo,
		Console:       true,
		ConsoleOutput: "stderr",
		ConsoleFormat: logger.ConsoleFormatText,
		File:          "",
		AddSource:     false,
	}); err != nil {
		// 若后备日志初始化失败，只能 panic（日志是关键组件）
		panic(fmt.Sprintf("failed to init fallback logger: %v", err))
	}
	
	// 优雅处理 panic
	defer func() {
		if r := recover(); r != nil {
			// 尝试用logger记录 若logger不可用则降级到stderr
			if logger.Default() != nil {
				logger.Error("程序发生 panic", "panic", r)
			} else {
				fmt.Fprintf(os.Stderr, "panic: %v\n", r)
			}
			os.Exit(1)
		}
	}()
	
	rootCmd := cmd.NewRootCmd()
	
	if err := rootCmd.Execute(); err != nil {
		// 错误已在命令中通过 logger 记录（子命令中使用了 os.Exit），
		// 但根命令级别（如未知子命令、标志错误）会在这里被捕获。
		// 我们根据错误类型输出友好信息，避免重复。
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "unknown command"):
			logger.Error("未知命令，请使用 'kraken --help' 查看可用命令")
		case strings.Contains(errMsg, "unknown flag"):
			logger.Error("未知参数，请检查命令用法", "error", errMsg)
		default:
			logger.Error("执行命令失败", "error", errMsg)
		}
		os.Exit(1)
	}
}
