package logger

import (
	"io"
	"log/slog"
	"os"
)

var (
	debugLogger *slog.Logger
	levelVar    = new(slog.LevelVar)
)

func init(cfg Config) {
	levelVar.Set(
		parseLevel(cfg.Level),
	)
	var handlers []slog.Handler
	//console
	if cfg.Console {
		var w io.Writer
		switch cfg.ConsoleOutPut {
		case "stdout":
			w = os.Stdout
		default:
			w = os.Stderr
		}
	}
	if cfg.ConsoleFormat == "json" {
		handlers = append(handlers, NewJSONHandler())
	}
}
