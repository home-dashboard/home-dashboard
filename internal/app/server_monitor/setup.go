package server_monitor

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	ginSessions "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_controller"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/notification"
	"github.com/siaikin/home-dashboard/internal/pkg/authority"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"github.com/siaikin/home-dashboard/internal/pkg/sessions"
	"github.com/siaikin/home-dashboard/third_party"
	"github.com/siaikin/home-dashboard/web/web_submodules"
	"log"
	"net/http"
	"strconv"
	"time"
)

var logger = comfy_log.New("[server_monitor]")

func setupEngine(mock bool) *gin.Engine {
	r := gin.Default()
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Printf("server set proxy failed, %s\n", err)
		return nil
	}

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(sessions.GetSessionMiddleware())

	r.Use(func(c *gin.Context) {
		c.Next()

		errs := c.Errors
		errSize := len(errs)

		if errSize > 0 {
			err := errs[errSize-1]

			switch err.Err.(type) {
			case comfy_errors.ResponseError:
				resErr := err.Err.(comfy_errors.ResponseError)

				_ = err.SetMeta(map[string]interface{}{
					"code":    resErr.Code,
					"message": resErr.Error(),
				})
			default:
			}

			c.JSON(c.Writer.Status(), err)
		}
	})

	r.Use(func(c *gin.Context) {
		c.Next()

		session := ginSessions.Default(c)

		if err := session.Save(); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.SessionStoreError, "session save failed"))
			log.Printf("save session failed, %s\n", err)
		}
	})

	if mock {
		log.Println("server starting mock mode")

		r.Use(cors.New(cors.Config{
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "x-requested-with"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
			AllowOrigins:     configuration.Get().ServerMonitor.Development.Cors.AllowOrigins,
		}))

	}

	return r
}

func setupRouter(router *gin.RouterGroup, mock bool) {
	if mock {
		// mock 模式下自动添加 session
		router.Use(func(context *gin.Context) {
			session := ginSessions.Default(context)

			if session.Get(authority.InfoKey) == nil {
				context.Next()

				// 仅在未登录的第一次请求设置 Options (通过是否存在 authority.InfoKey 判断登录).
				// 之后的 session 内容修改将导致响应加上 Set-Cookie 标头, 触发浏览器重新设置 Cookie.
				// 但后续的 Set-Cookie 没有设置下方的 Options , 在跨域场景下无法正确设置 Cookie ([Cors 第三方 Cookie]).
				//
				// [Cors 第三方 Cookie]: https://developer.mozilla.org/zh-CN/docs/Web/HTTP/CORS#%E8%8B%A5%E5%B9%B2%E8%AE%BF%E9%97%AE%E6%8E%A7%E5%88%B6%E5%9C%BA%E6%99%AF
				//
				// todo [ginSessions.Options] 多次调用时支持合并参数. 避免多次调用时互相覆盖 Options 参数.
				// todo [ginSessions.Session] 支持获取是否修改状态. 设置 Options 参数会通知浏览器更新 Cookie. 未修改时不需要设置 Options 参数, 可以避免无意义的更新 Cookie 动作.
				session.Options(ginSessions.Options{
					Path:     "/",
					SameSite: http.SameSiteNoneMode,
					Secure:   true,
					// 24 hours
					MaxAge: 60 * 60 * 24,
				})
			} else {
				context.Next()
			}
		})

	}

	router.POST("auth", monitor_controller.Authorize)
	router.POST("unauth", monitor_controller.Unauthorize)

	authorized := router.Group("", authority.AuthorizeMiddleware())

	authorized.GET("notification", notification.Notification)
	authorized.POST("notification/collect", notification.ModifyCollectStat)
	authorized.GET("notification/collect", notification.GetCollectStat)
	authorized.GET("info/device", monitor_controller.DeviceInfo)

	// 获取配置的更新信息
	authorized.GET("configuration/updates", monitor_controller.GetChangedConfiguration)

	// 启用第三方服务
	if err := third_party.Use(authorized); err != nil {
		logger.Fatal("third party service start failed, %s.\n", err)
	}
}

var server *http.Server

func startServer(port uint, mock bool) {
	portStr := strconv.FormatInt(int64(port), 10)

	engine := setupEngine(mock)

	setupRouter(engine.Group("/v1/web"), mock)

	// 嵌入 home-dashboard-web-ui 静态资源
	if err := web_submodules.EmbedHomeDashboardWebUI(engine); err != nil {
		log.Fatalf("embed home-dashboard-web-ui failed, %s\n", err)
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
