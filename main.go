package main

import (
	"github.com/kraken-pedestal/cmd"
	"github.com/kraken-pedestal/pkg/logger"
	"os"
	"strings"
)

func main() {

	// 1. 初始化默认 logger（保证所有输出格式统一）
	if err := logger.Init(&logger.Config{
		Level:      logger.LevelInfo,
		Format:     logger.FormatText,
		OutputPath: "stderr",
		NoColor:    true, // 默认无颜色，后续可被覆盖
	}); err != nil {
		// 如果初始化失败，直接 panic（因为日志是关键组件）
		panic(err)
	}

	// 优雅处理 panic
	defer func() {
		if r := recover(); r != nil {
			logger.Error("程序发生 panic", "panic", r)
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
