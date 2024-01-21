package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/host"
	"net/http"
)

// SystemInfo
// @Summary 获取系统信息
// @Description 获取系统信息
// @Tags SystemInfo
// @Accept json
// @Produce json
// @Success 200 {object} host.InfoStat
// @Router /system/info [get]
func SystemInfo(context *gin.Context) {
	info, err := host.Info()
	if err != nil {
		respondUnknownError(context, err.Error())
		return
	}

	context.JSON(http.StatusOK, info)
}
