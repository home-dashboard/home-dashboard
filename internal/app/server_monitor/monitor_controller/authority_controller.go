package monitor_controller

import (
	"errors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
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

	if err := session.Save(); err != nil {
		_ = context.AbortWithError(http.StatusBadRequest, err)
		return
	}
}
