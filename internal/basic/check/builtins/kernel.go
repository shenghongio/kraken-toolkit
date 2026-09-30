package builtins

import (
	"strings"
	
	"github.com/kraken-toolkit/internal/basic/check"
)

// KernelCheck 校验内核相关参数，目前支持 swap状态检查
type KernelCheck struct {
	name string
}

func (k *KernelCheck) Name() string { return k.name }

// Command 同时采集内核版本与 swap 状态，用分隔标记分段，供 Validate 拼接展示。
func (k *KernelCheck) Command() string {
	return "echo KERNEL=$(uname -r); echo SWAP_LINES=$(cat /proc/swaps | tail -n +2 | wc -l)"
}

// Validate 信息采集型：不判定 pass/fail，仅组装信息供人工判断，恒返回 ok=true。
func (k *KernelCheck) Validate(stdout string) (bool, string) {
	info := strings.TrimSpace(stdout)
	if info == "" {
		return false, "no kernel info collected"
	}
	return true, info
}

func init() {
	registry.Register("kernel", func(params map[string]any) check.Check {
		return &KernelCheck{name: ""}
	})
}
