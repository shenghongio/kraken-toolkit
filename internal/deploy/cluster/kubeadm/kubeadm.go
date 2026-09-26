package kubeadm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Kubeadm 封装 kubeadm 二进制，负责路径发现，版本校验
// 配置生成和命令行
type Kubeadm struct {
	binaryPath      string
	configPath      string
	expectedVersion string
}

// KubeadmConfig 创建kubeadm 实例所需的参数
type KubeadmConfig struct {
	BinaryPath      string //显示指定 kubeadm 路径，为空则自动查找
	ConfigPath      string //kubeadm 配置文件写入路径
	ExpectedVersion string //期望版本(来自ClusterConfig.KubernetesVersion)
}

//NewKubeadm 创建Kubeadm 实例
// 查找 kubeadm 二进制路径的顺序
// 1. cfg.BinaryPath -- 显示指定路径
// 2. $PATH 中的kubeadm
// 3. /usr/bin/kubeadm
func NewKubeadm(cfg KubeadmConfig) (*Kubeadm, error) {
	binaryPath := cfg.BinaryPath
	if binaryPath == "" {
		p, err := exec.LookPath("kubeadm")
		if err != nil {
			binaryPath = "/usr/bin/kubeadm"
		} else {
			binaryPath = p
		}
	}
	return &Kubeadm{
		binaryPath:      binaryPath,
		configPath:      cfg.ConfigPath,
		expectedVersion: cfg.ExpectedVersion,
	}, nil
}

// BinaryPath 返回 kubeadm 二进制路径
func (k *Kubeadm) BinaryPath() string {
	return k.binaryPath
}

// ConfigFilePath 返回配置文件的路径
func (k *Kubeadm) ConfigFilePath() string {
	return k.configPath
}

// Check  校验 kubeadm 版本是否符合预期且可执行
func (k *Kubeadm) Check(ctx context.Context) error {
	info, err := os.Stat(k.binaryPath)
	if err != nil {
		return fmt.Errorf("kubeadm binary not found at %s: %w", k.binaryPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("kubeadm binary is a directory: %s", k.binaryPath)
	}
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("kubeadm binary is not executable: %s", k.binaryPath)
	}
	return nil
}

// GetVersion 获取kubeadm 版本
func (k *Kubeadm) GetVersion(ctx context.Context) (string, error) {
	if err := k.Check(ctx); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, k.binaryPath, "version", "-o", "short")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get kubeadm version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ValidateVersion 校验 kubeadm 版本是否与期望版本匹配
// kubeadm 版本偏差策略: minor 版本必须一致
// 例如 ExpectedVersion="v1.30.0",kubeadm 必须时 v1.30.x
func (k *Kubeadm) ValidateVersion(ctx context.Context) error {
	if k.expectedVersion == "" {
		return nil
	}
	actual, err := k.GetVersion(ctx)
	if err != nil {
		return err
	}
	if !versionMatch(actual, k.expectedVersion) {
		return fmt.Errorf("kubeadm version mismatch expected %s (minor),got %s", k.expectedVersion, actual)
	}
	return nil
}

// versionMatch 比较 minor 版本是否一致
// "v1.30.5" vs "v.1.30.0". -> true  （minor 版本一致）
// "v1.28.5" vs "v.1.30.0". -> false （minor 版本不一致）
func versionMatch(actual, expected string) bool {
	actual = strings.TrimPrefix(actual, "v")
	expected = strings.TrimPrefix(expected, "v")
	
	aParts := strings.SplitN(actual, ".", 3)
	eParts := strings.SplitN(expected, ".", 3)
	
	if len(aParts) < 2 || len(eParts) < 2 {
		return false
	}
	return aParts[0] == eParts[0] && aParts[1] == eParts[1]
}

// BuildPhaseCommand  构建 kubeadm phase 命令
/*
	示例:
		BuildPhaseCommand("init","preflight) -> ["kubeadm","init","phase","preflight","--config=...."]
		BuildPhaseCommand("join","kubectl-start") -> ["kubeadm","join","phase","kubectl-start","--config=...."]
*/
func (k *Kubeadm) BuildPhaseCommand(action, phaseName string, extrArgs ...string) []string {
	args := []string{k.binaryPath, action, "phase", phaseName}
	if k.configPath != "" {
		args = append(args, "--config", k.configPath)
	}
	args = append(args, extrArgs...)
	return args
}

//BuildInitPhaseCommand 构建kubeadm init phase 命令
func (k *Kubeadm) BuildInitPhaseCommand(phaseName string, extrArgs ...string) []string {
	return k.BuildPhaseCommand("init", phaseName, extrArgs...)
}

//BuildJoinPhaseCommand 构建kubeadm join phase 命令
func (k *Kubeadm) BuildJoinPhaseCommand(phaseName string, extraArgs ...string) []string {
	return k.BuildPhaseCommand("join", phaseName, extraArgs...)
}

//RunCommand 在本地执行kubeadm命令，返回 stdout,stderr和错误
func (k *Kubeadm) RunCommand(ctx context.Context, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, k.binaryPath, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
	
}
