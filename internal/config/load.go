package config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

var (
	ErrorConfigNil   = errors.New("config is nil")
	ErrorConfigEmpty = errors.New("config is empty")
)

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
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf(`validate file "%s" error: %w`, path, err)
	}
	return &cfg, nil
}

// Validate 校验全局配置
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	// Basic 模块配置校验
	//if err := c.Basic.Validate(); err != nil {
	//	return fmt.Errorf("basic: %w", err)
	//}

	// TODO: 后续增加其他模块校验

	// c.Cluster.Validate()
	// c.Deploy.Validate()
	return nil
}
