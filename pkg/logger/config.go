package logger

type Config struct {
	LoggerLevel         Level
	LoggerConsoleFormat string
	LoggerAddsource     bool
}

func LoggerDefaultConfig() Config {
	return Config{
		LoggerLevel:         LevelInfo,
		LoggerConsoleFormat: "text",
		LoggerAddsource:     false,
	}
}
