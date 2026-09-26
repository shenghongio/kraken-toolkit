package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

const (
	GroupDeploy          = "deploy"
	GroupCluster         = "cluster"
	GroupTroubleshooting = "troubleshooting"
	GroupNetwork         = "network"
	GroupBasic           = "basic"
	GroupSettings        = "settings"
	GroupOther           = "other"
)

const SubCmdHelpTemplate = `Usage:
  {{.UseLine}}

  {{.Short}}

Examples:
  {{.Example}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

Global flags: Use 'kraken --help' for global options.
`

func PrintSubCmdHelp(cmd *cobra.Command) {
	if cmd.UseLine() != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "%s Usage:\n  %s\n", cmd.Short, cmd.UseLine())
	}
	if cmd.Long != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "\n  %s\n", cmd.Long)
	}
	if cmd.Example != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "\nExamples:\n  %s\n", cmd.Example)
	}
	if cmd.HasAvailableLocalFlags() {
		fmt.Fprintf(cmd.OutOrStdout(), "\nFlags:\n")
		fmt.Fprint(cmd.OutOrStdout(), cmd.LocalNonPersistentFlags().FlagUsages())
	}
	fmt.Fprintln(cmd.OutOrStdout(), "\nGlobal flags: Use 'kraken --help' for global options.")
}