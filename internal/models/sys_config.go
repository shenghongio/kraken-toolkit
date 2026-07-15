package models

// ChecklistItem  检查项
type ChecklistItem struct {
	Type        string      `yaml:"type"`            // sysctl, ulimit, kernel, cgroup, containerRuntime, port, system, storage
	Name        string      `yaml:"name"`            // 检查项名称
	Enabled     bool        `yaml:"enabled"`         //是否启用
	Expected    interface{} `yaml:"value,omitempty"` //期望值
	Description string      `yaml:"description"`     // 描述
}

// SysCheckConfig 系统检查配置
type SysCheckConfig struct {
	Checklist []ChecklistItem `yaml:"checklist"`
}

// CheckResult 检查结果
type CheckResult struct {
	Total    int               `json:"total"`    // 总检查项
	Passed   int               `json:"passed"`   // 通过数
	Failed   int               `json:"failed"`   // 失败数
	Warnings int               `json:"warnings"` // 警告数
	Items    []CheckItemResult `json:"items"`    // 检查项详情
	Summary  string            `json:"summary"`  // 总结信息
}

// CheckItemResult 单项检查结果
type CheckItemResult struct {
	Type        string      `json:"type"`
	Name        string      `json:"name"`
	Status      string      `json:"status"` // PASS, FAIL, WARN
	Expected    interface{} `json:"expected"`
	Actual      interface{} `json:"actual"`
	Message     string      `json:"message"`
	Description string      `json:"description"`
	Fixable     bool        `json:"fixable"`               // 是否可自动修复
	FixCommand  string      `json:"fix_command,omitempty"` // 修复命令
}
