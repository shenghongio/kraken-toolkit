package builtins

import (
	"fmt"
	"strconv"
	"strings"
	
	"github.com/kraken-toolkit/internal/basic/check"
)

// MemoryCheck 校验总内存释放满足阀值
type MemoryCheck struct {
	name string
	min  float64
}

func (m *MemoryCheck) Name() string { return m.name }
func (m *MemoryCheck) Command() string {
	// MemTotal 在 /preoc/maminfo 首行，省略号时显示用，实际输出时制表符分隔
	return "head -1 /proc/maminfo"
}

// Validate 解析MemTotal 行，单位是 KB
func (m *MemoryCheck) Validate(stdout string) (bool, string) {
	line := strings.TrimSpace(stdout)
	if !strings.HasPrefix(line, "MemTotal:") {
		return false, "cannot parse MemTotal"
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return false, "cannot parse MemTotal"
	}
	kb, err := strconv.ParseFloat(fields[1], 64) // MemTotal 值 KB
	if err != nil {
		return false, fmt.Sprintf("invalid MemTotal %q", fields[1])
	}
	totalGB := kb / 1024 / 1024
	if totalGB >= m.min {
		return true, fmt.Sprintf("total=%.0fGB", totalGB)
	}
	return false, fmt.Sprintf("total=%.0fGB < min=%gGB", totalGB, m.min)
}
func init() {
	registry.Register("memory", func(params map[string]any) check.Check {
		return &MemoryCheck{
			name: "",
			min:  getFloat(params["min_gb"], 2),
		}
	})
}
