package config

// Config 是kraken的全局配置。
// 一个配置文件包含多个模块的配置
type Config struct {
	Log        LogConfig        `yaml:"log"`
	Basic      BasicConfig      `yaml:"basic"`
	Cluster    ClusterConfig    `yaml:"cluster"`
	Deploy     DeployConfig     `yaml:"deploy"`
	Middleware MiddlewareConfig `yaml:"middleware"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`
	Source string `yaml:"source"`
}
