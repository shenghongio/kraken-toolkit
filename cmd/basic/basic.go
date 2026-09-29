package basic

import (
	"github.com/kraken-toolkit/pkg/cli"
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
//	kraken basic scripts
//	kraken basic ping
//	kraken basic check

func NewBasicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bc",
		Short: "Basic Commands",
		Long: "Host-level batch operations:\n" +
			"    * cmd      - Execute shell commands on remote hosts\n" +
			"    * script   - Execute local script files on remote hosts\n" +
			"    * adduser  - Create management user and deploy SSH key",
		Example: "kraken bc cmd -c \"uptime\" --ssh-host=10.32.9.138\n" +
			"  kraken bc script deploy.sh --config kraken.yaml",
		GroupID:      cli.GroupBasic,
		SilenceUsage: true,
	}
	slog.Debug("register basic subcommands")
	registerBasicFlags(cmd)

	// Register Basic sub command
	cmd.AddCommand(
		NewCmd(),
		NewAddUserCmd(),
		NewScriptCmd(),
		//NewFetchCmd(),
		//NewScriptCmd(),
		//NewPingCmd(),
		//NewCheckCmd(),

	)
	return cmd
}
