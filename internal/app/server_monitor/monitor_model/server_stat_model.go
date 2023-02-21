package monitor_model

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_realtime"
	"github.com/siaikin/home-dashboard/internal/pkg/database_types"
)

type StoredSystemStat struct {
	Model
	AdapterStat database_types.JSON
	CpuStat     database_types.JSON
	DiskStat    database_types.JSON
	MemoryStat  int8
}

type StoredSystemNetworkAdapterInfo struct {
	Model
	ID uint `gorm:"autoIncrement"`
	monitor_realtime.SystemNetworkAdapterInfo
}

type StoredSystemCpuInfo struct {
	Model
	CPU       int32   `json:"cpu"`
	Family    string  `json:"family"`
	Cores     int32   `json:"cores"`
	ModelName string  `json:"modelName"`
	Mhz       float64 `json:"mhz"`
}

type StoredSystemDiskInfo struct {
	Model
	Device     string `json:"device"`
	Mountpoint string `json:"mountpoint"`
	Fstype     string `json:"fstype"`
}
