package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/kraken-pedestal/pkg/cli"
	"github.com/spf13/cobra"
)

// VersionCmd  版本命令
func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print kraken version information",
		GroupID: cli.GroupOther,
		RunE:    runVersion,
	}
	// 添加命令行标志
	cmd.Flags().Bool("json", false, "Output version information in JSON format")
	return cmd
}

// runVersion 执行版本命令
func runVersion(cmd *cobra.Command, args []string) error {
	jsonoutput, err := cmd.Flags().GetBool("json")
	
	if err != nil {
		return fmt.Errorf("failed to retrieve JSON flag: %w", err)
	}
	
	info := cli.Get()
	
	if jsonoutput {
		jsonData, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to generate JSON: %v", err)
		}
		fmt.Println(string(jsonData))
		return nil
	}
	
	// 默认输出格式
	printVersionInfo(info)
	
	return nil
}

func printVersionInfo(info cli.VersionInfo) {
	buildTimeDisplay := info.BuildTime
	if t, err := info.GetBuildTime(); err == nil {
		buildTimeDisplay = t.Local().Format("2006-01-02 15:04:05")
	}
	fmt.Printf("Version:    %s\n", info.Version)
	fmt.Printf("GitCommit:  %s\n", info.GitCommit)
	fmt.Printf("GitBranch:  %s\n", info.GitBranch)
	fmt.Printf("BuildTime:  %s\n", buildTimeDisplay)
	fmt.Printf("GoVersion:  %s\n", info.GoVersion)
	fmt.Printf("OS/Arch:    %s/%s\n", info.OS, info.Arch)
	
	// 开发版本警告精简为一行（可选）
	if cli.IsDev() {
		fmt.Printf("(Development build)\n")
	}
}
