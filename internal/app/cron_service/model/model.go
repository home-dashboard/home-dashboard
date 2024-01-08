package model

import (
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/database"
)

var logger = comfy_log.New("[cron_service model]")

func MigrateModel() error {
	db := database.GetDB()

	if err := db.AutoMigrate(
		&Project{},
	); err != nil {
		logger.Fatal("auto generate table failed, %s\n", err)
	}

	return nil
}
