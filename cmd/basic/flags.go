package basic

import "github.com/spf13/cobra"

var BasicFlags = &Flags{
	Parameter: make(map[string]string),
}

type Flags struct {
	Parameter     map[string]string
	HostInventory []string
	Concurrency   int
	NoHistory     bool   // 本次执行不记录执行历史
	Timeout       int    // 单台执行超时（秒，0=不限制）
	RetryFile     string // 失败主机列表输出文件路径（空=不输出）
	Limit         string // 目标主机限制: @file (如 --limit @/tmp/failed.txt)
	Output        string // 输出格式: text(默认) / json
	Quiet         bool   // 静默模式：只输出命令 stdout，无装饰文本
	DryRun        bool   // 干跑模式：只显示将要执行的命令和目标，不实际执行
}

// These parameters are registered via basic.PersistentFlags(),
// All basic subcommands are available.

func registerBasicFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.IntVar(&BasicFlags.Concurrency, "concurrency", 100, "Maximum Concurrency level")
	flags.IntVar(&BasicFlags.Timeout, "timeout", 0, "Execution timeout per host (seconds, 0 = unlimited)")
	flags.StringVar(&BasicFlags.RetryFile, "retry-file", "", "Output failed hosts to file")
	flags.StringVar(&BasicFlags.Limit, "limit", "", "Limit target hosts")
	flags.StringVar(&BasicFlags.Output, "output", "", "Output format: text|json|yaml")
	flags.BoolVar(&BasicFlags.Quiet, "quiet", false, "Only output command stdout")
	flags.BoolVar(&BasicFlags.DryRun, "dry-run", false, "Show targets and operations without execution")
	flags.BoolVar(&BasicFlags.NoHistory, "no-history", false, "Do not show history")
}
