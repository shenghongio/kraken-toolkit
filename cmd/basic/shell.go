package basic

import (
	"fmt"
	"github.com/kraken-pedestal/internal/basic/executor"
	"github.com/kraken-pedestal/internal/basic/inventory"
	execRuntime "github.com/kraken-pedestal/internal/basic/runtime"
	"github.com/kraken-pedestal/internal/config"
	"github.com/kraken-pedestal/utils"
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
				v := cmd.Context().Value(config.ContextKeyConfig)
				cfg, ok := v.(*config.Config)
				if ok && cfg != nil && cfg.Basic.HostInventory {
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
	exec := executor.NewExecutor(executor.Options{
		Concurrency: 1,
		SSH:         sshOpt,
	})

	rt := execRuntime.New(inv, exec)
	results := rt.Shell(cmd.Context(), command)
	printResults(results)
	return nil
}

// 批量处理
func runBatchMode(cmd *cobra.Command, command string) error {
	v := cmd.Context().Value(config.ContextKeyConfig)
	cfg, ok := v.(*config.Config)
	if !ok || cfg == nil {
		return fmt.Errorf("batch mode requires config file (--config)")
	}
	if !cfg.Basic.HostInventory {
		return fmt.Errorf("batch mode requires host_inventory: true in config")
	}
	if len(cfg.Basic.IPList) == 0 {
		return fmt.Errorf("batch mode requires iplist in config")
	}

	sshOpt := executor.SSHOptions{
		User:       cfg.Basic.User,
		Port:       22,
		PrivateKey: cfg.Basic.PrivateKeyPath,
	}
	slog.Debug("batch sshOpt", "private_key", sshOpt.PrivateKey)
	hosts := make([]inventory.Host, len(cfg.Basic.IPList))
	for i, ip := range cfg.Basic.IPList {
		hosts[i] = inventory.Host{
			Address: ip,
			User:    cfg.Basic.User,
			Port:    22,
		}
	}
	inv := inventory.NewFromHosts(hosts)

	concurrency := cfg.Basic.Concurrency
	if concurrency <= 0 {
		concurrency = 5
	}
	exec := executor.NewExecutor(executor.Options{
		Concurrency: concurrency,
		SSH:         sshOpt,
	})
	slog.Debug("batch mode", "hosts", len(hosts), "concurrency", concurrency)
	rt := execRuntime.New(inv, exec)
	results := rt.Shell(cmd.Context(), command)
	printResults(results)
	return nil
}

// 打印结果
func printResults(results []executor.Result) {
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("[FAIL] %s: %v\n", result.Host.Address, result.Error)
			continue
		}
		fmt.Printf("[OK] %s\n%s\n", result.Host.Address, result.Stdout)
	}
}
