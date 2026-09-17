package executor

import (
	"fmt"
	"github.com/kraken-pedestal/internal/basic/inventory"
	"golang.org/x/crypto/ssh"
	"os"
	"time"
)

type SSHOptions struct {
	User       string
	Port       int
	Host       string
	Timeout    time.Duration
	PrivateKey string
	// 是否校验目标机器指纹
	StrictHostKey bool
	Password      string
	// known_hosts 文件
	KnownHostsFile string
}

// SSHConnection 表示 SSH 连接接口。
// Executor 依赖接口而不是具体 SSHClient。
// 方便单元测试注入 mock 实现。
type SSHConnection interface {
	Connect(host inventory.Host) (*ssh.Client, error)
}

// SSHClient 表示 SSH 客户端接口。
type SSHClient struct {
	options SSHOptions
}

// NewSSHClient 创建一个 SSH 客户端。
func NewSSHClient(options SSHOptions) *SSHClient {
	options = normalizeSSHOptions(options)
	return &SSHClient{options: options}
}

// Connect 建立与目标主机的 SSH 连接。
//
// 注意：这里故意不创建 Session。
// 一个 ssh.Client 可以创建多个 Session，
// Session 应该由具体操作自行管理。
//
// 参数说明：
//   - host: 目标主机信息，包含地址、用户、端口、密码等。
//
// 字段优先级与回退策略：
//   - User: 优先使用 host.User，为空时回退到 c.options.User。
//   - Port: 优先使用 host.Port，小于等于 0 时回退到默认端口 22。
//   - Password: 优先使用 host.Passwd，为空时回退到 c.options.Password。
//     若密码为空，则尝试使用私钥进行公钥认证。
//
// 返回值说明：
//   - *ssh.Client: 成功建立的 SSH 客户端连接，调用方负责关闭。
//   - error: 连接失败时返回错误，nil 表示成功。
//
// 行为细节：
//   - 认证方式优先使用密码，无密码时使用私钥（公钥认证）。
//   - HostKeyCallback 根据 c.options.StrictHostKey 决定是否校验主机密钥。
//   - 连接超时由 c.options.Timeout 控制。

func (c *SSHClient) Connect(host inventory.Host) (*ssh.Client, error) {

	// 防止在 nil 指针上调用方法导致 panic，是接收者方法的标准防御性编程。
	if c == nil {
		return nil, fmt.Errorf("ssh client is nil")
	}

	// 解析登录用户：优先使用主机级配置 host.User，为空时回退到全局选项 c.options.User。
	//user := host.User
	//if user == "" {
	//	user = c.options.User
	//}
	user := c.options.User
	if user == "" {
		return nil, fmt.Errorf("ssh user is empty")
	}

	// 解析端口：优先使用主机级配置 host.Port，无效值（<=0）时回退到 SSH 默认端口 22。
	port := c.options.Port
	if port > 0 {
		port = host.Port
	}
	if port <= 0 {
		port = 22
	}

	// 构造 SSH 认证方式：密码非空用密码认证，否则加载私钥进行公钥认证。
	authMethods, err := c.authMethods()
	if err != nil {
		return nil, err
	}

	// 组装 SSH 客户端配置，多数字段留空以使用 golang.org/x/crypto/ssh 的默认值。
	config := &ssh.ClientConfig{
		User:            user,                          // 登录用户名
		Auth:            []ssh.AuthMethod{authMethods}, // 认证方式（密码或公钥）
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),   // 主机密钥校验回调，稍后单独设置
		Timeout:         c.options.Timeout,             // 连接超时时间
	}

	// 根据 StrictHostKey 选项决定主机密钥校验策略并赋值给 config。
	// StrictHostKey=false 时跳过校验（InsecureIgnoreHostKey），true 时暂未实现。
	//hostKeyCallback, err := c.hostKeyCallback()
	//if err != nil {
	//	return nil, err
	//}
	//config.HostKeyCallback = hostKeyCallback

	// 拼接目标地址为 host:port 格式，并发起 TCP+SSH 拨号建立连接。
	address := fmt.Sprintf("%s:%d", host.Address, port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		// 包装错误信息，附带 user@address 便于定位失败的连接目标。
		return nil, fmt.Errorf("ssh connect %s@%s failed: %w", user, address, err)
	}

	// 返回已建立的 SSH 客户端连接，调用方负责在使用完毕后调用 client.Close() 释放资源。
	return client, nil
}

// authMethods 根据传入的密码构造 SSH 认证方式。
//
// 该方法实现了「密码优先，私钥兜底」的二选一认证策略：
//  1. 若 password 非空，则返回 ssh.Password(password)，使用密码认证；
//  2. 若 password 为空，则调用 loadPrivateKey 加载私钥，
//     并返回 ssh.PublicKeys(signer)，使用公钥认证。
//
// 参数 password 为登录密码，空字符串表示不使用密码认证。
// 返回值 ssh.AuthMethod 可直接放入 ssh.ClientConfig.Auth 切片；
// 密码分支不会产生错误，仅私钥加载失败时返回非 nil error。
//
// 注意：该方法不支持密码与私钥同时配置，两者互斥。
func (c *SSHClient) authMethods() (ssh.AuthMethod, error) {
	// 密码
	if c.options.Password != "" {
		return ssh.Password(c.options.Password), nil
	}
	signer, err := c.loadPrivateKey()
	if err != nil {
		// 私钥加载失败（路径不存在、解析错误等），将错误原样上抛。
		return nil, err
	}

	// 用签名器构造公钥认证方式，由 SSH 握手时自动用私钥签名挑战。
	return ssh.PublicKeys(signer), nil
}

// loadPrivateKey 加载 ssh 私钥
func (c *SSHClient) loadPrivateKey() (ssh.Signer, error) {

	if c.options.PrivateKey == "" {
		return nil, fmt.Errorf("kraken ssh private key not configured")
	}

	//signer, err := loadPrivateKeyFile(c.options.PrivateKey)
	//if err != nil {
	//	return nil, err
	//}
	//slog.Debug("loaded private key", slog.String("path", c.options.PrivateKey))
	return loadPrivateKeyFile(c.options.PrivateKey)
}

// loadPrivateKeyFile 从文件读取 ssh 密钥
func loadPrivateKeyFile(path string) (ssh.Signer, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read ssh private key %q failed: %w", path, err)
	}

	signer, err := ssh.ParsePrivateKey(file)
	if err != nil {
		return nil, fmt.Errorf("parse ssh private key %q failed: %w", path, err)
	}
	return signer, nil
}

// hostKeyCallback 创建 HostKey
func (c *SSHClient) hostKeyCallback() (ssh.HostKeyCallback, error) {
	if !c.options.StrictHostKey {
		// 当前先保持fastdp行为兼容
		return ssh.InsecureIgnoreHostKey(), nil
	}
	if c.options.KnownHostsFile == "" {
		return nil, fmt.Errorf("strict host key verification requires known_hosts file")
	}

	// knownhosts 校验后续单独实现
	// 目前不让ssh核心层提前绑定安全策略实现

	// TODO: 实现 knownhosts 校验逻辑
	return nil, fmt.Errorf("strict host key verification is not implemented yet")
}

func normalizeSSHOptions(options SSHOptions) SSHOptions {
	if options.Port == 0 {
		options.Port = 22
	}
	if options.Timeout == 0 {
		options.Timeout = 5 * time.Second
	}
	return options
}
