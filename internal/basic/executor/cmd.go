package executor

import (
	"context"

	"golang.org/x/crypto/ssh"
)

// RunOnHosts 在多台主机上并发执行同一 shell 命令
//
// 参数:
//   - ctx:   上下文，用于取消和超时控制
//   - hosts: 目标主机列表
//   - cmd:   要执行的 shell 命令字符串
//   - opts:  并发数和 SSH 连接参数
//
// 返回每台主机的执行结果，顺序与 hosts 一致
func RunOnHosts(ctx context.Context, hosts []Host, cmd string, opts Options) []Result {
	exec := NewExecutor(opts)
	operation := func(ctx context.Context, host Host, client *ssh.Client) Result {
		return RunShell(ctx, host, client, cmd)
	}
	return exec.Run(ctx, hosts, operation)
}
