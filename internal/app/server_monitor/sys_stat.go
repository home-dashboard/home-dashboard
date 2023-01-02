package server_monitor

import (
	psuDisk "github.com/shirou/gopsutil/v3/disk"
	psuMem "github.com/shirou/gopsutil/v3/mem"
	psuNet "github.com/shirou/gopsutil/v3/net"
	"golang.org/x/sys/windows"
	"strings"
	"unsafe"
)

func GetSystemStatic() *SystemStat {
	systemStat := SystemStat{
		Memory:  SystemMemoryStat{},
		Network: map[string]*SystemNetworkStat{},
		Disk:    map[string]*SystemDiskStat{},
	}

	vm, err := psuMem.VirtualMemory()
	if err == nil {
		systemStat.Memory.VirtualMemory = vm
		//if err := mergo.Merge(&systemStat.Memory.VirtualMemory, vm); err != nil {
		//	fmt.Println("virtual memory merge failed. ", err)
		//}
	}

	sm, err := psuMem.SwapMemory()
	if err == nil {
		systemStat.Memory.SwapMemory = sm
		//if err := mergo.Merge(&systemStat.Memory.SwapMemory, sm); err != nil {
		//	fmt.Println("swap memory merge failed. ", err)
		//}
	}

	ifMap := map[int]string{}

	ifs, _ := psuNet.Interfaces()

	for _, item := range ifs {
		networkStat := SystemNetworkStat{
			InterfaceStat: &SystemNetworkInterfaceStat{
				Stat:        item,
				Type:        0,
				Description: "",
			},
		}
		systemStat.Network[item.Name] = &networkStat

		ifMap[item.Index] = item.Name

		//if err := mergo.Merge(&networkStat.InterfaceStat, item); err != nil {
		//	fmt.Println("network interface merge failed. ", err)
		//}
	}

	netIOs, _ := psuNet.IOCounters(true)

	for _, item := range netIOs {
		networkStat := systemStat.Network[item.Name]

		networkStat.IoStat = item
		//if err := mergo.Merge(&networkStat.IoStat, item); err != nil {
		//	fmt.Println("network io merge failed. ", err)
		//}
	}

	ais, _ := getAdaptersInfo()

	for _, item := range ais {
		var name = ifMap[item.Index]
		networkStat := systemStat.Network[name]

		networkStat.InterfaceStat.Type = item.Type
		networkStat.InterfaceStat.Description = item.Description
	}

	parts, _ := psuDisk.Partitions(true)

	for _, item := range parts {
		diskStat := SystemDiskStat{
			PartitionStat: item,
		}

		systemStat.Disk[item.Device] = &diskStat
		//if err := mergo.Merge(&diskStat.PartitionStat, item); err != nil {
		//	fmt.Println("disk partition merge failed. ", err)
		//}

		usage, _ := psuDisk.Usage(item.Device)
		diskStat.UsageStat = usage
		//if err := mergo.Merge(&diskStat.UsageStat, usage); err != nil {
		//	fmt.Println("disk usage merge failed. ", err)
		//}
	}

	diskIOs, _ := psuDisk.IOCounters()

	for _, item := range diskIOs {
		diskStat := systemStat.Disk[item.Name]
		diskStat.IoStat = item

		//if err := mergo.Merge(&diskStat.IoStat, item); err != nil {
		//	fmt.Println("disk io merge failed. ", err)
		//}
	}

	return &systemStat
}

type SystemStat struct {
	Memory  SystemMemoryStat              `json:"memory"`
	Network map[string]*SystemNetworkStat `json:"network"`
	Disk    map[string]*SystemDiskStat    `json:"disk"`
}

type SystemNetworkStat struct {
	InterfaceStat *SystemNetworkInterfaceStat `json:"interfaceStat"`
	IoStat        psuNet.IOCountersStat       `json:"ioStat"`
}

type SystemNetworkInterfaceStat struct {
	Stat        psuNet.InterfaceStat `json:"stat"`
	Type        int                  `json:"type"`
	Description string               `json:"description"`
}

type SystemMemoryStat struct {
	VirtualMemory *psuMem.VirtualMemoryStat `json:"virtualMemory"`
	SwapMemory    *psuMem.SwapMemoryStat    `json:"swapMemory"`
}

type SystemDiskStat struct {
	PartitionStat psuDisk.PartitionStat  `json:"partitionStat"`
	UsageStat     *psuDisk.UsageStat     `json:"usageStat"`
	IoStat        psuDisk.IOCountersStat `json:"ioStat"`
}

func getAdaptersInfo() (SystemNetworkAdapterInfoList, error) {
	var nai []SystemNetworkAdapterInfo
	var ol uint32
	err := windows.GetAdaptersInfo(nil, &ol)
	if err == nil || ol == 0 {
		return nai, err
	}
	buf := make([]byte, int(ol))
	ai := (*windows.IpAdapterInfo)(unsafe.Pointer(&buf[0]))
	if err := windows.GetAdaptersInfo(ai, &ol); err != nil {
		return nai, err
	}

	for ; ai != nil; ai = ai.Next {
		nai = append(nai, SystemNetworkAdapterInfo{
			Index:       int(ai.Index),
			Description: strings.Trim(string(ai.Description[:]), " \t\n\000"),
			Type:        int(ai.Type),
		})
	}

	return nai, err
}

type SystemNetworkAdapterInfo struct {
	Index       int
	Description string
	Type        int
}
type SystemNetworkAdapterInfoList []SystemNetworkAdapterInfo
