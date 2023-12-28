package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"net/http"
)

// RefreshShortcutIcons 刷新所有快捷方式的图标
// @Summary RefreshShortcutIcons
// @Description RefreshShortcutIcons
// @Tags RefreshShortcutIcons
// @Produce json
// @Success 200
// @Router shortcut/icon/refresh [put]
func RefreshShortcutIcons(c *gin.Context) {
	if err := monitor_service.RefreshShortcutIcons(); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
