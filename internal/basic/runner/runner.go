package runner

import (
	"context"
	"fmt"
	
	"github.com/kraken-toolkit/internal/basic/executor"
	"github.com/kraken-toolkit/internal/config"
	"golang.org/x/crypto/ssh"
)

// Config 是批量执行的统一配置。
type Config struct {
	Hosts       []executor.Host
	SSHOptions  executor.SSHOptions
	Concurrency int
}

// BuildFromConfig 从 config.Config 构建 Runner 配置。
func BuildFromConfig(cfg *config.Config) (Config, error) {
	if cfg == nil {
		return Config{}, fmt.Errorf("config is nil")
	}
	if len(cfg.Basic.IPList) == 0 {
		return Config{}, fmt.Errorf("batch mode requires iplist in config")
	}

	hosts := make([]executor.Host, len(cfg.Basic.IPList))
	for i, ip := range cfg.Basic.IPList {
		hosts[i] = executor.Host{
			Address: ip,
			User:    cfg.Basic.User,
			Port:    22,
		}
	}

	sshOpt := executor.SSHOptions{
		User:       cfg.Basic.User,
		Port:       22,
		PrivateKey: cfg.Basic.PrivateKeyPath,
	}

	concurrency := cfg.Basic.Concurrency
	if concurrency <= 0 {
		concurrency = 5
	}

	return Config{
		Hosts:       hosts,
		SSHOptions:  sshOpt,
		Concurrency: concurrency,
	}, nil
}

// Run 执行批量操作。
func Run(ctx context.Context, cfg Config, operation func(context.Context, executor.Host, *ssh.Client) executor.Result) []executor.Result {
	exec := executor.NewExecutor(executor.Options{
		Concurrency: cfg.Concurrency,
		SSH:         cfg.SSHOptions,
	})
	return exec.Run(ctx, cfg.Hosts, operation)
}
