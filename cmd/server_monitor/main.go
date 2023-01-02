package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor"
	"net/http"
	"os"
)

var (
	serverPort = flag.Int("port", -1, "serve port")
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"siaikin": "abc242244",
	}))

	authorized.GET("server-stat", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"stat": server_monitor.GetSystemStatic()})
	})

	return r
}

func init() {
	flag.Parse()
	if *serverPort == -1 {
		fmt.Printf("port %d is invalid", *serverPort)
		os.Exit(2)
	}
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
