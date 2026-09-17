package executor

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kraken-pedestal/internal/basic/inventory"
	"golang.org/x/crypto/ssh"
	"strings"
	"time"
)

func runShell(ctx context.Context, host inventory.Host, client *ssh.Client, cmd string) Result {
	result := Result{
		Host: host,
	}

	start := time.Now()
	defer func() {
		result.Duration = time.Since(start)
	}()

	if client == nil {
		result.Error = fmt.Errorf("ssh client is nil")
		return result
	}

	select {
	case <-ctx.Done():
		result.Error = ctx.Err()
		return result
	default:
	}

	session, err := client.NewSession()
	if err != nil {
		result.Error = fmt.Errorf("create ssh session failed: %w", err)
		return result
	}

	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		result.Stdout = stdout.String()
		result.Stderr = stderr.String()
		result.Error = err
		if exitErr, ok := err.(*ssh.ExitError); ok {
			result.ExitCode = exitErr.ExitStatus()
		}
		return result
	}

	result.Success = true
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.ExitCode = 0
	return result

}

// Shell 在一组主机上并发执行 shell 命令，返回每一个主机的执行结果
// 这个事故 runShell 的对外封装： 通过闭包吧 cmd 绑定到 operation 中
// 叫个 executor.Run 做并发调度，ssh 连接,结果汇总
func (e *Executor) Shell(ctx context.Context, hosts []inventory.Host, cmd string) []Result {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		result := make([]Result, len(hosts))
		for i, host := range hosts {
			result[i] = Result{
				Host:  host,
				Error: fmt.Errorf("shell command is empty"),
			}
		}
		return result
	}

	return e.Run(ctx, hosts, func(ctx context.Context, host inventory.Host, client *ssh.Client) Result {
		return runShell(ctx, host, client, cmd)
	})
}
