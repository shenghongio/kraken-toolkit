package logger

type Config struct {
	// 日志登等级
	Level Level
	// 控制台输出
	Console bool
	// stdout/stderr
	ConsoleOutPut string
	//text/json
	ConsoleFormat string
	//是否开启颜色
	Color bool
	//是否显示代码位置
	Addsource bool
	//文件日志
	File string
	// lumberjack
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
}

func DefaultConfig() Config {
	return Config{
		Level:         LevelInfo,
		Console:       true,
		ConsoleOutPut: "stderr",
		ConsoleFormat: "text",
		Color:         true,
		Addsource:     false,
		MaxSize:       100,
		MaxAge:        7,
		MaxBackups:    3,
	}
}
