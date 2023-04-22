package authority

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"net/http"
)

var InfoKey = "authorityInfo"

// AuthorizeMiddleware 权限验证中间件
func AuthorizeMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		session := sessions.Default(context)

		info := session.Get(InfoKey)
		if info == nil {
			_ = context.AbortWithError(http.StatusUnauthorized, comfy_errors.NewResponseError(comfy_errors.LoginRequestError, "unauthorized request"))
		}

		context.Next()
	}
}
