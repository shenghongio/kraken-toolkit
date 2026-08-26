package logger

import (
	"context"
	"io"
	"log/slog"
)

type ConsoleOptions struct {
	// 是否开启颜色
	Color bool
	// 是否显示source
	Source bool
	// 时间格式
	TimeFormat string
}

type ConsoleHandler struct {
	writer io.Writer
	level  slog.Leveler
	opts   ConsoleOptions
	attrs  []slog.Attr
	groups []string
}

func (c ConsoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	//TODO implement me
	panic("implement me")
}

func (c ConsoleHandler) Handle(ctx context.Context, record slog.Record) error {
	//TODO implement me
	panic("implement me")
}

func (c ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	//TODO implement me
	panic("implement me")
}

func (c ConsoleHandler) WithGroup(name string) slog.Handler {
	//TODO implement me
	panic("implement me")
}

func NewConsoleHandler(writer io.Writer, level slog.Leveler, opst ConsoleOptions) slog.Handler {
	if opst.TimeFormat == "" {
		opst.TimeFormat = "15:04:05"
	}
	return &ConsoleHandler{
		writer: writer,
		level:  level,
		opts:   opst,
	}
	
}
