package monitor_db

import (
	"github.com/glebarez/sqlite"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"gorm.io/gorm"
	"log"
	"path"
)

var db *gorm.DB

func connectDataBase() error {
	_db, err := gorm.Open(sqlite.Open(path.Join(utils.WorkspaceDir, "server_stat.db")), &gorm.Config{})
	if err != nil {
		log.Panicf("data base open failed, %s", err)
		return err
	}

	db = _db

	if err := db.AutoMigrate(
		&monitor_model.StoredSystemStat{},
		&monitor_model.StoredSystemNetworkAdapterInfo{},
		&monitor_model.StoredSystemDiskInfo{},
		&monitor_model.StoredSystemCpuInfo{},
		&monitor_model.User{},
	); err != nil {
		log.Panicf("data base initial failed, %s", err)
		return err
	}

	return nil
}

func openOrCreateDB() error {
	if err := connectDataBase(); err != nil {
		log.Printf("data base connecting failed, %s", err)
		return err
	}

	return nil
}

func GetDB() *gorm.DB {
	if db == nil {
		if err := openOrCreateDB(); err != nil {
			log.Fatalf("open db failed, %s", err)
		}
	}

	return db
}
