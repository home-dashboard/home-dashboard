package main

import (
	"flag"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	serverPort = flag.Int("port", -1, "serve port")
)

func init() {
	flag.Parse()

	if *serverPort == -1 {
		log.Printf("port %d is invalid", *serverPort)
		os.Exit(2)
	}
}

func main() {
	flag.Parse()

	server_monitor.Initial()
	server_monitor.Start(*serverPort)
	defer server_monitor.Stop()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("receive exit signal")
}
