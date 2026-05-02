package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	globalLogger *Logger
	once         sync.Once
)

// 颜色定义
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// LogConfig 日志配置
type LogConfig struct {
	Level     string // 日志级别: debug, info, warn, error
	AddSource bool   // 是否添加调用位置信息
	NoColor   bool   // 是否禁用颜色
	User      string // 当前用户
	Output    string // 输出目标: stdout, stderr, 或文件路径
}

// Logger 全局日志管理器
type Logger struct {
	logger *slog.Logger
	config *LogConfig
	ctx    context.Context
	pid    int
	gid    uint64
	user   string
}

// CheckLogger 检查项日志
type CheckLogger struct {
	logger    *slog.Logger
	checkID   string
	checkName string
	checkType string
}

// CustomHandler 自定义日志处理器
type CustomHandler struct {
	writer  io.Writer
	level   slog.Level
	noColor bool
	pid     int
	gid     uint64
	user    string
}

// NewCustomHandler 创建自定义处理器
func NewCustomHandler(writer io.Writer, level slog.Level, noColor bool, pid int, gid uint64, user string) *CustomHandler {
	return &CustomHandler{
		writer:  writer,
		level:   level,
		noColor: noColor,
		pid:     pid,
		gid:     gid,
		user:    user,
	}
}

// getLevelColor 获取日志级别对应的颜色
func (h *CustomHandler) getLevelColor(level string) string {
	if h.noColor {
		return ""
	}
	switch level {
	case "DEBUG":
		return colorCyan
	case "INFO":
		return colorGreen
	case "WARN":
		return colorYellow
	case "ERROR":
		return colorRed
	default:
		return colorReset
	}
}

// getMessageColor 获取消息内容的颜色（与级别相同）
func (h *CustomHandler) getMessageColor(level string) string {
	if h.noColor {
		return ""
	}
	switch level {
	case "DEBUG":
		return colorCyan
	case "INFO":
		return colorGreen
	case "WARN":
		return colorYellow
	case "ERROR":
		return colorRed
	default:
		return colorReset
	}
}

// Enabled 判断日志级别是否启用
func (h *CustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle 处理日志记录
func (h *CustomHandler) Handle(_ context.Context, r slog.Record) error {
	// 格式: [时间] - [GID] - [user] - [级别] | 消息 key=value; key=value
	timestamp := r.Time.Format("2006-01-02T15:04:05.000Z07:00")
	
	// 获取日志级别
	level := r.Level.String()
	switch r.Level {
	case slog.LevelDebug:
		level = "DEBUG"
	case slog.LevelInfo:
		level = "INFO"
	case slog.LevelWarn:
		level = "WARN"
	case slog.LevelError:
		level = "ERROR"
	}
	
	// 获取级别颜色
	levelColor := h.getLevelColor(level)
	msgColor := h.getMessageColor(level)
	
	// 获取消息
	msg := r.Message
	
	// 构建属性字符串（使用分号分隔）
	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		// 格式化属性值
		var val string
		switch v := a.Value.Any().(type) {
		case string:
			val = v
		case int, int64, int32, uint, uint64:
			val = fmt.Sprintf("%d", v)
		case float64, float32:
			val = fmt.Sprintf("%.2f", v)
		case bool:
			val = fmt.Sprintf("%t", v)
		case time.Duration:
			val = v.String()
		case error:
			val = v.Error()
		default:
			val = fmt.Sprintf("%v", v)
		}
		attrs = append(attrs, fmt.Sprintf("%s=%s", a.Key, val))
		return true
	})
	
	// 构建输出
	var output string
	attrStr := ""
	if len(attrs) > 0 {
		attrStr = strings.Join(attrs, "; ")
	}
	
	if h.noColor {
		// 无颜色输出
		if attrStr != "" {
			output = fmt.Sprintf("[%s] - [%d] - [%s] - [%s] | %s %s\n",
				timestamp, h.gid, h.user, level, msg, attrStr)
		} else {
			output = fmt.Sprintf("[%s] - [%d] - [%s] - [%s] | %s\n",
				timestamp, h.gid, h.user, level, msg)
		}
	} else {
		// 彩色输出
		if attrStr != "" {
			output = fmt.Sprintf("%s[%s]%s - %s[%d]%s - %s[%s]%s - %s[%s]%s | %s%s%s %s%s%s\n",
				colorWhite, timestamp, colorReset,
				colorCyan, h.gid, colorReset,
				colorBlue, h.user, colorReset,
				levelColor, level, colorReset,
				msgColor, msg, colorReset,
				msgColor, attrStr, colorReset)
		} else {
			output = fmt.Sprintf("%s[%s]%s - %s[%d]%s - %s[%s]%s - %s[%s]%s | %s%s%s\n",
				colorWhite, timestamp, colorReset,
				colorCyan, h.gid, colorReset,
				colorBlue, h.user, colorReset,
				levelColor, level, colorReset,
				msgColor, msg, colorReset)
		}
	}
	
	_, err := h.writer.Write([]byte(output))
	return err
}

// WithAttrs 返回带有额外属性的处理器
func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

// WithGroup 返回带有分组的处理器
func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return h
}

// getGID 获取 goroutine ID
func getGID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// 解析 goroutine ID
	// 格式: "goroutine 123 [running]:"
	stack := string(buf[:n])
	var gid uint64
	fmt.Sscanf(stack, "goroutine %d", &gid)
	return gid
}

// Init 初始化全局日志
func Init(config *LogConfig) error {
	var initErr error
	once.Do(func() {
		initErr = initLogger(config)
	})
	return initErr
}

// initLogger 实际的初始化逻辑
func initLogger(config *LogConfig) error {
	// 解析日志级别
	var level slog.Level
	switch strings.ToLower(config.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	
	// 获取 PID 和 GID
	pid := os.Getpid()
	gid := getGID()
	user := config.User
	if user == "" {
		user = os.Getenv("USER")
		if user == "" {
			user = os.Getenv("USERNAME")
			if user == "" {
				user = "unknown"
			}
		}
	}
	
	// 确定输出目标
	var writer io.Writer
	switch config.Output {
	case "stderr":
		writer = os.Stderr
	case "":
		fallthrough
	default:
		writer = os.Stdout
	}
	
	// 创建自定义处理器
	handler := NewCustomHandler(writer, level, config.NoColor, pid, gid, user)
	
	globalLogger = &Logger{
		logger: slog.New(handler),
		config: config,
		ctx:    context.Background(),
		pid:    pid,
		gid:    gid,
		user:   user,
	}
	
	// 设置为默认logger
	slog.SetDefault(globalLogger.logger)
	return nil
}

// GetLogger 获取全局日志实例
func GetLogger() *Logger {
	if globalLogger == nil {
		_ = Init(&LogConfig{
			Level:   "info",
			NoColor: false,
		})
	}
	return globalLogger
}

// WithCheck 为检查项创建带字段的日志
func (l *Logger) WithCheck(checkID, checkName, checkType string) *CheckLogger {
	return &CheckLogger{
		logger: l.logger.With(
			slog.String("check_id", checkID),
			slog.String("check_name", checkName),
			slog.String("check_type", checkType),
		),
		checkID:   checkID,
		checkName: checkName,
		checkType: checkType,
	}
}

// WithHost 为主机创建带字段的日志
func (l *Logger) WithHost(host string) *Logger {
	return &Logger{
		logger: l.logger.With(slog.String("host", host)),
		config: l.config,
		ctx:    l.ctx,
		pid:    l.pid,
		gid:    l.gid,
		user:   l.user,
	}
}

// WithContext 从context中提取日志字段
func (l *Logger) WithContext(ctx context.Context) *Logger {
	attrs := []any{}
	
	if checkID, ok := ctx.Value("check_id").(string); ok {
		attrs = append(attrs, slog.String("check_id", checkID))
	}
	if checkName, ok := ctx.Value("check_name").(string); ok {
		attrs = append(attrs, slog.String("check_name", checkName))
	}
	if host, ok := ctx.Value("host").(string); ok {
		attrs = append(attrs, slog.String("host", host))
	}
	
	return &Logger{
		logger: l.logger.With(attrs...),
		config: l.config,
		ctx:    ctx,
		pid:    l.pid,
		gid:    l.gid,
		user:   l.user,
	}
}

// Debug 输出Debug级别日志
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info 输出Info级别日志
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn 输出Warn级别日志
func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error 输出Error级别日志
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// With 添加自定义字段
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		logger: l.logger.With(args...),
		config: l.config,
		ctx:    l.ctx,
		pid:    l.pid,
		gid:    l.gid,
		user:   l.user,
	}
}

// Debug 输出Debug级别日志（CheckLogger）
func (cl *CheckLogger) Debug(msg string, args ...any) {
	cl.logger.Debug(msg, args...)
}

// Info 输出Info级别日志（CheckLogger）
func (cl *CheckLogger) Info(msg string, args ...any) {
	cl.logger.Info(msg, args...)
}

// Warn 输出Warn级别日志（CheckLogger）
func (cl *CheckLogger) Warn(msg string, args ...any) {
	cl.logger.Warn(msg, args...)
}

// Error 输出Error级别日志（CheckLogger）
func (cl *CheckLogger) Error(msg string, args ...any) {
	cl.logger.Error(msg, args...)
}

// 全局便捷函数
func Debug(msg string, args ...any) {
	GetLogger().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	GetLogger().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	GetLogger().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	GetLogger().Error(msg, args...)
}

func WithCheck(checkID, checkName, checkType string) *CheckLogger {
	return GetLogger().WithCheck(checkID, checkName, checkType)
}

func WithHost(host string) *Logger {
	return GetLogger().WithHost(host)
}
