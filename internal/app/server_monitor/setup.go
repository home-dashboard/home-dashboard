package server_monitor

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	ginSessions "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/file_service"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_controller"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"github.com/siaikin/home-dashboard/internal/pkg/authority"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"github.com/siaikin/home-dashboard/internal/pkg/sessions"
	"github.com/siaikin/home-dashboard/third_party"
	"github.com/siaikin/home-dashboard/web/web_submodules"
	"net"
	"net/http"
	"time"
)

func setupEngine(mock bool) *gin.Engine {
	r := gin.Default()
	//if err := r.SetTrustedProxies(nil); err != nil {
	//	logger.Info("server set proxy failed, %s\n", err)
	//	return nil
	//}

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(sessions.GetSessionMiddleware())

	r.Use(func(c *gin.Context) {
		c.Next()

		errs := c.Errors
		errSize := len(errs)

		if errSize > 0 {
			err := errs[errSize-1]

			var responseError comfy_errors.ResponseError
			if errors.As(err.Err, &responseError) {
				logger.Error("\n%s\n", responseError.ErrorStack())

				_ = err.SetMeta(map[string]interface{}{
					"code":    responseError.Code,
					"message": responseError.Err.Error(),
				})
			}

			c.JSON(c.Writer.Status(), err)
		}
	})

	//r.Use(func(c *gin.Context) {
	//	c.Next()
	//
	//	session := ginSessions.Default(c)
	//
	//	if err := session.Save(); err != nil {
	//		_ = c.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.SessionStoreError, "session save failed"))
	//		logger.Info("save session failed, %s\n", err)
	//	}
	//})

	if mock {
		logger.Warn("server starting development mode\n")

		allowMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
		allowHeaders := []string{"Origin", "Content-Length", "Content-Type", "x-requested-with"}
		allowCredentials := true
		maxAge := time.Hour * 12
		allowOrigins := configuration.Get().ServerMonitor.Development.Cors.AllowOrigins
		r.Use(cors.New(cors.Config{
			AllowMethods:     allowMethods,
			AllowHeaders:     allowHeaders,
			AllowCredentials: allowCredentials,
			MaxAge:           maxAge,
			AllowOrigins:     allowOrigins,
		}))
		logger.Warn(`
			allow origins: %v
			allow methods: %v
			allow headers: %v
			allow credentials: %v
			max age: %v
`, allowOrigins, allowMethods, allowHeaders, allowCredentials, maxAge)
	}

	return r
}

func setupRouter(router *gin.RouterGroup, mock bool) {
	if mock {
		// mock 模式下自动添加 session
		router.Use(func(context *gin.Context) {
			session := ginSessions.Default(context)

			if session.Get(authority.InfoKey) == nil {
				user, err := monitor_service.GetUserByName(configuration.Get().ServerMonitor.Administrator.Username)
				if err != nil {
					_ = context.AbortWithError(http.StatusUnauthorized, comfy_errors.NewResponseError(comfy_errors.LoginRequestError, "auto login failed, %w", err))
					return
				}

				session.Set(authority.InfoKey, user.User)

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
	// 获取当前登录用户信息
	router.GET("user/current", monitor_controller.GetCurrentUser)

	// 该路由组下的接口需要登录
	authorized := router.Group("", authority.AuthorizeMiddleware())
	// 2FA 校验相关接口
	authorized.GET("auth/2fa/qrcode", monitor_controller.Generate2FABindingQRCode)
	authorized.POST("auth/2fa/bind/app", monitor_controller.Binding2FAByAuthenticatorApp)
	authorized.POST("auth/2fa/validate", monitor_controller.Validate2FA)
	authorized.POST("auth/2fa/disable", monitor_controller.Disable2FA)

	// 该路由组下的接口需要登录且 2FA 校验通过(如果 2FA 开启)
	authorizedAnd2faValidated := authorized.Group("", authority.Authorize2FAMiddleware())

	authorizedAnd2faValidated.GET("notification", monitor_controller.Notification)
	authorizedAnd2faValidated.POST("notification/collect", monitor_controller.ModifyCollectStat)
	authorizedAnd2faValidated.GET("notification/collect", monitor_controller.GetCollectStat)
	authorizedAnd2faValidated.GET("info/device", monitor_controller.DeviceInfo)

	// 通知消息相关的接口
	authorizedAnd2faValidated.GET("notification/list/unread", monitor_controller.ListUnreadNotifications)
	authorizedAnd2faValidated.PATCH("notification/read/:id", monitor_controller.MarkNotificationAsRead)
	authorizedAnd2faValidated.PATCH("notification/read/all", monitor_controller.MarkAllNotificationAsRead)

	// 获取配置的更新信息
	authorizedAnd2faValidated.GET("configuration/updates", monitor_controller.GetChangedConfiguration)

	// 服务升级接口
	authorizedAnd2faValidated.POST("upgrade", monitor_controller.Upgrade)

	// 获取系统信息
	authorizedAnd2faValidated.GET("system/info", monitor_controller.SystemInfo)

	// 获取版本信息
	authorizedAnd2faValidated.GET("version", monitor_controller.Version)

	// 书签相关接口
	// -> 书签文件夹接口
	authorizedAnd2faValidated.POST("shortcut/section/create", monitor_controller.CreateShortcutSection)
	authorizedAnd2faValidated.GET("shortcut/section/list", monitor_controller.ListShortcutSections)
	authorizedAnd2faValidated.PUT("shortcut/section/update/:id", monitor_controller.UpdateShortcutSection)
	authorizedAnd2faValidated.DELETE("shortcut/section/delete/:id", monitor_controller.DeleteShortcutSection)
	authorizedAnd2faValidated.DELETE("shortcut/section/delete/:id/items", monitor_controller.DeleteShortcutSectionItems)
	// -> 书签接口
	authorizedAnd2faValidated.GET("shortcut/item/extract-from-url", monitor_controller.ExtractShortcutItemInfoFromURL)
	authorizedAnd2faValidated.POST("shortcut/item/create", monitor_controller.CreateShortcutItem)
	authorizedAnd2faValidated.GET("shortcut/item/list", monitor_controller.ListShortcutItems)
	authorizedAnd2faValidated.PUT("shortcut/item/update/:id", monitor_controller.UpdateShortcutItem)
	authorizedAnd2faValidated.DELETE("shortcut/item/delete", monitor_controller.DeleteShortcutItem)
	authorizedAnd2faValidated.PUT("shortcut/item/refresh-image-icon-cache/:sectionId", monitor_controller.RefreshCachedShortcutItemImageIcon)
	// -> 书签图标接口
	authorizedAnd2faValidated.PUT("shortcut/icon/refresh", monitor_controller.RefreshShortcutIcons)
	// -> 收集书签使用情况
	authorizedAnd2faValidated.POST("shortcut/usage/collect", monitor_controller.CollectShortcutSectionItemUsages)

	// 启用第三方服务
	if err := third_party.Load(authorizedAnd2faValidated); err != nil {
		logger.Fatal("third party service start failed, %s.\n", err)
	}
}

var server *http.Server

func startServer(listener *net.Listener, mock bool) error {
	engine := setupEngine(mock)

	setupRouter(engine.Group("/v1/web"), mock)

	// 启用文件服务
	if err := file_service.Serve(engine.Group("/v1/file")); err != nil {
		return errors.Errorf("file service start failed, %w\n", err)
	}

	// 嵌入 home-dashboard-web-ui 静态资源
	if err := web_submodules.EmbedHomeDashboardWebUI(engine); err != nil {
		return errors.Errorf("embed home-dashboard-web-ui failed, %w\n", err)
	}

	server = &http.Server{
		Handler: engine,
	}
	return server.Serve(*listener)
}

func stopServer(ctx context.Context) error {
	if err := server.Shutdown(ctx); err != nil {
		return errors.Errorf("server shutdown failed, %w\n", err)
	}

	if err := third_party.Unload(); err != nil {
		return errors.Errorf("third party service stop failed, %w\n", err)
	}

	return nil
}
