package authority

import (
	"errors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

var InfoKey = "authorityInfo"

// AuthorizeMiddleware 权限验证中间件
func AuthorizeMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		session := sessions.Default(context)

		info := session.Get(InfoKey)
		if info == nil {
			err := errors.New("unauthorized request")
			_ = context.AbortWithError(http.StatusUnauthorized, err)
		}

		context.Next()
	}
}
