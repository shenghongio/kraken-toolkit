package executor

import (
	"encoding/json"
	"runtime"
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()

	// 验证基本字段不为空
	if info.Version == "" {
		t.Error("Version 字段不能为空")
	}
	if info.GitCommit == "" {
		t.Error("GitCommit 字段不能为空")
	}
	if info.GitBranch == "" {
		t.Error("GitBranch 字段不能为空")
	}
	if info.BuildTime == "" {
		t.Error("BuildTime 字段不能为空")
	}
	if info.GoVersion == "" {
		t.Error("GoVersion 字段不能为空")
	}
	if info.OS == "" {
		t.Error("OS 字段不能为空")
	}
	if info.Arch == "" {
		t.Error("Arch 字段不能为空")
	}

	// 验证运行时信息
	if info.OS != runtime.GOOS {
		t.Errorf("OS 不匹配: 期望 %s, 实际 %s", runtime.GOOS, info.OS)
	}
	if info.Arch != runtime.GOARCH {
		t.Errorf("Arch 不匹配: 期望 %s, 实际 %s", runtime.GOARCH, info.Arch)
	}
	if info.GoVersion != runtime.Version() {
		t.Errorf("GoVersion 不匹配: 期望 %s, 实际 %s", runtime.Version(), info.GoVersion)
	}
}

func TestString(t *testing.T) {
	info := Get()
	str := String()

	// 验证字符串包含关键信息
	expectedParts := []string{
		"sysint version",
		info.Version,
		info.GitCommit,
		info.GitBranch,
		info.BuildTime,
		info.OS,
		info.Arch,
	}

	for _, part := range expectedParts {
		if !contains(str, part) {
			t.Errorf("String() 输出中缺少 '%s'，实际输出: %s", part, str)
		}
	}
}

func TestShort(t *testing.T) {
	short := Short()
	expected := "sysint/" + Version

	if short != expected {
		t.Errorf("Short() 期望: %s, 实际: %s", expected, short)
	}
}

func TestIsDev(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"开发版本 dev", "dev", true},
		{"开发版本 unknown", "unknown", true},
		{"正式版本 v1.0.0", "v1.0.0", false},
		{"正式版本 1.0.0", "1.0.0", false},
		{"正式版本 v2.0.0-alpha", "v2.0.0-alpha", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始值
			originalVersion := Version
			defer func() { Version = originalVersion }()

			// 设置测试版本
			Version = tt.version

			result := IsDev()
			if result != tt.expected {
				t.Errorf("IsDev() 期望: %v, 实际: %v", tt.expected, result)
			}
		})
	}
}

func TestBuildInfoGetBuildTime(t *testing.T) {
	tests := []struct {
		name      string
		buildTime string
		wantError bool
	}{
		{"有效时间", "2024-01-15T10:30:00Z", false},
		{"未知时间", "unknown", true},
		{"无效格式", "2024-01-15", true},
		{"空字符串", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := BuildInfo{BuildTime: tt.buildTime}
			parsedTime, err := info.GetBuildTime()

			if tt.wantError {
				if err == nil {
					t.Error("期望返回错误，但没有错误")
				}
			} else {
				if err != nil {
					t.Errorf("不期望错误，但得到: %v", err)
				}
				if parsedTime.IsZero() {
					t.Error("解析的时间不应该为零值")
				}
			}
		})
	}
}

func TestJSONSerialization(t *testing.T) {
	info := Get()

	// 序列化
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}

	// 反序列化
	var decoded BuildInfo
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("JSON 反序列化失败: %v", err)
	}

	// 验证字段
	if decoded.Version != info.Version {
		t.Errorf("Version 不匹配: 期望 %s, 实际 %s", info.Version, decoded.Version)
	}
	if decoded.GitCommit != info.GitCommit {
		t.Errorf("GitCommit 不匹配: 期望 %s, 实际 %s", info.GitCommit, decoded.GitCommit)
	}
	if decoded.GitBranch != info.GitBranch {
		t.Errorf("GitBranch 不匹配: 期望 %s, 实际 %s", info.GitBranch, decoded.GitBranch)
	}
	if decoded.BuildTime != info.BuildTime {
		t.Errorf("BuildTime 不匹配: 期望 %s, 实际 %s", info.BuildTime, decoded.BuildTime)
	}
}

func TestBuildInfoJSONTags(t *testing.T) {
	info := BuildInfo{
		Version:   "1.0.0",
		GitCommit: "abc123",
		GitBranch: "main",
		BuildTime: "2024-01-15T10:30:00Z",
		GoVersion: "go1.21.5",
		OS:        "linux",
		Arch:      "amd64",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}

	// 验证 JSON 字段名使用下划线格式
	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	if err != nil {
		t.Fatalf("JSON 反序列化失败: %v", err)
	}

	// 检查 JSON 字段名
	expectedFields := []string{
		"version",
		"git_commit",
		"git_branch",
		"build_time",
		"go_version",
		"os",
		"arch",
	}

	for _, field := range expectedFields {
		if _, exists := jsonMap[field]; !exists {
			t.Errorf("JSON 中缺少字段: %s", field)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Get()
	}
}

func BenchmarkString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		String()
	}
}

func BenchmarkShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Short()
	}
}

// 辅助函数
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
