package basic

import (
	"fmt"
	"github.com/kraken-pedestal/internal/basic/executor"
	"github.com/kraken-pedestal/internal/basic/inventory"
	execRuntime "github.com/kraken-pedestal/internal/basic/runtime"
	"github.com/spf13/cobra"
	"log/slog"
)

func NewShellCmd() *cobra.Command {
	var (
		command string
	)

	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Execute shell commands on hosts",

		RunE: func(cmd *cobra.Command, args []string) error {
			if command == "" {
				return fmt.Errorf("command is required")
			}

			sshOpt := executor.SSHOptions{
				User:     BasicFlags.SSHUser,
				Port:     BasicFlags.SSHPort,
				Host:     BasicFlags.SSHHost,
				Password: BasicFlags.SSHPassword,
			}

			hosts := inventory.Host{
				Address: sshOpt.Host,
				User:    sshOpt.User,
				Port:    sshOpt.Port,
				Passwd:  sshOpt.Password,
			}
			inv := inventory.NewFromHosts([]inventory.Host{hosts})

			// 创建executor 复用 sshOpt
			exec := executor.NewExecutor(executor.Options{
				Concurrency: 1,
				SSH:         sshOpt,
			})

			slog.Debug("executor created", "options", exec.Options())
			// runtime
			rt := execRuntime.New(inv, exec)
			slog.Debug("runtime created", "options", rt)

			// execute
			results := rt.Shell(cmd.Context(), command)

			// print result
			for _, result := range results {
				if result.Error != nil {
					fmt.Printf("[FAIL] %s: %v\n", result.Host.Address, result.Error)
					continue
				}
				fmt.Printf("[OK] %s\n%s\n", result.Host.Address, result.Stdout)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&command, "command", "c", "", "shell command")
	return cmd

}
