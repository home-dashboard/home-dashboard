package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"github.com/siaikin/home-dashboard/third_party"
	"net/http"
	"strconv"
)

type ListNotificationsRequest struct {
	// Max 最大条数, 0 表示不限制.
	Max int64 `form:"max"`
}

// ListUnreadNotifications 列出指定条数的未读通知
func ListUnreadNotifications(c *gin.Context) {
	var body ListNotificationsRequest

	if err := c.ShouldBindQuery(&body); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
		return
	}

	notifications, err := monitor_service.LatestNNotification(body.Max, true)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
		return
	}

	result := gin.H{
		"notifications": notifications,
	}

	c.JSON(http.StatusOK, result)
}

// MarkNotificationAsRead 标记一条通知为已读, 通知的 ID 通过 URL 参数传入.
func MarkNotificationAsRead(c *gin.Context) {
	if id, err := strconv.ParseUint(c.Param("id"), 10, 0); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
	} else {
		if notification, err := monitor_service.GetNotification(uint(id)); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
		} else if err := third_party.DispatchEvent(third_party.NewUserNotificationMarkedAsReadEvent(c, notification.UserNotification)); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
		} else if err := monitor_service.MarkNotificationAsRead(uint(id)); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
		}
	}
}

// MarkAllNotificationAsRead 标记所有未读通知为已读
func MarkAllNotificationAsRead(c *gin.Context) {
	if err := monitor_service.MarkAllNotificationAsRead(); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
	}
}
