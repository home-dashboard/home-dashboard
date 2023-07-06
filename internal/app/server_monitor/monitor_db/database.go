package monitor_db

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"gorm.io/gorm"
)

var logger = comfy_log.New("[monitor_db]")

var db *gorm.DB

func Initial(_db *gorm.DB) {
	db = _db

	if err := db.AutoMigrate(
		&monitor_model.StoredSystemStat{},
		&monitor_model.StoredSystemNetworkAdapterInfo{},
		&monitor_model.StoredSystemDiskInfo{},
		&monitor_model.StoredSystemCpuInfo{},
		&monitor_model.User{},
		&monitor_model.StoredConfiguration{},
		&monitor_model.StoredNotification{},
	); err != nil {
		logger.Fatal("auto generate table failed, %s\n", err)
	}

}

func GetDB() *gorm.DB {
	if db == nil {
		logger.Panic("db is nil\n")
	}

	return db
}
