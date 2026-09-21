package executor

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"strings"
	"time"
)

func RunShell(ctx context.Context, host Host, client *ssh.Client, cmd string) Result {
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
		if result.Stderr != "" {
			result.Error = fmt.Errorf("%s: %s", strings.TrimRight(result.Stderr, "\n"), err)
		} else {
			result.Error = err
		}
		if exitErr, ok := err.(*ssh.ExitError); ok {
			result.ExitCode = exitErr.ExitStatus()
		}
		result.Stdout = stdout.String()
		result.Stderr = stderr.String()
		return result
	}

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.ExitCode = 0
	return result

}
