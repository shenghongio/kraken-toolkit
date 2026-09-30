package basic

import (
	"fmt"
	
	"github.com/kraken-toolkit/internal/basic/check"
	"github.com/kraken-toolkit/internal/basic/check/builtins"
	"github.com/kraken-toolkit/internal/basic/executor"
	"github.com/kraken-toolkit/internal/basic/runner"
	"github.com/kraken-toolkit/internal/config"
	"github.com/kraken-toolkit/pkg/cli"
	"github.com/kraken-toolkit/utils"
	"github.com/spf13/cobra"
)

func NewCheckCmd() *cobra.Command {
	var checkFile string
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Run checks from manifest against target hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cli.PrintSubCmdHelp(cmd)
				return nil
			}
			
			// 加载本地配置清单，+绑定注册表，构建checks
			manifest, err := check.LoadManifest(checkFile)
			if err != nil {
				return err
			}
			reg := builtins.Registry()
			var checks []check.Check
			for _, item := range manifest.Checks {
				build, err := reg.Build(item)
				if err != nil {
					return err
				}
				checks = append(checks, build)
			}
			if len(checks) == 0 {
				return fmt.Errorf("no checks defined in manifest")
			}
			//模式分发： 单机 --ssh-host / 批量 config(host_inventory)
			if cmd.Flags().Changed("ssh-host") {
				return runCheckSingle(cmd, checks)
			}
			fromContext, ok := config.FromContext(cmd.Context())
			if ok && fromContext.Basic.HostInventory {
				return runCheckBatch(cmd, checks)
			}
			if cmd.Flags().Changed("ssh-user") || cmd.Flags().Changed("ssh-password") {
				fmt.Println("Warning: ssh-user and ssh-password are mutually exclusive")
				fmt.Println("   In standalone mode, the target host needs to be specified: --ssh-host=<IP>")
				return nil
			}
			fmt.Println("Please select the execution mode:")
			fmt.Println("  Standalone: kraken bc check --ssh-host=<IP>")
			fmt.Println("  Batch: kraken bc check --config <file> (set host_inventory: true in the configuration)")
			return nil
		},
	}
	cmd.Flags().StringVar(&checkFile, "check-file", "", "check manifest file (default .check.yaml)")
	cmd.SetHelpTemplate(cli.SubCmdHelpTemplate)
	return cmd
}

// 单机模式，直接复用全局 basicFlags 构造hosts+opts
func runCheckSingle(cmd *cobra.Command, checks []check.Check) error {
	hosts := []executor.Host{
		{
			Address: BasicFlags.SSHHost,
			User:    BasicFlags.SSHUser,
			Port:    BasicFlags.SSHPort,
			Passwd:  BasicFlags.SSHPassword,
		},
	}
	opts := executor.Options{
		Concurrency: 1,
		SSH: executor.SSHOptions{
			User:     BasicFlags.SSHUser,
			Port:     BasicFlags.SSHPort,
			Password: BasicFlags.SSHPassword,
		},
	}
	results := executor.RunChecks(cmd.Context(), hosts, opts, checks)
	return utils.PrintCheckResults(results)
}

// 批量模式 runner,BuildFromconfig 提供 host+sshoptions +concurrency
func runCheckBatch(cmd *cobra.Command, checks []check.Check) error {
	cfg, ok := config.FromContext(cmd.Context())
	if !ok {
		return fmt.Errorf("batch mode requires config file (--config)")
	}
	fromConfig, err := runner.BuildFromConfig(cfg)
	if err != nil {
		return err
	}
	opts := executor.Options{
		Concurrency: fromConfig.Concurrency,
		SSH:         fromConfig.SSHOptions,
	}
	results := executor.RunChecks(cmd.Context(), fromConfig.Hosts, opts, checks)
	return utils.PrintCheckResults(results)
}
