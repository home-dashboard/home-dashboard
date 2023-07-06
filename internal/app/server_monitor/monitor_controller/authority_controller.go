package monitor_controller

import (
	"bytes"
	"github.com/gin-contrib/sessions"
	ginSessions "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"github.com/siaikin/home-dashboard/internal/pkg/authority"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"image/png"
	"net/http"
)

var logger = comfy_log.New("[server_monitor]")

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
		_ = context.AbortWithError(http.StatusUnauthorized, comfy_errors.NewResponseError(comfy_errors.LoginRequestError, "username or password invalid"))
		return
	}

	session := sessions.Default(context)

	// 如果用户开启了 2FA, 则将 [authority.TotpValidatedKey] 设置为 false, 以供中间件校验拦截.
	if user.Enable2FA {
		session.Set(authority.TotpValidatedKey, false)
	}

	session.Set(authority.InfoKey, user.User)
	session.Options(ginSessions.Options{
		// 24 hours
		MaxAge: 60 * 60 * 24,
	})

	if err := session.Save(); err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.SessionStoreError, "session save failed"))
		logger.Info("save session failed, %s\n", err)
	}

	context.JSON(http.StatusOK, gin.H{
		"role":      user.Role,
		"username":  user.Username,
		"enable2FA": user.Enable2FA,
	})
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
		_ = context.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.SessionStoreError, "session save failed"))
		logger.Info("save session failed, %s\n", err)
	}

}

// GetCurrentUser 获取当前用户信息
func GetCurrentUser(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(authority.InfoKey).(authority.User)

	if len(user.Username) <= 0 {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, "current user not found"))
	}

	if storedUser, err := monitor_service.GetUserByName(user.Username); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.UnknownError, "cannot get current user. %w", err))
	} else {
		// 移除敏感信息
		storedUser.Password = ""
		storedUser.Secret2FA = ""
		c.JSON(http.StatusOK, storedUser)
	}

}

// Disable2FA 禁用 2FA
func Disable2FA(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(authority.InfoKey).(authority.User)

	storedUser, err := monitor_service.GetUserByName(user.Username)
	if err != nil {
		_ = c.AbortWithError(http.StatusUnauthorized, err)
	}

	if storedUser.Enable2FA == false {
		return
	}

	storedUser.Enable2FA = false
	storedUser.Secret2FA = ""
	if err := monitor_service.UpdateUser(*storedUser); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.UnknownError, "cannot update user. %w", err))
	}

	// 禁用 2FA 后, 需要重新登录
	Unauthorize(c)
}

type Validate2FARequest struct {
	Code string `form:"code"`
}

// Validate2FA 校验 2FA 验证码
func Validate2FA(c *gin.Context) {
	var body Validate2FARequest
	if err := c.ShouldBindJSON(&body); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	session := sessions.Default(c)
	user := session.Get(authority.InfoKey).(authority.User)

	storedUser, err := monitor_service.GetUserByName(user.Username)
	if err != nil {
		_ = c.AbortWithError(http.StatusUnauthorized, err)
	}

	// 校验 2FA 验证码, 如果正确, 将 session 中的 [authority.TotpValidatedKey] 设置为 true.
	if authority.TOTP.Validate(body.Code, storedUser.Secret2FA) {
		session.Set(authority.TotpValidatedKey, true)
		saveSession(c, session)
	} else {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.LoginRequestError, "2fa code invalid"))
	}
}

// Binding2FAByAuthenticatorApp 绑定用于双因素身份验证的身份验证器应用
func Binding2FAByAuthenticatorApp(c *gin.Context) {
	var body Validate2FARequest
	if err := c.ShouldBindJSON(&body); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	session := sessions.Default(c)
	user := session.Get(authority.InfoKey).(authority.User)

	// 校验 2FA 验证码, 如果正确, 将 session 中的 [authority.TotpValidatedKey] 设置为 true.
	if authority.TOTP.Validate(body.Code, user.Secret2FA) {
		storedUser, err := monitor_service.GetUserByName(user.Username)
		if err != nil {
			_ = c.AbortWithError(http.StatusUnauthorized, err)
		}

		// 更新 Secret2FA
		storedUser.Secret2FA = user.Secret2FA
		storedUser.Enable2FA = true
		if err := monitor_service.UpdateUser(*storedUser); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.UnknownError, "cannot update user. %w", err))
		}

		// 绑定成功后, 重新登录
		Unauthorize(c)
	} else {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.LoginRequestError, "2fa code invalid"))
	}
}

// Generate2FABindingQRCode 生成绑定 2FA 的二维码
func Generate2FABindingQRCode(c *gin.Context) {
	session := sessions.Default(c)

	user := session.Get(authority.InfoKey).(authority.User)

	qrCode, secret, err := authority.TOTP.GenerateBindingQRCode(user.Username)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.UnknownError, "generate qr code failed"))
	}

	var buffer bytes.Buffer

	if err := png.Encode(&buffer, qrCode); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.UnknownError, "generate qr code failed"))
	} else {
		// 临时保存 secret, 用于校验 2FA 验证码
		user.Secret2FA = secret
		session.Set(authority.InfoKey, user)
		saveSession(c, session)

		c.Data(http.StatusOK, "image/png", buffer.Bytes())
	}
}

func saveSession(c *gin.Context, session sessions.Session) {
	if err := session.Save(); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.SessionStoreError, "session save failed"))
		logger.Info("save session failed, %s\n", err)
	}
}
