package database

import (
	"github.com/glebarez/sqlite"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"gorm.io/gorm"
	"path"
)

var logger = comfy_log.New("[database]")

var db *gorm.DB
var dsn = "file::memory:?cache=shared"

func connectDataBase() error {
	_db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Panic("data base open failed, %s\n", err)
		return err
	}

	db = _db

	return nil
}

func openOrCreateDB() error {
	if err := connectDataBase(); err != nil {
		logger.Info("data base connecting failed, %s\n", err)
		return err
	}

	return nil
}

func GetDB() *gorm.DB {
	if db == nil {
		if err := openOrCreateDB(); err != nil {
			logger.Fatal("open db failed, %s\n", err)
		}
	}

	return db
}

// SetSourceFilePath 设置数据库文件路径, 默认为内存数据库.
func SetSourceFilePath(_path string) {
	dsn = path.Join(utils.WorkspaceDir(), _path)
}
