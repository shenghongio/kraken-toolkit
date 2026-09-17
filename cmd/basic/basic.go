package basic

import (
	"github.com/kraken-pedestal/pkg/cli"
	"github.com/spf13/cobra"
	"log/slog"
)

//NewBasicCmd Create the root Basic command.
//
// Basic is a functional domain of Kraken, responsible for host-level batch operations.
//
//For example:
//
//	kraken basic shell
//	kraken basic copy
//	kraken basic fetch
//	kraken basic script
//	kraken basic ping
//	kraken basic check

func NewBasicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "basic",
		Short:        "Basic Commands",
		GroupID:      cli.GroupBasic,
		SilenceUsage: true,
	}
	slog.Debug("register basic subcommands")

	shellCmd := NewShellCmd()
	if shellCmd == nil {
		slog.Error("shellCmd is nil")
	} else {
		slog.Debug("shell command", "use", shellCmd.Use)
	}

	addUserCmd := NewAddUserCmd()
	if addUserCmd == nil {
		slog.Debug("adduser command", "use", addUserCmd.Use)
	}

	// Register the shared parameters for all subcommands of the Basic module
	registerBasicFlags(cmd)

	// Register Basic sub command
	cmd.AddCommand(
		NewShellCmd(),
		NewAddUserCmd(),
		//NewCopyCmd(),
		//NewFetchCmd(),
		//NewScriptCmd(),
		//NewPingCmd(),
		//NewCheckCmd(),

	)
	return cmd
}
