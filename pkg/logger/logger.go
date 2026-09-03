package logger

import (
	"fmt"
	consle "github.com/phsym/console-slog"
	"log/slog"
	"os"
)

type ConfigOptions struct {
	Level slog.Level
	// 是否开启颜色 开启自定义颜色设置
	Color bool
	// 是否显示source
	Source bool
	// 时间格式
	TimeFormat string
}

func Init(cfg ConfigOptions) *slog.Logger {
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = "01-02/15:04:05"
	}
	handler := consle.NewHandler(os.Stderr, &consle.HandlerOptions{
		AddSource:  cfg.Source,
		Level:      cfg.Level,
		NoColor:    true,
		TimeFormat: cfg.TimeFormat,
	},
	)
	log := slog.New(handler)
	slog.SetDefault(log)
	return log
}

func Default() *slog.Logger {
	return slog.Default()
}

// ParseLevel Cobra 参数是字符串 --log-level debug 需要一个转换函数。
func ParseLevel(value string) (slog.Level, error) {
	switch value {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid log level: %s", value)
	}
}
