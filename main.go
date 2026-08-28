package main

import (
	"github.com/kraken-pedestal/cmd"
	"github.com/kraken-pedestal/pkg/logger"
	"log/slog"
	"os"
	"strings"
)

func main() {
	
	// ------------------------------------------------------------
	// Fallback logger
	//
	// 用于 PersistentPreRunE 之前发生的错误：
	// - unknown command
	// - flag parse error
	// - command initialization error
	// ------------------------------------------------------------
	logger.Init(logger.ConfigOptions{
		Level:      slog.LevelInfo,
		Color:      true,
		Source:     true,
		TimeFormat: "[ 06-01-02/15:04:05 ]",
	})
	
	// ------------------------------------------------------------
	// Panic recovery
	// ------------------------------------------------------------
	defer func() {
		if r := recover(); r != nil {
			logger.Default().Error("程序发生 Panic", "panic", r)
			os.Exit(1)
		}
	}()
	
	// ------------------------------------------------------------
	// Cobra
	// ------------------------------------------------------------
	
	rootCmd := cmd.NewRootCmd()
	
	if err := rootCmd.Execute(); err != nil {
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "unknown command"):
			slog.Error("未知命令，请使用 'kraken --help' 查看可用命令")
		case strings.Contains(errMsg, "unknown flag"):
			slog.Error("未知参数，请检查命令用法", "message", errMsg)
		default:
			slog.Error("执行命令失败", "message", errMsg)
		}
	}
}
