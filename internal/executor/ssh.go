package executor

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"
	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	SSHKey   string
	Timeout  time.Duration
}

// SSHExecutor ssh执行器
type SSHExecutor struct {
	config *SSHConfig
	client *ssh.Client
}

// NewSSHEexecutor 创建SSH执行器
func NewSSHExecutor(config *SSHConfig) (*SSHExecutor, error) {
	// 验证必须参数
	if config.Host == "" {
		return nil, fmt.Errorf("主机地址不能为空")
	}
	if config.User == "" {
		return nil, fmt.Errorf("用户名称不能为空,SSH私钥不包含用户名信息")
	}
	if config.Port == "" {
		config.Port = "22"
	}
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	
	// 配置验证方法
	var authMethods []ssh.AuthMethod
	
	// 私钥认证
	if config.SSHKey != "" {
		key, err := ioutil.ReadFile(config.SSHKey)
		if err != nil {
			return nil, fmt.Errorf("读取SSH密钥失败: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("解析SSH密钥失败: %v", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	
	// 密码认证
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("请提供SSH私钥或密码进行认证")
	}
	
	// SSH客户端配置
	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            authMethods,
		Timeout:         config.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	
	// 连接
	adder := fmt.Sprintf("%s:%s", config.Host, config.Port)
	client, err := ssh.Dial("tcp", adder, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH连接失败: [ %s@%s] error: %s", config.User, adder, err)
	}
	return &SSHExecutor{
		config: config,
		client: client,
	}, nil
}

// Execute 执行命令
func (e *SSHExecutor) Execute(ctx context.Context, command string) (string, error) {
	session, err := e.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建session失败: %v", err)
	}
	defer session.Close()
	resultChan := make(chan struct {
		output string
		err    error
	}, 1)
	go func() {
		output, err := session.CombinedOutput(command)
		resultChan <- struct {
			output string
			err    error
		}{string(output), err}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-resultChan:
		return strings.TrimSpace(res.output), res.err
	}
}

// TestConnection 测试连接
func (e *SSHExecutor) TestConnection(ctx context.Context) error {
	_, err := e.Execute(ctx, "echo 'ok'")
	return err
}
