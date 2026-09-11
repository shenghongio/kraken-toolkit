package config

import (
	"fmt"
)

// BasicConfig Basic 模块配置
type BasicConfig struct {
	// 主机清单。
	//
	// 示例：
	//
	// host_inventory:
	//   web:
	//     - 192.168.1.[101:110]
	//   db:
	//     - 192.168.2.[1:5]
	//
	HostInventory map[string][]string `yaml:"host_inventory"`

	// 默认并发数
	// 0表示使用程序默认
	Concurrency int `yaml:"concurrency"`

	//SSH 配置
	SSH BasicSSHConfig `yaml:"ssh"`

	// Fetch 操作的默认目标目录
	DefaultFetchPath string `yaml:"default_fetch_path"`

	// 操作历史记录配置
	History BasicConfigHistory `yaml:"history"`
}

// BasicSSHConfig Basic 模块 SSH 配置。
type BasicSSHConfig struct {
	// SSH 服务端口。
	SSHPort int `yaml:"ssh_port"`

	// SSH 用户
	SSHUser string `yaml:"ssh_user"`

	// SSH密码
	SSHPassword string `yaml:"ssh_password"`

	// SSH 连接超时时间，单位：秒。
	SSHTimeout int `yaml:"ssh_timeout"`
}

// BasicConfigHistory Basic 模块历史记录配置。
type BasicConfigHistory struct {
	// 是否开启历史记录。
	Enabled bool `yaml:"enabled"`
	// 历史记录文件路径。
	Log string `yaml:"log"`
}

// Validate 校验Basic 模块配置
func (c BasicConfig) Validate() error {
	if c.Concurrency < 0 {
		return fmt.Errorf("concurrency must be greater than or equal to  zero")
	}
	if c.SSH.SSHPort < 0 || c.SSH.SSHPort > 65535 {
		return fmt.Errorf("ssh port must be between 0 and 65535")
	}
	if c.SSH.SSHTimeout < 0 {
		return fmt.Errorf("ssh timeout must be greater than or equal to  zero")
	}
	return nil
}
