package version

import (
	"fmt"
	"runtime"
	"time"
)

// 版本信息，通过ldflags在编译时注入
var (
	// Version 版本号
	Version = "dev"

	// GitCommit Git提交哈希
	GitCommit = "unknown"

	// GitBranch Git分支
	GitBranch = "unknown"

	// BuildTime 编译时间
	BuildTime = "unknown"

	// GoVersion Go版本
	GoVersion = runtime.Version()
)

// BuildInfo 构建信息
type BuildInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	GitBranch string `json:"git_branch"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get 获取版本信息
func Get() BuildInfo {
	return BuildInfo{
		Version:   Version,
		GitCommit: GitCommit,
		GitBranch: GitBranch,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String 返回版本字符串
func String() string {
	info := Get()
	return fmt.Sprintf("sysint version %s (commit: %s, branch: %s, build: %s, %s/%s)",
		info.Version,
		info.GitCommit,
		info.GitBranch,
		info.BuildTime,
		info.OS,
		info.Arch,
	)
}

// Short 返回简短版本字符串
func Short() string {
	return fmt.Sprintf("sysint/%s", Version)
}

// IsDev 是否是开发版本
func IsDev() bool {
	return Version == "dev" || Version == "unknown"
}

// GetBuildTime 解析构建时间
func (b BuildInfo) GetBuildTime() (time.Time, error) {
	if b.BuildTime == "unknown" {
		return time.Time{}, fmt.Errorf("build time unknown")
	}
	return time.Parse(time.RFC3339, b.BuildTime)
}
