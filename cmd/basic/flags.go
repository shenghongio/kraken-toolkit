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

// NewSingleExecutor 从全局BasicFlags 构建单机模式的 Executor 和host
func NewSingleExecutor() (*executor.Executor, executor.Host) {
	sshOpt := executor.SSHOptions{
		User:     BasicFlags.SSHUser,
		Port:     BasicFlags.SSHPort,
		Password: BasicFlags.SSHPassword,
	}
	host := executor.Host{
		Address: BasicFlags.SSHHost,
		User:    BasicFlags.SSHUser,
		Port:    BasicFlags.SSHPort,
		Passwd:  BasicFlags.SSHPassword,
	}
	exec := executor.NewExecutor(executor.Options{
		Concurrency: 1,
		SSH:         sshOpt,
	})
	return exec, host
}

// NewBatchExecutor RunBatchExecutor 从全局BasicFlags 构建批量模式的 Executor 和host
func NewBatchExecutor(cmd *cobra.Command, operation func(ctx context.Context, host executor.Host, client *ssh.Client) executor.Result) error {
	cfg, ok := config.FromContext(cmd.Context())
	if !ok {
		return fmt.Errorf("batch mode requites config file (--config)")
	}
	runnerCfg, err := runner.BuildFromConfig(cfg)
	if err != nil {
		return err
	}
	slog.Debug("batch mode", "hosts", len(runnerCfg.Hosts), "concurrency", runnerCfg.Concurrency)
	results := runner.Run(cmd.Context(), runnerCfg, operation)
	return utils.PrintResultAndCheck(results)
}
