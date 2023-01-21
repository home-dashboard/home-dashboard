package server_monitor

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_controller"
	"log"
	"net/http"
	"strconv"
	"time"
)

func setupEngine() *gin.Engine {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	if err := r.SetTrustedProxies(nil); err != nil {
		return nil
	}

	return r
}

func setupRouter(engine *gin.Engine) {
	authorized := engine.Group("/v1/web", gin.BasicAuth(gin.Accounts{
		"siaikin": "abc242244",
	}))

	authorized.GET("notification", monitor_controller.Notification)
	authorized.GET("info/device", monitor_controller.DeviceInfo)
}

func setupRouterMock(engine *gin.Engine) {
	engine.Use(cors.Default())

	mockRouter := engine.Group("/mock/v1/web")

	mockRouter.GET("notification", monitor_controller.Notification)
	mockRouter.GET("info/device", monitor_controller.DeviceInfo)
}

var server *http.Server

func startServer(port int, mock *bool) {
	portStr := strconv.FormatInt(int64(port), 10)

	engine := setupEngine()
	if *mock {
		log.Println("server starting mock mode")
		setupRouterMock(engine)
	} else {
		setupRouter(engine)
	}

	server = &http.Server{
		Addr:    ":" + portStr,
		Handler: engine,
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
