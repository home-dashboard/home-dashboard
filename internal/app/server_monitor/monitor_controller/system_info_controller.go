package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
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
		_ = context.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
	}

	context.JSON(http.StatusOK, info)
}
