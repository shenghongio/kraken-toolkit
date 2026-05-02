package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sysint/internal/logger"
)

// NewRootCmd 创建根命令
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sysint",
		Short: "系统初始化检查工具",
		Long: `系统初始化检查工具 - 用于检查和初始化系统环境

	支持功能:
		- 系统初始化检查(OS版本、磁盘、内存、CPU等)
		- 批量命令执行
		- 配置文件驱动
		- 多种输出格式`,
	}
	
	// 添加全局日志标志
	cmd.PersistentFlags().String("log-level", "info", "日志级别 (debug, info, warn, error)")
	cmd.PersistentFlags().Bool("log-no-color", false, "禁用彩色输出")
	cmd.PersistentFlags().String("log-output", "stdout", "日志输出 (stdout, stderr)")
	cmd.PersistentFlags().String("log-user", "", "日志中显示的用户名")
	
	// 添加子命令
	cmd.AddCommand(versionCmd())
	
	// 可以添加其他命令
	// cmd.AddCommand(checkCmd())
	
	return cmd
}

// initLogger 初始化日志系统
func initLogger(cmd *cobra.Command) error {
	// 获取标志值
	logLevel, _ := cmd.Flags().GetString("log-level")
	logNoColor, _ := cmd.Flags().GetBool("log-no-color")
	logOutput, _ := cmd.Flags().GetString("log-output")
	logUser, _ := cmd.Flags().GetString("log-user")
	
	// 创建日志配置
	config := &logger.LogConfig{
		Level:   logLevel,
		NoColor: logNoColor,
		Output:  logOutput,
		User:    logUser,
	}
	
	// 验证配置
	if err := config.Validate(); err != nil {
		return err
	}
	
	// 初始化日志系统
	if err := logger.Init(config); err != nil {
		return err
	}
	
	logger.Info("日志系统初始化成功",
		"level", logLevel,
		"no_color", logNoColor,
		"output", logOutput,
		"user", logUser)
	
	return nil
}
