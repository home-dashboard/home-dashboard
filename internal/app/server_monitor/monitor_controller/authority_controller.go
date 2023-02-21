package monitor_controller

import (
	"errors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/notification"
	"github.com/siaikin/home-dashboard/internal/pkg/authority"
	"net/http"
)

type AuthorizeRequest struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

func Authorize(context *gin.Context) {
	var body AuthorizeRequest

	if err := context.ShouldBindJSON(&body); err != nil {
		_ = context.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := monitor_service.GetUserByName(body.Username)
	if err != nil || user.Password != body.Password {
		err := errors.New("username or password invalid")
		_ = context.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	session := sessions.Default(context)

	session.Set(authority.InfoKey, user)
	session.Set(notification.CollectStatConfigSessionKey, notification.DefaultCollectStatConfig())

	if err := session.Save(); err != nil {
		_ = context.AbortWithError(http.StatusBadRequest, err)
		return
	}
}

// Unauthorize 登出并删除 session
func Unauthorize(context *gin.Context) {
	session := sessions.Default(context)

	if session.Get(authority.InfoKey) == nil {
		return
	}

	// this will mark the session as "written" and hopefully remove the username
	session.Set(authority.InfoKey, nil)
	// 设置 MaxAge 为 -1 将删除 session
	session.Options(sessions.Options{MaxAge: -1})

	if err := session.Save(); err != nil {
		_ = context.AbortWithError(http.StatusBadRequest, err)
		return
	}
}
