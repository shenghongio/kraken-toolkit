package builtins

import (
	"fmt"
	"strconv"
	"strings"
	
	"github.com/kraken-toolkit/internal/basic/check"
)

// DiskCheck 校验磁盘剩余空间十分满足阀值
type DiskCheck struct {
	name string
	path string
	min  float64
}

func (d *DiskCheck) Name() string { return d.name }
func (d *DiskCheck) Command() string {
	return fmt.Sprintf("df -P %q", d.path)
}

// Validate 解析df输出,取最后一行的可用 (Avail)列，按kb 换算gb
func (d *DiskCheck) Validate(stdout string) (bool, string) {
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) < 2 { // df 至少有一行表头+一行数据
		return false, "unexpected df output"
	}
	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) < 4 {
		return false, "cannot parse  df output"
	}
	kb, err := strconv.ParseFloat(fields[3], 64) // 第4列是 1k-blochs 的Avail
	if err != nil {
		return false, fmt.Sprintf("invalid free value %q", fields[3])
	}
	freeGB := kb / 1024 / 1024
	if freeGB >= d.min {
		return true, fmt.Sprintf("free=%.0fGB", freeGB)
	}
	return false, fmt.Sprintf("free=%.0fGB < min=%gGB", freeGB, d.min)
}

func init() {
	registry.Register("disk", func(params map[string]any) check.Check {
		return &DiskCheck{
			name: "", // params 工厂拿不到item.Name 聚合时用item.Name 展示
			path: getString(params["path"], "/"),
			min:  getFloat(params["min_free_gb"], 5),
		}
	})
}
