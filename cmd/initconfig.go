package cmd

import (
	"fmt"
	"github.com/kraken-pedestal/pkg/cli"
	"github.com/spf13/cobra"
	"os"
)

func NewCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh]",
		Short: "Generate the autocompletion script for the specified shell",
		Long: `Generate shell autocompletion scripts for kraken.

Supported shells:
  - bash
  - zsh

Example:
  kraken completion bash > /etc/bash_completion.d/kraken
  kraken completion zsh > "${fpath[1]}/_kraken"
`,
		GroupID: cli.GroupSettings,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("requires a shell type (bash, zsh)")
				// 具体生成代码，可以调用 cmd.Root().GenBashCompletion(os.Stdout) 等
			}
			shellType := args[0]
			root := cmd.Root()
			switch shellType {
			case "bash":
				return root.GenBashCompletion(os.Stdout)
			case "zsh":
				return root.GenZshCompletion(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell type %q", shellType)
			}
		},
	}
	cmd.ValidArgs = []string{"bash", "zsh"}
	return cmd
}
