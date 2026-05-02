package cmd

import (
	"encoding/json"
	"fmt"
	
	"github.com/spf13/cobra"
	"github.com/sysint/internal/version"
)

// versionCmd 版本命令
func versionCmd() *cobra.Command {
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
		fmt.Println(version.Short())
		return nil
	}
	
	if jsonoutput {
		info := version.Get()
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
func printVersionInfo() {
	info := version.Get()
	
	// 标题
	fmt.Printf("╔════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("                    系统初始化检查工具 - 版本信息                      \n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════╝\n\n")
	
	// 版本信息
	fmt.Printf("版本信息:\n")
	fmt.Printf("  版本号:     %s\n", info.Version)
	fmt.Printf("  Git提交:    %s\n", info.GitCommit)
	fmt.Printf("  Git分支:    %s\n", info.GitBranch)
	fmt.Printf("  编译时间:   %s\n", info.BuildTime)
	fmt.Printf("  Go版本:     %s\n", info.GoVersion)
	fmt.Printf("  操作系统:   %s\n", info.OS)
	fmt.Printf("  架构:       %s\n\n", info.Arch)
	
	// 开发版本警告
	if version.IsDev() {
		fmt.Printf("注意: 当前为开发版本，建议使用正式发布的版本\n\n")
	}
}
