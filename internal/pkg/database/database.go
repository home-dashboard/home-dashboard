package database

import (
	"github.com/glebarez/sqlite"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"gorm.io/gorm"
	"log"
	"path"
)

var db *gorm.DB
var dsn = "file::memory:?cache=shared"

func connectDataBase() error {
	_db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panicf("data base open failed, %s", err)
		return err
	}

	db = _db

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

// SetSourceFilePath 设置数据库文件路径, 默认为内存数据库.
func SetSourceFilePath(_path string) {
	dsn = path.Join(utils.WorkspaceDir, _path)
}
