package main

import (
	"errors"
	"github.com/kraken-pedestal/cmd"
	"github.com/kraken-pedestal/pkg/logger"
	"log/slog"
	"os"
)

func main() {

	// Fallback logger
	//
	// 用于 PersistentPreRunE 之前发生的错误：
	// - unknown command
	// - flag parse error
	// - command initialization error
	logger.Init(logger.ConfigOptions{
		Level:      slog.LevelInfo,
		Color:      true,
		Source:     true,
		TimeFormat: "[ 06-01-02/15:04:05 ]",
	})

	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			logger.Default().Error("程序发生 Panic", "panic", r)
			os.Exit(1)
		}
	}()

	// Cobra
	rootCmd := cmd.NewRootCmd()

	if err := rootCmd.Execute(); err != nil {

	}
}

// handleError is the single error entry point of the application
func handleError(err error) {
	if err == nil {
		return
	}

	var appErr *logger.Error
	if errors.As(err, &appErr) {
		slog.Error(appErr.Message, "error", appErr.Err, "code", appErr.Code.String())
		os.Exit(exitCode(logger.CodeUnknown))
	}
}

// exitCode converts application error codes to process exit codes,
func exitCode(code logger.Code) int {
	switch code {
	case logger.CodeInvalidArgument:
		return 2
	case logger.CodeConfiguration:
		return 3
	case logger.CodeNetwork:
		return 4
	case logger.CodePermission:
		return 5
	case logger.CodeKubernetes:
		return 6
	case logger.CodeSystem:
		return 7
	case logger.CodeContainer:
		return 8
	case logger.CodeSSH:
		return 9
	default:
		return 1
	}
}
