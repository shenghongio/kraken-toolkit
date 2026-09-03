package system

import (
	"fmt"
	"github.com/kraken-pedestal/internal/models"
	"github.com/kraken-pedestal/pkg/printer"
)

func PrintSystemInfoTable(info *models.GetSystemInfoResult) printer.Table {
	fmt.Println("System Information:")
	return printer.Table{
		Headers: []string{
			"PROPERTY",
			"VALUE",
		},
		Rows: [][]string{
			{"Hostname", info.StaticHostname},
			{"MachineID", info.MachineID},
			{"BootID", info.BootID},
			{"OS", info.OperatingSystem},
			{"Architecture", info.Architecture},
			{"Kernel", info.Kernel},
		},
	}

	//// Network
	//fmt.Println()
	//fmt.Println("Network Information:")
	//
	//return
}
