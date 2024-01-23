package database

import (
	"github.com/glebarez/sqlite"
	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"path"
)

var logger = comfy_log.New("[database]")

var db *gorm.DB
var dsn = "file::memory:?cache=shared"

func connectDataBase() error {
	isDev := configuration.Get().ServerMonitor.Development.Enable
	_db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(lo.Ternary(isDev, gormlogger.Info, gormlogger.Warn)),
	})
	if err != nil {
		return errors.New(err)
	}

	db = _db

	return nil
}

func openOrCreateDB() error {
	if err := connectDataBase(); err != nil {
		return errors.New(err)
	}

	return nil
}

func GetDB() *gorm.DB {
	if db == nil {
		if err := openOrCreateDB(); err != nil {
			logger.Fatal("open db failed, %w\n", err)
		}
	}

	return db
}

// SetSourceFilePath 设置数据库文件路径. 该路径是相对于工作目录的相对路径.
func SetSourceFilePath(dsnPath string) {
	dsn = path.Join(utils.WorkspaceDir(), dsnPath)
}
