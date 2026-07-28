package ocli

import "github.com/spf13/cobra"

var OcliCmd = &cobra.Command{
	Use:   "ocli",
	Short: "Operations CLI - Diagnostics and troubleshooting",
}

func init() {
	oc
}
