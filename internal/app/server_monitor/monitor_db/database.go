package monitor_db

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/database"
)

var logger = comfy_log.New("[monitor_db]")

func init() {
	db := database.GetDB()

	if err := db.AutoMigrate(
		&monitor_model.StoredSystemStat{},
		&monitor_model.StoredSystemNetworkAdapterInfo{},
		&monitor_model.StoredSystemDiskInfo{},
		&monitor_model.StoredSystemCpuInfo{},
		&monitor_model.User{},
		&monitor_model.StoredConfiguration{},
	); err != nil {
		logger.Fatal("auto generate table failed, %s", err)
	}
}
