package logger

import (
	"testing"
)

func TestInit(t *testing.T) {
	config := &LogConfig{
		Level:   "debug",
		NoColor: true,
		User:    "test",
	}

	if err := Init(config); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	if globalLogger == nil {
		t.Error("globalLogger 未初始化")
	}
}

func TestLogLevels(t *testing.T) {
	config := &LogConfig{
		Level:   "warn",
		NoColor: true,
	}

	if err := Init(config); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	// 这些日志应该被记录
	Warn("警告消息")
	Error("错误消息")

	// 这些日志不应该被记录（级别太低）
	Debug("调试消息")
	Info("信息消息")
}

func TestWithFields(t *testing.T) {
	config := &LogConfig{
		Level:   "debug",
		NoColor: true,
	}

	if err := Init(config); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	// 测试带字段的日志
	logger := GetLogger().With("key1", "value1", "key2", "value2")
	logger.Info("带字段的消息")
}

func TestWithCheck(t *testing.T) {
	config := &LogConfig{
		Level:   "debug",
		NoColor: true,
	}

	if err := Init(config); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	checkLogger := WithCheck("check-001", "磁盘检查", "storage")
	checkLogger.Info("开始检查")
}
