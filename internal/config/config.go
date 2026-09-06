package config

import (
	"fmt"
	"os"
)

type ConfigList struct {
	Log        LogCnfig         `yaml:"log"`
	Basic      BasicConfig      `yaml:"basic"`
	Cluster    ClusterConfig    `yaml:"cluster"`
	Deploy     DeployConfig     `yaml:"deploy"`
	Middleware MiddlewareCofnig `yaml:"middleware"`
}

func Load(path string) (*ConfigList, error) {
	readFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg ConfigList
	
}

type LogCnfig struct{}
type BasicConfig struct{}

type ClusterConfig struct{}
type DeployConfig struct{}
type MiddlewareCofnig struct{}

func ParseConfig(filepath string) (*ConfigList, error) {

}
