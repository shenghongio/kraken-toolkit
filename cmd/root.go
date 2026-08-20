package cmd

import (
	_ "embed"
	"fmt"
	"github.com/kraken-pedestal/internal/deploy/executor"
	"github.com/kraken-pedestal/pkg/logger"
	"github.com/spf13/cobra"
)

//go:embed templates/logo.txt
var AsciiLogo string

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
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion, _ := cmd.Flags().GetBool("version"); showVersion {
				fmt.Println(executor.PrintString())
				return nil
			}
			fmt.Print(AsciiLogo)
			return cmd.Help()
		},
	}
	
	cmd.Flags().BoolP("version", "v", false, "显示版本信息并退出")
	
	cmd.PersistentFlags().String("log-level", "info", "日志级别 (debug, info, warn, error)")
	cmd.PersistentFlags().String("log-output", "stderr", "日志输出目标 (stdout, stderr, 或文件路径)")
	cmd.PersistentFlags().String("console-format", "text", "控制台输出格式 (text, json)")
	cmd.PersistentFlags().Bool("log-source", false, "在日志中包含代码位置 (文件:行号)")
	
	cmd.AddCommand(VersionCmd())
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return cmd
}

func initLogger(cmd *cobra.Command) error {
	logLevel, err := cmd.Flags().GetString("log-level")
	if err != nil {
		return fmt.Errorf("获取 log-level 失败: %w", err)
	}
	logOutput, err := cmd.Flags().GetString("log-output")
	if err != nil {
		return fmt.Errorf("获取 log-output 失败: %w", err)
	}
	consoleFormat, err := cmd.Flags().GetString("console-format")
	if err != nil {
		return fmt.Errorf("获取 console-format 失败: %w", err)
	}
	logSource, err := cmd.Flags().GetBool("log-source")
	if err != nil {
		return fmt.Errorf("获取 log-source 失败: %w", err)
	}
	
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
	
	if consoleFormat != "text" && consoleFormat != "json" {
		return fmt.Errorf("不支持的 console-format: %s (可选: text, json)", consoleFormat)
	}
	
	cfg := logger.Config{
		Level:         level,
		Console:       false,
		ConsoleOutput: "",
		ConsoleFormat: consoleFormat,
		File:          "",
		AddSource:     logSource,
		MaxSize:       10,
		MaxBackups:    5,
		MaxAge:        30,
		Compress:      true,
	}
	
	switch logOutput {
	case "stdout", "stderr":
		cfg.Console = true
		cfg.ConsoleOutput = logOutput
	default:
		cfg.File = logOutput
	}
	
	if err := logger.Init(cfg); err != nil {
		return fmt.Errorf("日志初始化失败: %w", err)
	}
	
	logger.Debug("日志初始化成功", "level", logLevel, "output", logOutput, "console-format", consoleFormat, "source", logSource)
	logger.Warn("磁盘检测失败")
	logger.Error("初始化 sysctl 失败", "error", err)
	logger.Info("开始执行初始化系统任务")
	return nil
}
