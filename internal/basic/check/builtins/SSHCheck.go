package builtins

import (
	"fmt"
	"strings"
	
	"github.com/kraken-toolkit/internal/basic/check"
)

// SSHCkect 针对sshd 服务配置的检查，（认证策略优化）
type SSHCheck struct {
	name            string
	hostKeyAlgo     string // 期望的认证密钥类型 如 ed25519
	passwordAuth    string // 期望的PasswordAuthentication 值，如 no
	permitRootLogin string // 期望的permitRootLogin 值，如prohibit-password
}

func (s *SSHCheck) Name() string { return s.name }

// Command 收集 sshd 当前生效配置 (-T dump 实际生效值，非配置源文件
func (s *SSHCheck) Command() string {
	return "/usr/sbin/sshd -T 2>/dev/null || sshd -T 2>/dev/null"
}

// Validate 解析 sshd -T 输出为 key-value ，挨个param判定，空param 表示不检查该字段
func (s *SSHCheck) Validate(stdout string) (bool, string) {
	kv := make(map[string]string)
	for _, line := range strings.Split(stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			kv[fields[0]] = fields[1]
		}
	}
	var fails []string
	if s.hostKeyAlgo != "" {
		hk := kv["hostkey"]
		// hostkey 可能是 "ssh-ed25519" 或 "/path/key.pub", 按包含判断
		if !strings.Contains(hk, s.hostKeyAlgo) {
			fails = append(fails, fmt.Sprintf("hostkey lachs %s (got %q)", s.hostKeyAlgo, hk))
		}
	}
	if s.passwordAuth != "" {
		if pa := kv["passwordauthentication"]; pa != s.passwordAuth {
			fails = append(fails, fmt.Sprintf("password authentication %s want %s", pa, s.passwordAuth))
		}
	}
	if s.permitRootLogin != "" {
		if rl := kv["permitrootlogin"]; rl != s.permitRootLogin {
			fails = append(fails, fmt.Sprintf("permitrootlogin %s want %s", rl, s.permitRootLogin))
		}
	}
	if len(fails) == 0 {
		return true, "sshd config matches"
	}
	return false, strings.Join(fails, "; ")
}
func init() {
	registry.Register("ssh", func(params map[string]any) check.Check {
		return &SSHCheck{
			name:            "",
			hostKeyAlgo:     getString(params["hostkey_algo"], ""),
			passwordAuth:    getString(params["password_auth"], ""),
			permitRootLogin: getString(params["permit_root_login"], ""),
		}
		
	})
}
