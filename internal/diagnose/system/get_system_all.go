package system

import (
	"fmt"
	"github.com/kraken-pedestal/internal/models"
	"github.com/kraken-pedestal/pkg/logger"
	"github.com/shirou/gopsutil/v4/host"
	"os"
)

func GetSystemInfo() (*models.GetSystemInfoResult, error) {
	result := &models.GetSystemInfoResult{}

	// system info
	if err := getBaseSystemInfo(result); err != nil {
		return nil, logger.Wrap(logger.CodeSystem, "获取主机名称失败", err)
	}
	return result, nil
}

func getBaseSystemInfo(result *models.GetSystemInfoResult) error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("get hostname: %w", err)
	}
	hostInfo, err := host.Info()
	if err != nil {
		return fmt.Errorf("get host information: %w", err)
	}
	result.StaticHostname = hostname
	result.MachineID = hostInfo.HostID
	//result.BootID = hostInfo
	result.OperatingSystem = hostInfo.OS
	result.Kernel = hostInfo.KernelVersion
	result.Architecture = hostInfo.Platform
	return nil
}
