package server_monitor_test

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor"
	"github.com/siaikin/home-dashboard/internal/pkg/database"
	"testing"
)

func TestInitial(t *testing.T) {
	// 使用默认的内存数据库
	db := database.GetDB()
	server_monitor.Initial(db)
}
