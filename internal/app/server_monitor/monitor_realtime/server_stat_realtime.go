package monitor_realtime

import (
	"github.com/jinzhu/copier"
	psuCpu "github.com/shirou/gopsutil/v3/cpu"
	psuDisk "github.com/shirou/gopsutil/v3/disk"
	psuHost "github.com/shirou/gopsutil/v3/host"
	psuMem "github.com/shirou/gopsutil/v3/mem"
	psuNet "github.com/shirou/gopsutil/v3/net"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/notification"
	"golang.org/x/net/context"
	"strconv"
	"time"
)

var logger = comfy_log.New("[monitor_realtime]")

func getSystemRealtimeStatic() *SystemRealtimeStat {
	systemStat := SystemRealtimeStat{
		Memory:    SystemMemoryStat{},
		Network:   []*SystemNetworkStat{},
		Disk:      []*SystemDiskStat{},
		Cpu:       map[string]*SystemCpuStat{},
		Host:      SystemHostStat{},
		Timestamp: time.Now().UnixMilli(),
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
			logger.Info("network interface copy failed, %s\n", err)
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

	diskIndexMap := map[string]int{}

	parts, _ := psuDisk.Partitions(true)

	for _, item := range parts {
		diskStat := SystemDiskStat{
			PartitionStat: item,
		}

		systemStat.Disk = append(systemStat.Disk, &diskStat)
		diskIndexMap[item.Device] = len(systemStat.Disk) - 1

		usage, _ := psuDisk.Usage(item.Device)
		diskStat.UsageStat = usage
	}

	diskIOs, _ := psuDisk.IOCounters()

	for _, item := range diskIOs {
		diskStat := systemStat.Disk[diskIndexMap[item.Name]]
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

var currentSystemStat = getSystemRealtimeStatic()

const (
	MessageType = "realtimeStat"
)

func Loop(context context.Context, d time.Duration) {
	ticker := time.NewTicker(d)

	go func() {
		for {
			select {
			case <-context.Done():
				ticker.Stop()
				logger.Info("system realtime stat loop stopped\n")
				return
			case <-ticker.C:
				logger.Info("system realtime stat collecting\n")
				currentSystemStat = getSystemRealtimeStatic()
				logger.Info("system realtime stat collected\n")

				logger.Info("system realtime stat sending\n")
				notification.Send(MessageType, map[string]any{MessageType: currentSystemStat})
				logger.Info("system realtime stat sent\n")
			}
		}
	}()
}

func GetCachedSystemRealtimeStat() *SystemRealtimeStat {
	return currentSystemStat
}

func GetCpuPercent() float64 {
	percent := float64(0)

	for _, item := range currentSystemStat.Cpu {
		percent += item.Percent
	}

	return percent
}

func GetMemoryPercent() float64 {
	return currentSystemStat.Memory.VirtualMemory.UsedPercent
}
