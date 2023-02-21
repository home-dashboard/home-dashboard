package server_monitor

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	ginSessions "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_controller"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/pkg/authority"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"github.com/siaikin/home-dashboard/internal/pkg/sessions"
	"log"
	"net/http"
	"strconv"
	"time"
)

func setupEngine() *gin.Engine {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(sessions.GetSessionMiddleware())

	if err := r.SetTrustedProxies(nil); err != nil {
		log.Printf("server set proxy failed, %s\n", err)
		return nil
	}

	r.Use(func(c *gin.Context) {
		c.Next()

		errs := c.Errors
		errSize := len(errs)
		if errSize > 0 {
			c.JSON(c.Writer.Status(), gin.H{"error": errs[errSize-1]})
		}
	})

	return r
}

func setupRouter(router *gin.RouterGroup) {
	router.POST("auth", monitor_controller.Authorize)

	authorized := router.Group("", authority.AuthorizeMiddleware())

	authorized.GET("notification", monitor_controller.Notification)
	authorized.GET("info/device", monitor_controller.DeviceInfo)
}

var server *http.Server

func startServer(port uint, mock bool) {
	portStr := strconv.FormatInt(int64(port), 10)

	engine := setupEngine()
	if mock {
		log.Println("server starting mock mode")
		engine.Use(cors.Default())

		// mock 模式下自动添加 session
		engine.Use(func(context *gin.Context) {
			session := ginSessions.Default(context)

			if session.Get(authority.InfoKey) == nil {
				log.Println("auto generate mock session")
				session.Set(authority.InfoKey, monitor_model.User{
					Username: configuration.Config.ServerMonitor.Administrator.Username,
					Password: configuration.Config.ServerMonitor.Administrator.Username,
					Role:     monitor_model.RoleAdministrator,
				})

				if err := session.Save(); err != nil {
					log.Fatalf("save mock session failed, %s\n", err)
				}
			}

			context.Next()
		})
	}

	setupRouter(engine.Group("/v1/web"))

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
