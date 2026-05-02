package models

import (
	"fmt"
	"time"
)

//NodeConfig 节点配置
type NodeConfig struct {
	Name     string        `yaml:"name" json:"name"`
	IP       string        `yaml:"ip" json:"ip"`
	Port     int           `yaml:"port" json:"port"`
	User     string        `yaml:"user" json:"user"`
	Password string        `yaml:"password" json:"password"`
	SSHKey   string        `yaml:"ssh_key" json:"ssh_key"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
}

// Validate 验证节点配置
func (n *NodeConfig) Validate() error {
	if n.IP == "" {
		return fmt.Errorf("节点IP不能为空")
	}
	if n.User == "" {
		return fmt.Errorf("节点 %s 的用户名不能为空", n.Name)
	}
	if n.SSHKey == "" && n.Password == "" {
		return fmt.Errorf("节点 %s 需要提供SSH密钥或密码", n.Name)
	}
	if n.Port == 0 {
		n.Port = 22
	}
	if n.Timeout == 0 {
		n.Timeout = 30 * time.Second
	}
	return nil
}

//ClusterConfig 集群配置
type ClusterConfig struct {
	Name  string       `yaml:"name" json:"name"`
	Nodes []NodeConfig `yaml:"nodes" json:"nodes"`
}

// CheckResult 单个节点检查结果
type CheckResult struct {
	NodeName string        `yaml:"nodeName" json:"nodeName"`
	NodeIP   string        `yaml:"nodeIP" json:"nodeIP"`
	NodePort int           `yaml:"nodePort" json:"nodePort"`
	Password string        `yaml:"password" json:"password"`
	Message  string        `yaml:"message" json:"message"`
	Duration time.Duration `yaml:"duration" json:"duration"`
	Detail   string        `yaml:"detail" json:"detail"`
	Error    string        `yaml:"error" json:"error"`
}

// CheckDetail 检查详情
type CheckDetail struct {
	Name     string `json:"name"`
	Status   string `json:"status"` // pass/fail/warning
	Actual   string `json:"actual"`
	Expected string `json:"expected"`
	Message  string `json:"message"`
}

// BatchCheckResult 批量检查结果
type BatchCheckResult struct {
	TotalNodes    int                     `json:"total_nodes"`
	SuccessNodes  int                     `json:"success_nodes"`
	FailedNodes   int                     `json:"failed_nodes"`
	Results       map[string]*CheckResult `json:"results"`
	TotalDuration time.Duration           `json:"total_duration"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	Hostname      string  `json:"hostname"`
	OS            string  `json:"os"`
	KernelVersion string  `json:"kernel_version"`
	CPU           CPUInfo `json:"cpu"`
	Memory        MemInfo `json:"memory"`
}

type CPUInfo struct {
	Cores int    `json:"cores"`
	Model string `json:"model"`
}

type MemInfo struct {
	TotalGB  float64 `json:"total_gb"`
	UsedGB   float64 `json:"used_gb"`
	FreeGB   float64 `json:"free_gb"`
	UsagePct float64 `json:"usage_pct"`
}
