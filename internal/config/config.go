package config

type GlobalFlags struct {
	Config string
}

type Config struct {
	//Log        LogConfig        `yaml:"log"`
	Basic      BasicConfig      `yaml:"basic"`
	Cluster    ClusterConfig    `yaml:"cluster"`
	Deploy     DeployConfig     `yaml:"deploy"`
	Middleware MiddlewareConfig `yaml:"middleware"`
}
