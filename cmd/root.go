package cmd

import (
	"fmt"
	"github.com/kraken-pedestal/cmd/basic"
	"github.com/kraken-pedestal/pkg/cli"
	"github.com/kraken-pedestal/pkg/logger"
	"github.com/kraken-pedestal/utils"
	"github.com/spf13/cobra"
	"log/slog"
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
			// 1.初始化日志
			if err := initLogger(cmd); err != nil {
				return err
			}

			//2. 加载全局配置
			//if err := loadConfig(); err != nil {
			//	return err
			//}
			return nil
		},

		// 加载全局配置文件

	}

	// Disable Cobra's auto-generated completion
	cmd.CompletionOptions.DisableDefaultCmd = true

	// Global Flags

	//cmd.PersistentFlags().StringVar(
	//	&globalFlags.Config,
	//	"config",
	//	"",
	//	"config file path",
	//)
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
			ID:    cli.GroupBasic,
			Title: "Basic Commands",
		},
		&cobra.Group{
			ID:    cli.GroupSettings,
			Title: "Settings Commands",
		},
		&cobra.Group{
			ID:    cli.GroupOther,
			Title: "Other Commands",
		},
	)

	// Commands
	cmd.AddCommand(
		basic.NewBasicCmd(),
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
			utils.PrintRootHelp(out, command)
			return
		}

		// Subcommand
		utils.PrintCommandHelp(out, command)
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
