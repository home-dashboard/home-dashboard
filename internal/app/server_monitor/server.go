package server_monitor

import (
	"context"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_controller"
	"log"
	"net/http"
	"strconv"
	"time"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	if err := r.SetTrustedProxies(nil); err != nil {
		return nil
	}

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"siaikin": "abc242244",
	}))

	authorized.GET("notification", monitor_controller.Notification)
	authorized.GET("info/device", monitor_controller.DeviceInfo)

	return r
}

var server *http.Server

func startServer(port int) {
	portStr := strconv.FormatInt(int64(port), 10)

	router := setupRouter()
	server = &http.Server{
		Addr:    ":" + portStr,
		Handler: router,
	}

	go func() {
		log.Printf("server listening and serving on port %s\n", portStr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
}

func stopServer(timeout time.Duration) {
	// 优雅地关闭进程, 5秒后将强制结束进程
	// ctx 用于通知服务器有5秒来结束当前正在处理的请求
	// https://github.com/gin-gonic/examples/blob/master/graceful-shutdown/graceful-shutdown/notify-without-context/server.go
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalln("server forced to shutdown: ", err)
	}

	log.Println("server graceful shutdown")
}
