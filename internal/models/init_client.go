package models

type AuthMethod string

const (
	AuthPassword AuthMethod = "password"
	AuthKey      AuthMethod = "key"
)

type NodeItem struct {
	IP       string `mapstructure:"ip"`
	PASSWORD string `mapstructure:"password"`
}

type GlobalConfig struct {
	DeployUser string     `mapstructure:"deploy_user"`
	SSHUser    string     `mapstructure:"ssh_user"`
	SSHPort    string     `mapstructure:"ssh_port"`
	SSHKey     string     `mapstructure:"ssh_key"`
	Nodes      []NodeItem `mapstructure:"nodes"`
}

type Result struct {
	Node   string
	Output string
	Error  error
}

type Node struct {
	IP       string
	Port     int
	User     string
	Password string // 仅当使用密码认证时
	KeyPath  string // 仅当使用密钥认证时
}

type Client struct {
	User     string
	Password string
	KeyPath  string
	Port     int
	Timeout  int // 秒
}
