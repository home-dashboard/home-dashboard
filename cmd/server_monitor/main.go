package main

import (
	"fmt"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor"
	"github.com/siaikin/home-dashboard/internal/pkg/arguments"
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

func init() {
	if *arguments.ServerPort == -1 {
		fmt.Printf("port %d is invalid", *arguments.ServerPort)
		os.Exit(2)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Printf("version %s, commit %s, built at %s\n", version, commit, date)

	server_monitor.Initial()
	server_monitor.Start(*arguments.ServerPort)
	defer server_monitor.Stop()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("receive exit signal")
}
