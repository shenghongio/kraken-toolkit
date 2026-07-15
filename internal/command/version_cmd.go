package command

import (
	"encoding/json"
	"fmt"
	"github.com/kraken-pedestal/internal/executor"
	"strings"
	
	"github.com/spf13/cobra"
)

// versionCmd 版本命令
func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Long:  "显示系统初始化检查工具的版本信息和编译详情",
		RunE:  runVersion,
	}
	
	// 添加命令行标志
	cmd.Flags().Bool("short", false, "只显示简短的版本号")
	cmd.Flags().Bool("json", false, "以JSON格式输出版本信息")
	
	return cmd
}

// runVersion 执行版本命令
func runVersion(cmd *cobra.Command, args []string) error {
	
	short, _ := cmd.Flags().GetBool("short")
	jsonoutput, _ := cmd.Flags().GetBool("json")
	
	if short {
		fmt.Println(executor.Short())
		return nil
	}
	
	if jsonoutput {
		info := executor.Get()
		jsonData, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("生成JSON失败: %v", err)
		}
		fmt.Println(string(jsonData))
		return nil
	}
	
	// 默认输出格式
	printVersionInfo()
	
	return nil
}

// printVersionInfo 打印版本信息
//func printVersionInfo() {
//	info := executor.Get()
//
//	// 标题
//	fmt.Printf("╔════════════════════════════════════════════════════════════════╗\n")
//	fmt.Printf("                    系统初始化检查工具 - 版本信息                      \n")
//	fmt.Printf("╚════════════════════════════════════════════════════════════════╝\n\n")
//
//	// 版本信息
//	fmt.Printf("版本信息:\n")
//	fmt.Printf("  版本号:     %s\n", info.Version)
//	fmt.Printf("  Git提交:    %s\n", info.GitCommit)
//	fmt.Printf("  Git分支:    %s\n", info.GitBranch)
//	fmt.Printf("  编译时间:   %s\n", info.BuildTime)
//	fmt.Printf("  Go版本:     %s\n", info.GoVersion)
//	fmt.Printf("  操作系统:   %s\n", info.OS)
//	fmt.Printf("  架构:       %s\n\n", info.Arch)
//
//	// 开发版本警告
//	if executor.IsDev() {
//		fmt.Printf("注意: 当前为开发版本，建议使用正式发布的版本\n\n")
//	}
//}

func printVersionInfo() {
	info := executor.Get()
	
	// 颜色定义（支持大多数终端）
	const (
		reset   = "\033[0m"
		bold    = "\033[1m"
		green   = "\033[32m"
		yellow  = "\033[33m"
		cyan    = "\033[36m"
		magenta = "\033[35m"
		white   = "\033[37m"
		gray    = "\033[90m"
	)
	
	// ---------- 标题区域 ----------
	fmt.Printf("\n%s%s%s\n", bold, green, strings.Repeat("═", 56))
	fmt.Printf("%s  %s系统初始化检查工具%s  %s\n", bold, cyan, reset, bold)
	fmt.Printf("%s  %s版本信息%s  %s\n", bold, white, reset, bold)
	fmt.Printf("%s%s\n", green, strings.Repeat("═", 56))
	fmt.Println()
	
	// ---------- 应用程序信息 ----------
	fmt.Printf("%s●%s %s应用程序信息%s\n", cyan, reset, bold, reset)
	fmt.Printf("  %-12s %s\n", "版本号:", colorValue(info.Version))
	fmt.Printf("  %-12s %s\n", "Git提交:", colorValue(info.GitCommit))
	fmt.Printf("  %-12s %s\n", "Git分支:", colorValue(info.GitBranch))
	fmt.Printf("  %-12s %s\n", "编译时间:", colorValue(info.BuildTime))
	fmt.Println()
	
	// ---------- 构建环境 ----------
	fmt.Printf("%s●%s %s构建环境%s\n", cyan, reset, bold, reset)
	fmt.Printf("  %-12s %s\n", "Go版本:", colorValue(info.GoVersion))
	fmt.Printf("  %-12s %s\n", "操作系统:", colorValue(info.OS))
	fmt.Printf("  %-12s %s\n", "架构:", colorValue(info.Arch))
	fmt.Println()
	
	// ---------- 开发版本提示 ----------
	if executor.IsDev() {
		fmt.Printf("%s⚠  %s开发版本%s  — 建议使用正式发布的版本\n", yellow, bold, reset)
		fmt.Println()
	}
	
	// ---------- 底部装饰 ----------
	fmt.Printf("%s%s\n", gray, strings.Repeat("─", 56))
	fmt.Printf("%s  使用 %s--help%s 查看完整命令帮助\n", gray, cyan, gray)
	fmt.Printf("%s%s\n", gray, strings.Repeat("─", 56))
}

// colorValue 根据内容类型返回带颜色的值
func colorValue(val string) string {
	const (
		reset = "\033[0m"
		green = "\033[32m"
		white = "\033[37m"
	)
	if val == "" || val == "unknown" || val == "dev" {
		return fmt.Sprintf("%s%s%s", "\033[33m", val, reset) // 黄色高亮
	}
	return fmt.Sprintf("%s%s%s", white, val, reset)
}
