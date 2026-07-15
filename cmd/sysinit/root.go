package sysinit

import (
	"fmt"
	"github.com/kraken-pedestal/internal/command"
	"github.com/kraken-pedestal/internal/logger"
	"github.com/spf13/cobra"
	"os"
)

// NewRootCmd 创建根命令
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kraken",
		Short: "系统初始化检查工具",
		Long: `系统初始化检查工具 - 用于检查和初始化系统环境

支持功能:
  - 版本信息查看 (version)
  - SSH密钥对生成 (gen-key)，支持 ECDSA (P256/P384/P521)
  - 统一的日志输出配置（级别、格式、颜色、文件滚动）

示例:
  # 查看版本
  kraken version

  # 生成 ECDSA P256 密钥对（默认）
  kraken gen-key

  # 生成 P384 密钥对，指定输出目录和注释
  kraken gen-key --curve P384 --output /root/.ssh --comment "prod-key"

  # 调试模式运行（显示详细日志和代码位置）
  kraken gen-key --log-level debug --log-source

  # JSON 格式日志输出到文件
  kraken gen-key --log-format json --log-output /var/log/kraken.log

更多帮助请使用子命令的 -h 选项，例如: kraken gen-key -h`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initLogger(cmd)
		},
	}

	// 日志相关标志
	cmd.PersistentFlags().String("log-level", "info", "日志级别 (debug, info, warn, error)")
	cmd.PersistentFlags().String("log-format", "text", "日志格式 (text, json)")
	cmd.PersistentFlags().Bool("no-color", false, "禁用彩色输出")
	cmd.PersistentFlags().String("log-output", "stderr", "日志输出目标 (stdout, stderr, 或文件路径)")
	cmd.PersistentFlags().Bool("log-source", false, "在日志中包含代码位置 (文件:行号)")
	cmd.PersistentFlags().String("ssh-key", "", "SSH私钥文件路径（用于批量执行认证）")
	cmd.PersistentFlags().String("version", "version", "显示版本信息")
	cmd.PersistentFlags().String("sysint", "sysint", "系统工具集")
	// 添加子命令
	cmd.AddCommand(command.VersionCmd())
	//cmd.AddCommand(command.GenKeyCmd()) // 生成SSH密钥对
	// command.AddCommand(checkCmd()) // 系统检查子命令
	// command.AddCommand(k8sDeployCmd()) // K8s 部署子命令
	// command.AddCommand(runCmd()) // 执行 shell / python 命令

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

// Execute 执行根命令
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "执行失败: %v\n", err)
		os.Exit(1)
	}
}
