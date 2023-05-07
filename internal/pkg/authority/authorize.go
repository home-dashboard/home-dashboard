package authority

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"net/http"
)

type User struct {
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	Enable2FA bool   `json:"enable2FA,omitempty"`
	// 2FA 密钥
	Secret2FA string `json:"secret2FA,omitempty"`
}

// InfoKey session 中存储用户信息的 key. 用户信息类型为 User
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

// Authorize2FAMiddleware 2FA 权限验证中间件. 当用户开启了 2FA 时, 该中间件会验证用户是否已经通过 2FA.
// 该中间件应该在 [AuthorizeMiddleware] 之后使用.
func Authorize2FAMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		session := sessions.Default(context)

		info := session.Get(InfoKey).(User)
		if info.Enable2FA {
			validated := session.Get(TotpValidatedKey)
			if validated == nil {
				_ = context.AbortWithError(http.StatusUnauthorized, comfy_errors.NewResponseError(comfy_errors.LoginRequestError, "unauthorized request"))
			} else if validated == false {
				_ = context.AbortWithError(http.StatusUnauthorized, comfy_errors.NewResponseError(comfy_errors.LoginRequestError, "2fa not validated"))
			}
		}

		context.Next()
	}
}
