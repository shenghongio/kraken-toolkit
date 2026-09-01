package models

type GetSystemInfoResult struct {

	//System
	StaticHostname  string `mapstructure:"static_hostname"`
	MachineID       string `mapstructure:"machine_id"`
	BootID          string `mapstructure:"boot_id"`
	OperatingSystem string `mapstructure:"operating_system"`
	Kernel          string `mapstructure:"kernel"`
	Architecture    string `mapstructure:"architecture"`

	//Network
	Network NetworkInfo `mapstructure:"network"`
}

// NetworkInfo represents network information
type NetworkInfo struct {
	// Hostname reported by the network subsystem
	Hostname    string          `mapstructure:"hostname"`
	Ineterfaces []InterfaceInfo `mapstructure:"interfaces"`
}

// InterfaceInfo represents a network interface
type InterfaceInfo struct {
	// Interface name, e.g. eth0
	Name string `mapstructure:"name"`
	// MAC addresses.
	MacAddresses string `mapstructure:"mac_addresses"`
	//IPv4/IPv6 Addresses
	Addresses []string `mapstructure:"addresses"`
	// Maximum transmission unit
	MTU int `mapstructure:"mtu"`
	// Interface state, e.g. UP/DOWN.
	State string `json:"state" mapstructure:"state"`
}
