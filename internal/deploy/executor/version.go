package executor

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// 版本信息，通过ldflags在编译时注入
var (
	// Version 版本号
	version = "dev"
	
	// GitCommit Git提交哈希
	gitCommit = "unknown"
	
	// GitBranch Git分支
	gitBranch = "unknown"
	
	// BuildTime 编译时间
	buildTime = "unknown"
	
	// GoVersion Go版本
	goVersion = runtime.Version()
)

// BuildInfo 构建信息
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	GitBranch string `json:"git_branch"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get 获取版本信息
func Get() VersionInfo {
	return VersionInfo{
		Version:   version,
		GitCommit: gitCommit,
		GitBranch: gitBranch,
		BuildTime: buildTime,
		GoVersion: goVersion,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String 返回版本字符串
func PrintString() string {
	info := Get()
	return fmt.Sprintf("kraken version %s (commit: %s, branch: %s, build: %s, %s/%s)",
		info.Version,
		info.GitCommit,
		info.GitBranch,
		info.BuildTime,
		info.OS,
		info.Arch,
	)
}

// IsDev 判断是否为开发版本（版本号包含 "dev" 或未设置）
func IsDev() bool {
	// 如果版本号是 "dev" 或包含 "-dev" 后缀（常见于预发布版本），视为开发版
	return strings.Contains(version, "dev") || version == "dev"
}

// GetBuildTime 解析构建时间
func (b VersionInfo) GetBuildTime() (time.Time, error) {
	if b.BuildTime == "unknown" {
		return time.Time{}, fmt.Errorf("build time unknown")
	}
	return time.Parse(time.RFC3339, b.BuildTime)
}
