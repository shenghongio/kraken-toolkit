package cmd

import (
	_ "embed"
	"fmt"
	"github.com/kraken-pedestal/pkg/logger"
	"github.com/spf13/cobra"
)

//go:embed templates/logo.txt
var AsciiLogo string
var Version = "dev"

// NewRootCmd 创建根命令
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kraken",
		Version: Version,
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
			//fmt.Print(AsciiLogo)
			// 显示帮助信息或执行其他默认逻辑
			return cmd.Help()
		},
	}
	
	// 日志相关标志
	cmd.PersistentFlags().String("log-level", "info", "日志级别 (debug, info, warn, error)")
	cmd.PersistentFlags().String("log-format", "text", "日志格式 (text, json)")
	cmd.PersistentFlags().Bool("no-color", false, "禁用彩色输出")
	cmd.PersistentFlags().String("log-output", "stderr", "日志输出目标 (stdout, stderr, 或文件路径)")
	cmd.PersistentFlags().Bool("log-source", false, "在日志中包含代码位置 (文件:行号)")
	//cmd.PersistentFlags().String("stol", "stol", "系统工具集")
	// 添加子命令
	cmd.AddCommand(VersionCmd())
	//cmd.AddCommand(stol.NewStolCommand())
	// 系统工具集
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return cmd
}

// initLogger  从命令行标志初始化日志系统
func initLogger(cmd *cobra.Command) error {
	// 获取标志值
	// 统一使用 PersistentFlags获取
	logLevel, err := cmd.Flags().GetString("log-level")
	if err != nil {
		return fmt.Errorf("获取 log-level 失败: %w", err)
	}
	logFormat, err := cmd.Flags().GetString("log-format")
	if err != nil {
		return fmt.Errorf("获取 log-format 失败: %w", err)
	}
	
	logOutput, err := cmd.Flags().GetString("log-output")
	if err != nil {
		return fmt.Errorf("获取 log-output 失败: %w", err)
	}
	
	noColor, err := cmd.Flags().GetBool("no-color")
	if err != nil {
		return fmt.Errorf("获取 no-color 标志失败: %w", err)
	}
	
	logSource, err := cmd.Flags().GetBool("log-source")
	if err != nil {
		return fmt.Errorf("获取 log-source 失败: %w", err)
	}
	
	//解析日志级别
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
		return fmt.Errorf("不支持的日志级别: %s (可选: debug, info, warn, error)", logLevel)
	}
	
	// 解析格式
	var format logger.LogFormat
	switch logFormat {
	case "json":
		format = logger.FormatJSON
	case "text":
		format = logger.FormatText
	default:
		return fmt.Errorf("不支持的日志格式: %s (可选: text, json)", logFormat)
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
		return fmt.Errorf("日志初始化失败: %w", err)
	}
	
	logger.Debug("日志初始化成功",
		"level", logLevel,
		"format", logFormat,
		"output", logOutput,
		"source", logSource,
	)
	return nil
}
