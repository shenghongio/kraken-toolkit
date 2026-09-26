package config

type BasicConfig struct {
	HostInventory  bool               `yaml:"host_inventory"`
	User           string             `yaml:"user"`
	PrivateKeyPath string             `yaml:"private_key_path"`
	IPList         []string           `yaml:"iplist"`
	Concurrency    int                `yaml:"concurrency"`
	Timeout        int                `yaml:"timeout"`
	RetryFile      string             `yaml:"retry_file"`
	Limit          string             `yaml:"limit"`
	Output         string             `yaml:"output"`
	Quiet          bool               `yaml:"quiet"`
	DryRun         bool               `yaml:"dry_run"`
	History        BasicConfigHistory `yaml:"history"`
	PublicKeyPath  string             `yaml:"public_key_path"`
	Bootstrap      BootstrapConfig    `yaml:"bootstrap"`
}

type BasicConfigHistory struct {
	Enabled bool   `yaml:"enabled"`
	Log     string `yaml:"log"`
}

type BootstrapConfig struct {
	User string `yaml:"user"`
}
