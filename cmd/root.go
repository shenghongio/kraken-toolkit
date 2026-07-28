package cmd

import (
	_ "embed"
	"fmt"
	"github.com/kraken-pedestal/pkg/logger"
	"github.com/spf13/cobra"
)

//go:embed templates/logo.txt
var AsciiLogo string

// NewRootCmd 创建根命令
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "kraken",
		Long: `Kraken is a unified CLI tool for deploying and operating Kubernetes clusters and middleware.
  
  kraken dcli  - Deployment CLI (Build): Install K8s, deploy middleware
  kraken ocli  - Operations CLI (Fix): Diagnose network, pods, and systems
  kraken version  - Print kraken version information`,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initLogger(cmd)
		},
		// 仅在根命令（无子命令）时执行
		RunE: func(cmd *cobra.Command, args []string) error {
			// 只在根命令时打印 Logo（建议）
			fmt.Print(AsciiLogo)
			// 显示帮助信息或执行其他默认逻辑
			return cmd.Help()
		},
	}

	// 日志相关标志
	cmd.PersistentFlags().String("log-level", "info", "日志级别 (debug, info, warn, error)")
	cmd.PersistentFlags().String("log-format", "text", "日志格式 (text, json)")
	cmd.PersistentFlags().Bool("no-color", true, "禁用彩色输出")
	cmd.PersistentFlags().String("log-output", "stderr", "日志输出目标 (stdout, stderr, 或文件路径)")
	cmd.PersistentFlags().Bool("log-source", false, "在日志中包含代码位置 (文件:行号)")
	cmd.PersistentFlags().String("version", "version", "显示版本信息")
	//cmd.PersistentFlags().String("stol", "stol", "系统工具集")
	// 添加子命令
	cmd.AddCommand(VersionCmd())
	//cmd.AddCommand(stol.NewStolCommand())
	// 系统工具集
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return cmd
}

// initLogger 初始化日志系统
func initLogger(cmd *cobra.Command) error {
	// 获取标志值
	logLevel, _ := cmd.Flags().GetString("log-level")
	logFormat, _ := cmd.Flags().GetString("log-format")
	noColor, _ := cmd.Flags().GetBool("no-color")
	logOutput, _ := cmd.Flags().GetString("log-output")
	logSource, _ := cmd.Flags().GetBool("log-source")

	// 解析日志级别
	var level logger.LogLevel
	switch logLevel {
	case "debug":
		level = logger.LevelDebug
	case "info":
		level = logger.LevelInfo

	case "warn":
		level = logger.LevelWarn
	case "error":
		level = logger.LevelError
	default:
		level = logger.LevelInfo
	}

	// 解析格式
	var format logger.LogFormat
	switch logFormat {
	case "json":
		format = logger.FormatJSON
	default:
		format = logger.FormatText

	}
	// 创建日志配置
	cfg := &logger.Config{
		Level:      level,
		Format:     format,
		NoColor:    noColor,
		OutputPath: logOutput,
		AddSource:  logSource,
		MaxSize:    10, // 默认10MB滚动
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	// 验证配置
	if err := logger.Init(cfg); err != nil {
		return err
	}

	logger.Debug("日志初始化成功",
		"level", logLevel,
		"format", logFormat,
		"output", logOutput,
		"source", logSource,
	)
	return nil
}
