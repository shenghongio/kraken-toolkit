package cmd

import (
	_ "embed"
	"fmt"
	"github.com/kraken-pedestal/pkg/logger"
	"github.com/kraken-pedestal/pkg/printer"
	"github.com/spf13/cobra"
	"log/slog"
)

var AsciiLogo string

//func NewRootCmd() *cobra.Command {
//	cmd := &cobra.Command{
//		Use: "kraken",
//		Long: `Kraken is a unified CLI tool for deploying and operating Kubernetes clusters and middleware.
//
//  kraken dcli  - Deployment CLI (Build): Install K8s, deploy middleware
//  kraken ocli  - Operations CLI (Fix): Diagnose network, pods, and systems
//  kraken version  - Print kraken version information`,
//
//		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
//			return initLogger(cmd)
//		},
//		RunE: func(cmd *cobra.Command, args []string) error {
//			if showVersion, _ := cmd.Flags().GetBool("version"); showVersion {
//				fmt.Println(executor.PrintString())
//				return nil
//			}
//			fmt.Print(AsciiLogo)
//			return cmd.Help()
//		},
//	}
//
//	cmd.Flags().BoolP("version", "v", false, "Display detailed information of the current version")
//	cmd.PersistentFlags().String("log-level", "info", "日志级别 (debug, info, warn, error)")
//	cmd.PersistentFlags().Bool("log-source", false, "在日志中包含代码位置 (文件:行号)")
//
//	cmd.AddCommand(VersionCmd())
//	cmd.SilenceErrors = true
//	cmd.SilenceUsage = true
//	return cmd
//}

// Command Group
const (
	GroupDeploy          = "deploy"
	GroupCluster         = "cluster"
	GroupTroubleshooting = "troubleshooting"
	GroupNetwork         = "network"
	GroupBasic           = "basic"
	GroupSettings        = "settings"
	GroupOther           = "other"
)

// NewRootCmd create the root command
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "kraken",
		Short:         "Kraken is a unified CLI tool for deploying and operating Kubernetes clusters and middleware.",
		SilenceErrors: true,
		SilenceUsage:  true,
		//Long:          `Kraken is a unified CLI tool for deploying and operating Kubernetes clusters and middleware.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},

		// Global logger initialization
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initLogger(cmd)
		},
	}

	// Disable Cobra's auto-generated completion
	cmd.CompletionOptions.DisableDefaultCmd = true

	// Global Flags
	cmd.PersistentFlags().String(
		"log-level",
		"info",
		"log level [debug|info|warn|error|fatal]",
	)
	cmd.PersistentFlags().Bool(
		"log-source",
		false,
		"The log includes code locations (file:line number)",
	)

	// Command Groups
	cmd.AddGroup(
		//&cobra.Group{ID: GroupDeploy, Title: "Deploy Commands:"},
		&cobra.Group{
			ID:    GroupBasic,
			Title: "Basic Commands",
		},
		&cobra.Group{
			ID:    GroupSettings,
			Title: "Settings Commands",
		},
		&cobra.Group{
			ID:    GroupOther,
			Title: "Other Commands",
		},
	)

	// Commands
	cmd.AddCommand(
		NewBasicCmd(),
		VersionCmd(),
		NewCompletionCmd(),
	)

	// Custom help information
	cmd.SetHelpFunc(func(command *cobra.Command, args []string) {
		out := cmd.OutOrStdout()

		// root command
		if command == command.Root() {
			// Description
			fmt.Fprintln(out, cmd.Short)
			fmt.Fprintln(out)

			printer.PrintRootHelp(out, command)
			return
		}

		// Subcommand
		printer.PrintSubCommandHelp(out, command)
	})
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
