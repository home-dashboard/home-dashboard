package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"github.com/siaikin/home-dashboard/internal/pkg/database"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var config *configuration.Configuration

var logger = comfy_log.New("[server_monitor]")

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	config = configuration.Get()
	if config.ServerMonitor.Port == 0 {
		logger.Fatal("port %d is invalid\n", config.ServerMonitor.Port)
	}
	if config.ServerMonitor.Development.Enable == false {
		gin.SetMode(gin.ReleaseMode)
	}
}

func main() {
	logger.Info("version %s, commit %s, built at %s config %s\n", version, commit, date, config)

	// 设置数据库文件路径
	database.SetSourceFilePath("home-dashboard.db")
	db := database.GetDB()

	server_monitor.Initial(db)
	server_monitor.Start(config.ServerMonitor.Port)
	defer server_monitor.Stop()

	// 收到中断信号, 程序退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("receive exit signal\n")
}
