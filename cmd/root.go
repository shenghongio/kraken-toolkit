package cmd

import (
	_ "embed"
	"fmt"
	"github.com/kraken-pedestal/internal/deploy/executor"
	"github.com/kraken-pedestal/pkg/logger"
	"github.com/spf13/cobra"
	"log/slog"
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
	
	cmd.Flags().BoolP("version", "v", false, "Display detailed information of the current version")
	cmd.PersistentFlags().String("log-level", "info", "日志级别 (debug, info, warn, error)")
	cmd.PersistentFlags().Bool("log-source", false, "在日志中包含代码位置 (文件:行号)")
	
	cmd.AddCommand(VersionCmd())
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return cmd
}

// 初始化日志
func initLogger(cmd *cobra.Command) error {
	getLevel, err := cmd.Flags().GetString("log-level")
	if err != nil {
		return fmt.Errorf("get log level: %w", err)
	}
	level, err := logger.ParseLevel(getLevel)
	if err != nil {
		return err
	}
	addSource, err := cmd.Flags().GetBool("log-source")
	if err != nil {
		return fmt.Errorf("get log source: %w", err)
	}
	
	logger.Init(logger.ConfigOptions{
		Level:      level,
		Color:      true,
		Source:     addSource,
		TimeFormat: "[ 06-01-02/15:04:05 ]",
	})
	slog.Debug("logger initialization completed", "level", level.String(), "source", addSource)
	return nil
}
