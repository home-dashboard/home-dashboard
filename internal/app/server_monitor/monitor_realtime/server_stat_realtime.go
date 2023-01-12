package monitor_realtime

import (
	psuCpu "github.com/shirou/gopsutil/v3/cpu"
	psuDisk "github.com/shirou/gopsutil/v3/disk"
	psuHost "github.com/shirou/gopsutil/v3/host"
	psuMem "github.com/shirou/gopsutil/v3/mem"
	psuNet "github.com/shirou/gopsutil/v3/net"
	"golang.org/x/sys/windows"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

func getSystemRealtimeStatic() *SystemRealtimeStat {
	systemStat := SystemRealtimeStat{
		Memory:  SystemMemoryStat{},
		Network: map[string]*SystemNetworkStat{},
		Disk:    map[string]*SystemDiskStat{},
		Cpu:     map[string]*SystemCpuStat{},
		Host:    SystemHostStat{},
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

	ifMap := map[uint32]string{}

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

		ifMap[uint32(item.Index)] = item.Name

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

		usage, _ := psuDisk.Usage(item.Device)
		diskStat.UsageStat = usage
	}

	diskIOs, _ := psuDisk.IOCounters()

	for _, item := range diskIOs {
		diskStat := systemStat.Disk[item.Name]
		diskStat.IoStat = item
	}

	cpuStat := SystemCpuStat{}

	cpuInfos, _ := psuCpu.Info()

	for _, item := range cpuInfos {
		cpuStat.InfoStat = item
	}

	cpuStat.LogicalCounts, _ = psuCpu.Counts(true)
	cpuStat.PhysicalCounts, _ = psuCpu.Counts(false)
	cpuStat.Percents, _ = psuCpu.Percent(time.Second, true)

	systemStat.Cpu[strconv.FormatInt(int64(cpuStat.InfoStat.CPU), 10)] = &cpuStat

	systemStat.Host.InfoStat, _ = psuHost.Info()

	return &systemStat
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
			Index:       ai.Index,
			Description: strings.Trim(string(ai.Description[:]), " \t\n\000"),
			Type:        int(ai.Type),
		})
	}

	return nai, err
}

//var SystemRealtimeStatCacheKey = &SystemRealtimeStatKey

var systemStat = getSystemRealtimeStatic()

var ticker *time.Ticker
var done = make(chan bool)

func StartSystemRealtimeStatLoop(d time.Duration) {
	ticker = time.NewTicker(d)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				systemStat = getSystemRealtimeStatic()
			}
		}
	}()
}

func StopSystemRealtimeStatLoop() {
	ticker.Stop()
	done <- true
}

func GetCachedSystemRealtimeStat() *SystemRealtimeStat {
	return systemStat
}
