package logger

import (
	"github.com/fatih/color"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// LogLevel 日志级别类型
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// LogFormat 日志格式
type LogFormat string

const (
	FormatText LogFormat = "text" // 人类可读，带颜色
	FormatJSON LogFormat = "json" // 结构化 JSON
)

// Config 日志配置
type Config struct {
	Level      LogLevel  // 日志级别
	Format     LogFormat // 输出格式
	NoColor    bool      // 禁用颜色
	OutputPath string    // 输出路径：stdout, stderr, 或文件路径
	AddSource  bool      // 是否显示调用位置（文件:行号）
	MaxSize    int       // 日志文件最大大小（MB），仅当 OutputPath 为文件时有效
	MaxBackups int       // 保留旧文件的最大个数
	MaxAge     int       // 保留旧文件的最大天数
	Compress   bool      // 是否压缩旧文件
}

var defaultLogger *slog.Logger

// Init 初始化全局日志器
func Init(cfg *Config) error {
	if cfg == nil {
		cfg = &Config{
			Level:      LevelInfo,
			Format:     FormatText,
			NoColor:    false,
			OutputPath: "stderr",
			AddSource:  false,
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}
	}

	// 解析日志级别
	var slogLevel slog.Level
	switch cfg.Level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// 配置输出 writer
	writer, err := getWriter(cfg.OutputPath, cfg)
	if err != nil {
		return err
	}

	// 自定义 Handler 选项
	handlerOpts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: cfg.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 时间格式化
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05"))
			}
			// 级别文本大写并着色
			if a.Key == slog.LevelKey {
				level := a.Value.String()
				// 转为大写
				//level = strings.ToUpper(level)
				level = strings.ToLower(level)
				if !cfg.NoColor {
					switch level {
					case "error":
						level = color.RedString(level)
					case "warn":
						level = color.YellowString(level)
					case "info":
						level = color.GreenString(level)
					case "debug":
						level = color.CyanString(level)
					}
				}
				a.Value = slog.StringValue(level)
			}
			// 如果启用了 AddSource，且 key 为 source，
			if cfg.AddSource && a.Key == slog.SourceKey {
				if src, ok := a.Value.Any().(*slog.Source); ok && src != nil {
					file := filepath.Base(src.File)
					a.Value = slog.StringValue(file + ":" + itoa(src.Line))
				}
			}
			return a
		},
	}

	var handler slog.Handler
	if cfg.Format == FormatJSON {
		handler = slog.NewJSONHandler(writer, handlerOpts)
	} else {
		handler = slog.NewTextHandler(writer, handlerOpts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
	return nil
}

// getWriter 根据输出路径返回 io.Writer
func getWriter(path string, cfg *Config) (io.Writer, error) {
	switch path {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		// 确保目录存在
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
		// 使用 lumberjack 实现日志滚动
		return &lumberjack.Logger{
			Filename:   path,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}, nil
	}
}

// itoa 整数转字符串（避免 fmt.Sprintf 开销）
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

// ----- 对外导出的日志函数（兼容已有调用）-----

// Debug 输出 Debug 级别日志
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

// Info 输出 Info 级别日志
func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

// Warn 输出 Warn 级别日志
func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

// Error 输出 Error 级别日志
func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

// With 返回带有预设字段的子 Logger
func With(args ...any) *slog.Logger {
	return defaultLogger.With(args...)
}

// Default 返回当前默认 logger（用于外部检查）
func Default() *slog.Logger {
	return defaultLogger
}

// SetLevel 动态修改日志级别（可用于运行时调整）
func SetLevel(level LogLevel) {
	var l slog.Level
	switch level {
	case LevelDebug:
		l = slog.LevelDebug
	case LevelInfo:
		l = slog.LevelInfo
	case LevelWarn:
		l = slog.LevelWarn
	case LevelError:
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	slog.SetLogLoggerLevel(l)
}
