package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"net/http"
)

// CollectShortcutSectionItemUsages 收集 monitor_model.ShortcutItem 使用情况
// @Summary CollectShortcutSectionItemUsages
// @Description CollectShortcutSectionItemUsages
// @Tags CollectShortcutSectionItemUsages
// @Produce json
// @Param usages body []ShortcutSectionItemUsage true "usages"
// @Success 200
// @Router shortcut/usage/collect [post]
func CollectShortcutSectionItemUsages(c *gin.Context) {
	var body struct {
		Usages []monitor_model.ShortcutSectionItemUsage `form:"usages" json:"usages" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		respondUnknownError(c, err.Error())
		return
	}

	if _, err := monitor_service.CreateOrUpdateShortcutSectionItemUsages(body.Usages); err != nil {
		respondUnknownError(c, err.Error())
		return
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}
