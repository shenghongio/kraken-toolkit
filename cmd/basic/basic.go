package basic

import (
	"github.com/kraken-pedestal/pkg/cli"
	"github.com/spf13/cobra"
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
		Use:     "basic",
		Short:   "Basic Commands",
		GroupID: cli.GroupBasic,
	}

	// Register the shared parameters for all subcommands of the Basic module
	registerBasicFlags(cmd)

	// Register Basic sub command
	cmd.AddCommand(
	//NewShellCmd(),
	//NewCopyCmd(),
	//NewFetchCmd(),
	//NewScriptCmd(),
	//NewPingCmd(),
	//NewCheckCmd(),

	)
	return cmd
}
