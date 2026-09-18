package basic

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/kraken-pedestal/internal/basic/executor"
	"github.com/kraken-pedestal/internal/basic/runner"
	"github.com/kraken-pedestal/internal/config"
	"github.com/kraken-pedestal/utils"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"log/slog"
	"os"
	"strings"
	"text/template"
)

//go:embed scripts/create_kraken_user.sh
var createUserScript string

func NewAddUserCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "adduser",
		Short: "Create management user and deploy SSH key on target hosts",

		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, ok := config.FromContext(cmd.Context())
			if !ok {
				utils.PrintUsage(cmd, "Example: kraken basic adduser --config kraken.yaml --confirm")
				return nil
			}
			if len(cfg.Basic.IPList) == 0 {
				return fmt.Errorf("adduser requires iplist in config")
			}
			if cfg.Basic.Bootstrap.User == "" {
				return fmt.Errorf("adduser requires bootstrap.user in config")
			}
			if cfg.Basic.PrivateKeyPath == "" {
				return fmt.Errorf("adduser requires private_key_path in config")
			}
			if cfg.Basic.PublicKeyPath == "" {
				return fmt.Errorf("adduser requires public_key_path in config")
			}

			// 读取公钥内容
			pubKeyBytes, err := os.ReadFile(cfg.Basic.PublicKeyPath)
			if err != nil {
				return fmt.Errorf("read public key %q: %w", cfg.Basic.PublicKeyPath, err)
			}
			pubKeyContext := strings.TrimSpace(string(pubKeyBytes))
			newUser := cfg.Basic.User

			// 确认
			if !confirm {
				fmt.Printf("WARNING: 将在 %d 台主机上创建用户 '%s' 并部署公钥\n", len(cfg.Basic.IPList), newUser)
				fmt.Printf("  登录用户: %s\n", cfg.Basic.Bootstrap.User)
				fmt.Printf("  公钥文件: %s\n", cfg.Basic.PublicKeyPath)
				fmt.Println("\n使用 --confirm 确认执行")
				return nil
			}

			runnerCfg, err := runner.BuildFromConfig(cfg)
			if err != nil {
				return err
			}

			// adduser 需要用bootstrap用户登录，覆盖默认的ssh用户
			runnerCfg.SSHOptions.User = cfg.Basic.Bootstrap.User
			results := runner.Run(cmd.Context(), runnerCfg, func(ctx context.Context, host executor.Host, client *ssh.Client) executor.Result {
				err := setUpKrakenUser(client, newUser, pubKeyContext)
				return executor.Result{
					Host:    host,
					Error:   err,
					Message: fmt.Sprintf("create user %s", newUser),
				}
			})
			return utils.PrintResultAndCheck(results)
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm adduser")
	return cmd
}

func setUpKrakenUser(client *ssh.Client, username, pubKey string) error {
	// 渲染脚本模版
	tmpl, err := template.New("create_user_script").Parse(createUserScript)
	if err != nil {
		return fmt.Errorf("parse scripts template: %w", err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"Username":  username,
		"PublicKey": pubKey,
	})
	if err != nil {
		return fmt.Errorf("reader scripts template: %w", err)
	}
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("create ssh session failed: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(buf.String())
	if err != nil {
		return fmt.Errorf("create user failed: %w\n%s", err, strings.TrimSpace(string(output)))
	}
	slog.Debug("KRAKEN user", "user", username, "output", strings.TrimSpace(string(output)))
	return nil
}
