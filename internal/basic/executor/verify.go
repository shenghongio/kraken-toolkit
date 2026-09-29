package executor

import (
	"context"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// PingSSH 对全部 hosts 做连接+鉴权校验，任一失败返回聚合错误。
// 供需要 SSH 操作的功能在执行前调用，作为连接级预防。
//
// 参数说明：
//   - ctx: 上下文，用于取消和超时控制
//   - hosts: 目标主机列表
//   - opts: 并发数和 SSH 连接参数（config 模式带私钥认证）
func PingSSH(ctx context.Context, hosts []Host, opts Options) error {
	results := NewExecutor(opts).Run(ctx, hosts, func(ctx context.Context, h Host, c *ssh.Client) Result {
		return RunShell(ctx, h, c, "echo kraken-ssh-ok")
	})
	for _, r := range results {
		if r.Error != nil {
			return fmt.Errorf("host %s: ssh ping failed: %w", r.Host.Address, r.Error)
		}
	}
	return nil
}