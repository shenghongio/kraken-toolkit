package basic

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/kraken-toolkit/internal/basic/executor"
	"github.com/kraken-toolkit/internal/config"
	"github.com/kraken-toolkit/pkg/cli"
	"github.com/kraken-toolkit/utils"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"log/slog"
	"os"
	"strings"
)

func NewScriptCmd() *cobra.Command {
	var (
		scriptArgs  string
		interpreter string
	)
	cmd := &cobra.Command{
		Use:   "script <local-path>",
		Short: "Execute local script file on remote hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cli.PrintSubCmdHelp(cmd)
				return nil
			}
			localPath := args[0]

			// 单机模式
			if cmd.Flags().Changed("ssh-host") {
				return runScriptSingle(cmd, localPath, scriptArgs, interpreter)
			}
			// 批量模式
			cfg, ok := config.FromContext(cmd.Context())
			if ok && cfg.Basic.HostInventory {
				return runScriptBatch(cmd, localPath, scriptArgs, interpreter)
			}
			// 用户提供了 SSH 凭据但漏了 --ssh-host，给出针对性提示
			if cmd.Flags().Changed("ssh-user") || cmd.Flags().Changed("ssh-password") {
				fmt.Println("缺少 --ssh-host 参数")
				fmt.Println("  单机模式需要指定目标主机: --ssh-host=<IP>")
				return nil
			}
			fmt.Println("请选择执行模式:")
			fmt.Println("  单机: kraken bc script <file> --ssh-host=<IP>")
			fmt.Println("  批量: kraken bc script <file> --config <file> (配置中设置 host_inventory: true)")
			return nil
		},
	}
	cmd.Flags().StringVar(&scriptArgs, "args", "", "arguments passed to the script")
	cmd.Flags().StringVar(&interpreter, "interpreter", "", "script interpreter [bash|sh|python|python3|perl] (default: auto-detect from shebang)")
	cmd.Long = "Script is base64-encoded and piped to interpreter via SSH.\n" +
		"    * Interpreter auto-detected from shebang, override with --interpreter.\n" +
		"    * Pass arguments to the script with --args."
	cmd.Example = "kraken bc script deploy.sh --ssh-host=10.32.9.138"
	cmd.SetHelpTemplate(cli.SubCmdHelpTemplate)
	return cmd
}

func runScriptSingle(cmd *cobra.Command, localPath string, scriptArgs string, interpreter string) error {
	// 读取本地脚本文件
	scriptContext, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("read script %q: %w", localPath, err)
	}
	// 构造远程命令
	remoteCmd := buildRemoteCommand(scriptContext, interpreter, scriptArgs)
	slog.Debug("remote command", "interpreter", interpreter, "cmd", remoteCmd, "script_size", len(scriptContext))
	exec, host := NewSingleExecutor()
	results := exec.Run(cmd.Context(), []executor.Host{host}, func(ctx context.Context, host executor.Host, client *ssh.Client) executor.Result {
		return executor.RunShell(ctx, host, client, remoteCmd)
	})
	return utils.PrintResultAndCheck(results)
}

func runScriptBatch(cmd *cobra.Command, localPath string, scriptArgs string, interpreter string) error {
	// 读取本地脚本文件
	scriptContext, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("read script %q: %w", localPath, err)
	}
	remoteCmd := buildRemoteCommand(scriptContext, interpreter, scriptArgs)
	slog.Debug("remote command", "interpreter", interpreter, "cmd", remoteCmd, "script_size", len(scriptContext))
	return NewBatchExecutor(cmd, func(ctx context.Context, host executor.Host, client *ssh.Client) executor.Result {
		return executor.RunShell(ctx, host, client, remoteCmd)
	})
}

// buildRemoteCommand 将本地脚本内容编码为 base64，构造远程管道执行命令
// 编码测试: echo '<base64>' | base64 -d | <interpreter> <args>
func buildRemoteCommand(content []byte, interpreter string, scriptArgs string) string {
	if interpreter == "" {
		interpreter = detectInterpreter(content)
	}
	encoded := base64.StdEncoding.EncodeToString(content)
	cmd := fmt.Sprintf("echo '%s' | base64 -d | %s", encoded, interpreter)
	if scriptArgs != "" {
		// 使用 - 明确告诉解释器从 stdin 读取脚本, -- 分割解释器参数和脚本参数
		cmd += fmt.Sprintf(" - %s", scriptArgs)
	}
	return cmd
}

// datectInterpreter从 shebang 中检测脚本解释器
// 解析规则:
//   - "#!/bin/bash"                        → "bash"
//   - "#!/usr/bin/env python3"             → "python3"
//   - 无 shebang                            → "bash"（默认）
func detectInterpreter(content []byte) string {
	firstLine := string(content)
	if idx := strings.IndexByte(firstLine, '\n'); idx > 0 {
		firstLine = firstLine[:idx]
	}
	if !strings.HasPrefix(firstLine, "#!") {
		return "bash"
	}
	// 取 #! 后面第一个可执行文件名
	shebang := strings.TrimSpace(firstLine[2:])
	if idx := strings.LastIndex(shebang, "/"); idx >= 0 {
		interpreter := strings.TrimSpace(shebang[idx+1:])
		// 处理 "#!/bin/bash/evn bash" --> 取最后一个 "bash"
		if strings.HasPrefix(interpreter, "env") {
			interpreter = strings.TrimSpace(interpreter[4:])
		}
		if interpreter != "" {
			return interpreter
		}
	}
	interpreter := strings.TrimSpace(shebang)
	if interpreter != "" {
		return interpreter
	}
	return "bash"
}
