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

// Authorize 登录并创建 session
// @Summary 登录并创建 session
// @Description 登录并创建 session. 如果用户开启了 2FA, 则将 [authority.TotpValidatedKey] 设置为 false.
// @Tags 登录
// @Accept json
// @Produce json
// @Param username body string true "username"
// @Param password body string true "password"
// @Success 200 {string} string "OK"
// @Router /auth [post]
func Authorize(context *gin.Context) {
	var body AuthorizeRequest

	if err := context.ShouldBindJSON(&body); err != nil {
		respondEntityValidationError(context, err.Error())
		return
	}

	user, err := monitor_service.GetUserByName(body.Username)
	if err != nil || user.Password != body.Password {
		respondLoginError(context, "username or password invalid")
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
		abortWithError(context, http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.SessionStoreError, "session save failed. %w", err))
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"role":      user.Role,
		"username":  user.Username,
		"enable2FA": user.Enable2FA,
	})
}

// Unauthorize 登出并删除 session
// @Summary 登出并删除 session
// @Description 登出并删除 session
// @Tags 登出
// @Accept json
// @Produce json
// @Success 200 {string} string "OK"
// @Router /unauth [delete]
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
		abortWithError(context, http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.SessionStoreError, "session save failed"))
		logger.Info("save session failed, %s\n", err)
	}

}

// GetCurrentUser 获取当前用户信息
func GetCurrentUser(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(authority.InfoKey).(authority.User)

	if len(user.Username) <= 0 {
		respondUnknownError(c, "current user not found")
		return
	}

	if storedUser, err := monitor_service.GetUserByName(user.Username); err != nil {
		respondUnknownError(c, "cannot get current user. %w", err)
		return
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
		respondLoginError(c, "cannot get current user. %w", err)
		return
	}

	if storedUser.Enable2FA == false {
		return
	}

	storedUser.Enable2FA = false
	storedUser.Secret2FA = ""
	if err := monitor_service.UpdateUser(storedUser); err != nil {
		abortWithError(c, http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.UnknownError, "cannot update user. %w", err))
		return
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
		respondLoginError(c, err.Error())
		return
	}

	session := sessions.Default(c)
	user := session.Get(authority.InfoKey).(authority.User)

	storedUser, err := monitor_service.GetUserByName(user.Username)
	if err != nil {
		respondLoginError(c, "cannot get current user. %w", err)
		return
	}

	// 校验 2FA 验证码, 如果正确, 将 session 中的 [authority.TotpValidatedKey] 设置为 true.
	if authority.TOTP.Validate(body.Code, storedUser.Secret2FA) {
		session.Set(authority.TotpValidatedKey, true)
		saveSession(c, session)
	} else {
		respondLoginError(c, "2fa code invalid")
		return
	}
}

// Binding2FAByAuthenticatorApp 绑定用于双因素身份验证的身份验证器应用
func Binding2FAByAuthenticatorApp(c *gin.Context) {
	var body Validate2FARequest
	if err := c.ShouldBindJSON(&body); err != nil {
		respondLoginError(c, err.Error())
		return
	}

	session := sessions.Default(c)
	user := session.Get(authority.InfoKey).(authority.User)

	// 校验 2FA 验证码, 如果正确, 将 session 中的 [authority.TotpValidatedKey] 设置为 true.
	if authority.TOTP.Validate(body.Code, user.Secret2FA) {
		storedUser, err := monitor_service.GetUserByName(user.Username)
		if err != nil {
			respondLoginError(c, err.Error())
			return
		}

		// 更新 Secret2FA
		storedUser.Secret2FA = user.Secret2FA
		storedUser.Enable2FA = true
		if err := monitor_service.UpdateUser(storedUser); err != nil {
			abortWithError(c, http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.UnknownError, "cannot update user. %w", err))
			return
		}

		// 绑定成功后, 重新登录
		Unauthorize(c)
	} else {
		respondLoginError(c, "2fa code invalid")
		return
	}
}

// Generate2FABindingQRCode 生成绑定 2FA 的二维码
func Generate2FABindingQRCode(c *gin.Context) {
	session := sessions.Default(c)

	user := session.Get(authority.InfoKey).(authority.User)

	qrCode, secret, err := authority.TOTP.GenerateBindingQRCode(user.Username)
	if err != nil {
		respondUnknownError(c, "generate qr code failed")
		return
	}

	var buffer bytes.Buffer

	if err := png.Encode(&buffer, qrCode); err != nil {
		respondUnknownError(c, "generate qr code failed")
		return
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
		abortWithError(c, http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.SessionStoreError, "session save failed"))
		return
	}
}
