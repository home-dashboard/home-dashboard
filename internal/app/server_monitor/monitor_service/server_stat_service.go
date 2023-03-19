package monitor_service

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/pkg/database"
)

func GetNetworkAdapterInfo() *[]monitor_model.StoredSystemNetworkAdapterInfo {
	db := database.GetDB()

	var adapterInfos []monitor_model.StoredSystemNetworkAdapterInfo

	db.Find(&adapterInfos)

	return &adapterInfos
}

func GetCpuInfo() *[]monitor_model.StoredSystemCpuInfo {
	db := database.GetDB()

	var cpuInfos []monitor_model.StoredSystemCpuInfo
	db.Find(&cpuInfos)

	return &cpuInfos
}

func GetDiskInfo() *[]monitor_model.StoredSystemDiskInfo {
	db := database.GetDB()

	var diskInfos []monitor_model.StoredSystemDiskInfo
	db.Find(&diskInfos)

	return &diskInfos
}
