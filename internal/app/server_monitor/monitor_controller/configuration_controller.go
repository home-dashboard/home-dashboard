package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"net/http"
)

type ConfigurationRequest struct {
}

func GetChangedConfiguration(c *gin.Context) {
	configs, err := monitor_service.LatestNConfiguration(2)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
		return
	}

	result := gin.H{
		"current":  nil,
		"previous": nil,
	}

	size := len(*configs)

	if size >= 1 {
		result["current"] = (*configs)[size-1]
	}

	if size >= 2 {
		result["previous"] = (*configs)[size-2]
	}

	c.JSON(http.StatusOK, result)
}
