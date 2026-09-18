package basic

import (
	"context"
	"fmt"
	"github.com/kraken-pedestal/internal/basic/executor"
	"github.com/kraken-pedestal/internal/basic/runner"
	"github.com/kraken-pedestal/internal/config"
	"github.com/kraken-pedestal/utils"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
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
				//cmd.Help()
				utils.PrintUsage(cmd, "Example: kraken basic shell -c \"uptime\" --ssh-host=10.32.9.138")
				return nil
			}
			// --ssh-user, 有值，走单机模式，负责批量处理模式
			//  有 --ssh-host	单机模式
			//  无 --ssh-host，但有 config + host_inventory: true	批量模式
			//  无 --ssh-host，无 config 或 host_inventory: false	打印提示，告诉用户怎么用
			if !cmd.Flags().Changed("ssh-host") {
				// 检查 config 配置文件是否启用 批量模式开关host_inventory=true
				cfg, ok := config.FromContext(cmd.Context())
				if ok && cfg.Basic.HostInventory {
					return runBatchMode(cmd, command)
				}
				// 没有 config 或 host_inventory != true
				fmt.Println("未检测到批量配置（host_inventory: true）")
				fmt.Println("请选择执行模式:")
				fmt.Println("  单机: kraken basic shell -c <cmd> --ssh-host=<IP>")
				fmt.Println("  批量: kraken basic shell -c <cmd> --config <file> (配置中设置 host_inventory: true)")
				return nil
			}
			return runSingleMode(cmd, command)
		},
	}

	cmd.Flags().StringVarP(&command, "command", "c", "", "shell command")
	return cmd

}

// 单机模式
func runSingleMode(cmd *cobra.Command, command string) error {
	sshOpt := executor.SSHOptions{
		User:     BasicFlags.SSHUser,
		Port:     BasicFlags.SSHPort,
		Password: BasicFlags.SSHPassword,
	}
	hosts := executor.Host{
		Address: BasicFlags.SSHHost,
		User:    BasicFlags.SSHUser,
		Port:    BasicFlags.SSHPort,
		Passwd:  BasicFlags.SSHPassword,
	}
	exec := executor.NewExecutor(executor.Options{
		Concurrency: 1,
		SSH:         sshOpt,
	})

	results := exec.Run(cmd.Context(), []executor.Host{hosts}, func(ctx context.Context, host executor.Host, client *ssh.Client) executor.Result {
		return executor.RunShell(ctx, host, client, command)
	})

	return utils.PrintResultAndCheck(results)
}

// 批量处理
func runBatchMode(cmd *cobra.Command, command string) error {
	cfg, ok := config.FromContext(cmd.Context())
	if !ok {
		return fmt.Errorf("batch mode requires config file (--config)")
	}

	runnerCfg, err := runner.BuildFromConfig(cfg)
	if err != nil {
		return err
	}
	slog.Debug("batch mode", "hosts", len(runnerCfg.Hosts), "concurrency", runnerCfg.Concurrency)
	results := runner.Run(cmd.Context(), runnerCfg, func(ctx context.Context, host executor.Host, client *ssh.Client) executor.Result {
		return executor.RunShell(ctx, host, client, command)
	})
	return utils.PrintResultAndCheck(results)
}
