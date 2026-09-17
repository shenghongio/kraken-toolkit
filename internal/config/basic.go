package config

type BasicConfig struct {
	HostInventory    map[string][]string `yaml:"host_inventory"`
	Concurrency      int                 `yaml:"concurrency"`
	SSH              BasicSSHConfig      `yaml:"ssh"`
	DefaultFetchPath string              `yaml:"default_fetch_path"`
	History          BasicConfigHistory  `yaml:"history"`
}

type BasicSSHConfig struct {
	SSHPort     int    `yaml:"ssh_port"`
	SSHUser     string `yaml:"ssh_user"`
	SSHPassword string `yaml:"ssh_password"`
	SSHHost     string `yaml:"ssh_host"`
	SSHTimeout  int    `yaml:"ssh_timeout"`
	PrivateKey  string `yaml:"private_key"`
}

type BasicConfigHistory struct {
	Enabled bool   `yaml:"enabled"`
	Log     string `yaml:"log"`
}
