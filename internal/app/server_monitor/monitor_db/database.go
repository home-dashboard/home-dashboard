package monitor_db

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"gorm.io/gorm"
)

var logger = comfy_log.New("[monitor_db]")

var database *gorm.DB

func Initial(db *gorm.DB) error {
	database = db

	return database.AutoMigrate(
		&monitor_model.StoredSystemStat{},
		&monitor_model.StoredSystemNetworkAdapterInfo{},
		&monitor_model.StoredSystemDiskInfo{},
		&monitor_model.StoredSystemCpuInfo{},
		&monitor_model.User{},
		&monitor_model.StoredConfiguration{},
		&monitor_model.StoredNotification{},
		&monitor_model.ShortcutSection{},
		&monitor_model.ShortcutItem{},
		&monitor_model.ShortcutIcon{},
		&monitor_model.ShortcutSectionItemUsage{},
		&monitor_model.UserAgent{},
	)
}

func GetDB() *gorm.DB {
	if database == nil {
		logger.Panic("database is not initialized\n use monitor_db.Initial(db *gorm.DB) to initialize database\n")
	}

	return database
}
