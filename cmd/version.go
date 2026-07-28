package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/kraken-pedestal/internal/deploy/executor"
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

func printVersionInfo() {
	info := executor.Get()

	// 简约版本信息（无框线，简洁字段）
	fmt.Printf("Version:    %s\n", info.Version)
	fmt.Printf("GitCommit:  %s\n", info.GitCommit)
	fmt.Printf("GitBranch:  %s\n", info.GitBranch)
	fmt.Printf("BuildTime:  %s\n", info.BuildTime)
	fmt.Printf("GoVersion:  %s\n", info.GoVersion)
	fmt.Printf("OS/Arch:    %s/%s\n", info.OS, info.Arch)

	// 开发版本警告精简为一行（可选）
	if executor.IsDev() {
		fmt.Printf("(Development build)\n")
	}
}
