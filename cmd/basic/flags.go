package basic

import "github.com/spf13/cobra"

type Flags struct {
	SSHUser     string
	SSHPort     int
	SSHPassword string
	SSHHost     string
}

var BasicFlags Flags

// These parameters are registered via basic.PersistentFlags(),
// All basic subcommands are available.

func registerBasicFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	// ssh 相关参数
	flags.StringVar(
		&BasicFlags.SSHUser,
		"ssh-user",
		"root",
		"ssh username",
	)

	flags.IntVar(
		&BasicFlags.SSHPort,
		"ssh-port",
		22,
		"ssh port",
	)

	flags.StringVar(
		&BasicFlags.SSHPassword,
		"ssh-password",
		"",
		"ssh password",
	)
	flags.StringVar(
		&BasicFlags.SSHHost,
		"ssh-host",
		"",
		"ssh host",
	)
}
