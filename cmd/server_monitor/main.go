package main

import (
	"flag"
	"fmt"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor"
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

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	config = configuration.Get()
	if config.ServerMonitor.Port == 0 {
		log.Fatalf("port %d is invalid", config.ServerMonitor.Port)
	}

}

func main() {
	fmt.Printf("version %s, commit %s, built at %s config %s\n", version, commit, date, config)

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
	fmt.Println("receive exit signal")
}
