package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
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
		respondUnknownError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
