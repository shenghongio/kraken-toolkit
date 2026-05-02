package logger

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// DefaultConfig 返回默认配置
func DefaultConfig() *LogConfig {
	return &LogConfig{
		Level:     "info",
		AddSource: false,
		NoColor:   true,
		User:      getCurrentUser(),
		Output:    "stdout",
	}
}

// getCurrentUser 获取当前用户
func getCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *LogConfig {
	config := DefaultConfig()

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = level
	}

	if noColor := os.Getenv("LOG_NO_COLOR"); noColor != "" {
		if val, err := strconv.ParseBool(noColor); err == nil {
			config.NoColor = val
		}
	}

	if output := os.Getenv("LOG_OUTPUT"); output != "" {
		config.Output = output
	}

	if user := os.Getenv("LOG_USER"); user != "" {
		config.User = user
	}

	return config
}

// LoadConfigFromFlags 从命令行标志加载配置
func LoadConfigFromFlags() *LogConfig {
	config := DefaultConfig()

	// 定义命令行标志（这些标志需要在主函数中解析）
	logLevel := flag.String("log-level", config.Level, "日志级别 (debug, info, warn, error)")
	logNoColor := flag.Bool("log-no-color", config.NoColor, "禁用彩色输出")
	logOutput := flag.String("log-output", config.Output, "日志输出 (stdout, stderr)")
	logUser := flag.String("log-user", config.User, "日志中显示的用户名")

	// 注意：这里不调用 flag.Parse()，让调用者处理

	config.Level = *logLevel
	config.NoColor = *logNoColor
	config.Output = *logOutput
	config.User = *logUser

	return config
}

// Validate 验证配置
func (c *LogConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLevels[c.Level] {
		return &ConfigError{Field: "level", Value: c.Level, Message: "无效的日志级别"}
	}

	validOutputs := map[string]bool{
		"stdout": true,
		"stderr": true,
		"":       true,
	}

	if !validOutputs[c.Output] && c.Output != "" {
		return &ConfigError{Field: "output", Value: c.Output, Message: "无效的输出目标"}
	}

	return nil
}

// ConfigError 配置错误
type ConfigError struct {
	Field   string
	Value   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("配置错误: %s=%s, %s", e.Field, e.Value, e.Message)
}
