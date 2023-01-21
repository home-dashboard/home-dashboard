package monitor_realtime

import (
	"github.com/jinzhu/copier"
	psuCpu "github.com/shirou/gopsutil/v3/cpu"
	psuDisk "github.com/shirou/gopsutil/v3/disk"
	psuHost "github.com/shirou/gopsutil/v3/host"
	psuMem "github.com/shirou/gopsutil/v3/mem"
	psuNet "github.com/shirou/gopsutil/v3/net"
	"log"
	"strconv"
	"time"
)

func getSystemRealtimeStatic() *SystemRealtimeStat {
	systemStat := SystemRealtimeStat{
		Memory:  SystemMemoryStat{},
		Network: []*SystemNetworkStat{},
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

	ifMap := map[string]int{}

	ifs, _ := psuNet.Interfaces()

	for _, item := range ifs {
		networkStat := SystemNetworkStat{
			InterfaceStat: &SystemNetworkInterfaceStat{
				Type:        0,
				Description: "",
			},
		}
		if err := copier.Copy(networkStat.InterfaceStat, item); err != nil {
			log.Printf("network interface copy faild, %s\n", err)
		}

		systemStat.Network = append(systemStat.Network, &networkStat)
		length := len(systemStat.Network) - 1

		ifMap[strconv.FormatInt(int64(item.Index), 10)] = length
		ifMap[item.Name] = length
	}

	netIOs, _ := psuNet.IOCounters(true)

	for _, item := range netIOs {
		networkStat := systemStat.Network[ifMap[item.Name]]

		networkStat.IoStat = item
	}

	ais, _ := getAdaptersInfo()

	for _, item := range ais {
		networkStat := systemStat.Network[ifMap[strconv.FormatInt(int64(item.Index), 10)]]

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
	cpuStat.PerPercents, _ = psuCpu.Percent(0, true)
	avtPercents, _ := psuCpu.Percent(0, false)
	cpuStat.Percent = avtPercents[0]

	systemStat.Cpu[strconv.FormatInt(int64(cpuStat.InfoStat.CPU), 10)] = &cpuStat

	systemStat.Host.InfoStat, _ = psuHost.Info()

	return &systemStat
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
