package config

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type GlobalFlags struct {
	Config string
}

var (
	ErrorConfigEmpty = errors.New("config is empty")
)

type Config struct {
	//Log        LogConfig        `yaml:"log"`
	Basic      BasicConfig      `yaml:"basic"`
	Cluster    ClusterConfig    `yaml:"cluster"`
	Deploy     DeployConfig     `yaml:"deploy"`
	Middleware MiddlewareConfig `yaml:"middleware"`
}

// contextKey 是包内私有的类型，防止其他包意外使用或修改
type contextKey string

// ContextKeyConfig 用于在 context 中存取 *Config
const ContextKeyConfig = contextKey("config")

// Load loads and validates a Kraken configuration file.
//
// The function performs:
//
//	1. Validate path
//	2. Read file
//	3. Parse YAML
//	4. Validate configuration

func Load(path string) (*Config, error) {
	// 1.校验配置文件路径
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, ErrorConfigEmpty
	}

	// 2.读取配置文件
	readFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(`read config file "%q" error: %w`, path, err)
	}

	//3，yaml解析
	var cfg Config
	if err := yaml.Unmarshal(readFile, &cfg); err != nil {
		return nil, fmt.Errorf(`unmarshal file "%s" error: %w`, path, err)
	}
	//4. 配置校验
	//if err := cfg.Validate(); err != nil {
	//	return nil, fmt.Errorf(`validate file "%s" error: %w`, path, err)
	//}
	return &cfg, nil
}

type DeployConfig struct {
	//TODO: 具体配置
}

type ClusterConfig struct {
	//TODO: 实现具体配置
}

type MiddlewareConfig struct {
	//TODO: 具体配置
}

func FromContext(ctx context.Context) (*Config, bool) {
	if ctx == nil {
		return nil, false
	}
	v := ctx.Value(ContextKeyConfig)
	if v == nil {
		return nil, false
	}
	cfg, ok := v.(*Config)
	return cfg, ok
}
