package basic

import (
	"context"
	"fmt"
	"github.com/kraken-toolkit/internal/basic/executor"
	"github.com/kraken-toolkit/internal/config"
	"github.com/kraken-toolkit/pkg/cli"
	"github.com/kraken-toolkit/utils"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func NewCmd() *cobra.Command {
	var (
		command string
	)

	cmd := &cobra.Command{
		Use:   "cmd",
		Short: "Execute commands on hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cli.PrintSubCmdHelp(cmd)
				return nil
			}
			if command == "" {
				cli.PrintSubCmdHelp(cmd)

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
				// 用户提供了 SSH 凭据但漏了 --ssh-host，给出针对性提示
				if cmd.Flags().Changed("ssh-user") || cmd.Flags().Changed("ssh-password") {
					fmt.Println("缺少 --ssh-host 参数")
					fmt.Println("  单机模式需要指定目标主机: --ssh-host=<IP>")
					return nil
				}
				// 没有 config 或 host_inventory != true
				fmt.Println("未检测到批量配置（host_inventory: true）")
				fmt.Println("请选择执行模式:")
				fmt.Println("  单机: kraken bc cmd -c <cmd> --ssh-host=<IP>")
				fmt.Println("  批量: kraken bc cmd -c <cmd> --config <file> (配置中设置 host_inventory: true)")
				return nil
			}
			return runSingleMode(cmd, command)
		},
	}

	cmd.Flags().StringVarP(&command, "command", "c", "", "shell command")
	cmd.Long = "Only supports shell command mode, not script files.\n" +
		"    * For scripts, use: kraken bc script <file>"
	cmd.Example = "kraken bc cmd -c \"uptime\" --ssh-host=10.32.9.138"
	cmd.SetHelpTemplate(cli.SubCmdHelpTemplate)
	return cmd

}

// 单机模式
func runSingleMode(cmd *cobra.Command, command string) error {
	exec, host := NewSingleExecutor()
	results := exec.Run(cmd.Context(), []executor.Host{host}, func(ctx context.Context, host executor.Host, client *ssh.Client) executor.Result {
		return executor.RunShell(ctx, host, client, command)
	})
	return utils.PrintResultAndCheck(results)
}

// 批量处理
func runBatchMode(cmd *cobra.Command, command string) error {
	return NewBatchExecutor(cmd, func(ctx context.Context, host executor.Host, client *ssh.Client) executor.Result {
		return executor.RunShell(ctx, host, client, command)
	})
}
