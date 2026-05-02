package cmd

import (
	"bytes"
	"encoding/json"
	"github.com/sysint/internal/version"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionCmd(t *testing.T) {
	cmd := versionCmd()
	if cmd == nil {
		t.Fatal("versionCmd() 返回 nil")
	}

	// 验证命令属性
	if cmd.Use != "version" {
		t.Errorf("期望 Use = 'version', 实际 = '%s'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Short 描述不能为空")
	}
	if cmd.Long == "" {
		t.Error("Long 描述不能为空")
	}
	if cmd.RunE == nil {
		t.Error("RunE 函数不能为空")
	}

	// 验证标志
	shortFlag := cmd.Flags().Lookup("short")
	if shortFlag == nil {
		t.Error("缺少 --short 标志")
	}
	if shortFlag.Value.Type() != "bool" {
		t.Errorf("--short 标志类型应为 bool, 实际为 %s", shortFlag.Value.Type())
	}

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Error("缺少 --json 标志")
	}
	if jsonFlag.Value.Type() != "bool" {
		t.Errorf("--json 标志类型应为 bool, 实际为 %s", jsonFlag.Value.Type())
	}
}

func TestRunVersion(t *testing.T) {
	tests := []struct {
		name       string
		setupFlags func(*cobra.Command)
		validate   func(*testing.T, string, error)
	}{
		{
			name: "默认输出格式",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("short", "false")
				cmd.Flags().Set("json", "false")
			},
			validate: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Errorf("不期望错误: %v", err)
				}
				// 验证输出包含关键信息
				info := version.Get()
				expectedStrings := []string{
					"系统初始化检查工具",
					info.Version,
					info.GitCommit,
					info.GitBranch,
				}
				for _, expected := range expectedStrings {
					if !containsString(output, expected) {
						t.Errorf("输出中缺少 '%s'", expected)
					}
				}
				// 如果是开发版本，检查警告信息
				if version.IsDev() {
					if !containsString(output, "开发版本") {
						t.Error("开发版本应该显示警告信息")
					}
				}
			},
		},
		{
			name: "short 模式",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("short", "true")
				cmd.Flags().Set("json", "false")
			},
			validate: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Errorf("不期望错误: %v", err)
				}
				expected := version.Short()
				// 去除换行符进行比较
				output = trimNewline(output)
				if output != expected {
					t.Errorf("期望输出: '%s', 实际: '%s'", expected, output)
				}
			},
		},
		{
			name: "json 模式",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("short", "false")
				cmd.Flags().Set("json", "true")
			},
			validate: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Errorf("不期望错误: %v", err)
				}
				var info version.BuildInfo
				if err := json.Unmarshal([]byte(output), &info); err != nil {
					t.Errorf("JSON 解析失败: %v\n输出: %s", err, output)
				}
				expected := version.Get()
				if info.Version != expected.Version {
					t.Errorf("Version 不匹配: 期望 %s, 实际 %s", expected.Version, info.Version)
				}
				if info.GitCommit != expected.GitCommit {
					t.Errorf("GitCommit 不匹配: 期望 %s, 实际 %s", expected.GitCommit, info.GitCommit)
				}
			},
		},
		{
			name: "short 优先于 json",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("short", "true")
				cmd.Flags().Set("json", "true")
			},
			validate: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Errorf("不期望错误: %v", err)
				}
				expected := version.Short()
				output = trimNewline(output)
				if output != expected {
					t.Errorf("short 应该优先于 json, 期望: '%s', 实际: '%s'", expected, output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("short", false, "")
			cmd.Flags().Bool("json", false, "")

			tt.setupFlags(cmd)

			// 捕获输出
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runVersion(cmd, []string{})

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			os.Stdout = oldStdout

			tt.validate(t, buf.String(), err)
		})
	}
}

func TestRunVersionWithInvalidFlags(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("short", false, "")
	cmd.Flags().Bool("json", false, "")

	// 测试未设置的标志（默认应该正常工作）
	err := runVersion(cmd, []string{})
	if err != nil {
		t.Errorf("默认标志不应该返回错误: %v", err)
	}
}

// 辅助函数
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findString(s, substr)))
}

func findString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func trimNewline(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}
