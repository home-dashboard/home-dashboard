package monitor_realtime

import (
	psuCpu "github.com/shirou/gopsutil/v3/cpu"
	psuDisk "github.com/shirou/gopsutil/v3/disk"
	psuHost "github.com/shirou/gopsutil/v3/host"
	psuMem "github.com/shirou/gopsutil/v3/mem"
	psuNet "github.com/shirou/gopsutil/v3/net"
)

type SystemRealtimeStat struct {
	Memory    SystemMemoryStat          `json:"memory"`
	Network   []*SystemNetworkStat      `json:"network"`
	Disk      []*SystemDiskStat         `json:"disk"`
	Cpu       map[string]*SystemCpuStat `json:"cpu"`
	Host      SystemHostStat            `json:"host"`
	Timestamp int64                     `json:"timestamp"`
}
type SystemNetworkStat struct {
	InterfaceStat *SystemNetworkInterfaceStat `json:"interfaceStat"`
	IoStat        psuNet.IOCountersStat       `json:"ioStat"`
}
type SystemNetworkInterfaceStat struct {
	psuNet.InterfaceStat
	Type        int    `json:"type"`
	Description string `json:"description"`
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
type SystemCpuStat struct {
	InfoStat       psuCpu.InfoStat `json:"infoStat"`
	PhysicalCounts int             `json:"physicalCounts"`
	LogicalCounts  int             `json:"logicalCounts"`
	Percent        float64         `json:"percent"`
	PerPercents    []float64       `json:"PerPercents"`
}

type SystemHostStat struct {
	InfoStat *psuHost.InfoStat
}

type SystemNetworkAdapterInfo struct {
	Index       uint32 `json:"index"`
	Description string `json:"description"`
	Type        int    `json:"type"`
}
type SystemNetworkAdapterInfoList []SystemNetworkAdapterInfo
