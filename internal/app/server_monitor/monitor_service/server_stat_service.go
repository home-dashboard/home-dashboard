package monitor_service

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_realtime"
)

func GetSystemRealtimeStat() *monitor_realtime.SystemRealtimeStat {
	return monitor_realtime.GetCachedSystemRealtimeStat()
}

func GetNetworkAdapterInfo() *[]monitor_model.StoredSystemNetworkAdapterInfo {
	db := monitor_db.GetDB()

	var adapterInfos []monitor_model.StoredSystemNetworkAdapterInfo

	db.Find(&adapterInfos)

	return &adapterInfos
}

func GetCpuInfo() *[]monitor_model.StoredSystemCpuInfo {
	db := monitor_db.GetDB()

	var cpuInfos []monitor_model.StoredSystemCpuInfo
	db.Find(&cpuInfos)

	return &cpuInfos
}

func GetDiskInfo() *[]monitor_model.StoredSystemDiskInfo {
	db := monitor_db.GetDB()

	var diskInfos []monitor_model.StoredSystemDiskInfo
	db.Find(&diskInfos)

	return &diskInfos
}
