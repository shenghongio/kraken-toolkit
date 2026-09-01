package cmd

import (
	"github.com/kraken-pedestal/internal/diagnose/system"
	"github.com/kraken-pedestal/internal/models"
	"github.com/kraken-pedestal/pkg/logger"
	"github.com/kraken-pedestal/pkg/printer"
	"github.com/spf13/cobra"
)

func NewSystemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "system",
		Short:   "Inspect and diagnose system resources",
		GroupID: GroupSystem,

		Example: `  #  Display system information
kraken system info
# Display system information in JSON format
kraken system info --json
		`,
	}

	cmd.AddCommand(NewSystemInfo())
	return cmd
}

func NewSystemInfo() *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Display system information kraken system info",
		RunE: func(cmd *cobra.Command, args []string) error {
			systemInfo, err := system.GetSystemInfo()
			if err != nil {
				return err
			}
			return selectoroutput(systemInfo, output)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format. One of: json|yaml|table")
	return cmd
}

func selectoroutput(info *models.GetSystemInfoResult, output string) error {
	switch output {
	case "json":
		return printer.PrintJSON(info)
	case "yaml":
		return printer.PrintYAML(info)
	case "table":
		table := system.PrintSystemInfoTable(info)
		return printer.PrintTable(table)
	default:
		return logger.New(logger.CodeInvalidArgument, "不支持的输出格式:"+output)
	}
}
