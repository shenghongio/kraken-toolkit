package cmd

import "github.com/spf13/cobra"

var (
	sysCheckCommand string
	sysCheckAutoFix bool
	sysCheckDetail  bool
	sysCheckOutput  string
)

// sysCheckCmd 环境检查命令
var sysCheckCmd = &cobra.Command{
	Use:   "syscheck",
	Short: "检查系统环境",
	Long: `检查系统环境，包括：
  - 操作系统版本
  - CPU/内存/磁盘资源
  - 网络配置
  - 系统参数配置
  - 必要软件包

示例：
  sysint syscheck -c system              # 基础系统检查
  sysint syscheck -c system --auto-fix   # 自动修复可修复的问题
  sysint syscheck -c system --detail     # 显示详细信息
  sysint syscheck -c k8s                 # K8s环境检查（待实现）`,
	RunE: runSysCheck,
}
