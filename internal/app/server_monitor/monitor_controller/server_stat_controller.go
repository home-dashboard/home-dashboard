package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"net/http"
)

type DeviceType string

const (
	All            DeviceType = "all"
	Cpu                       = "cpu"
	Disk                      = "disk"
	NetworkAdapter            = "networkAdapter"
)

type DeviceInfoRequest struct {
	Type DeviceType `form:"type"`
}

func DeviceInfo(context *gin.Context) {
	var body DeviceInfoRequest

	if err := context.ShouldBindQuery(&body); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		panic(err)
	}

	switch body.Type {
	case Cpu:
		context.JSON(http.StatusOK, gin.H{Cpu: monitor_service.GetCpuInfo()})
	case Disk:
		context.JSON(http.StatusOK, gin.H{Disk: monitor_service.GetDiskInfo()})
		break
	case NetworkAdapter:
		context.JSON(http.StatusOK, gin.H{NetworkAdapter: monitor_service.GetNetworkAdapterInfo()})
		break
	case All:
	default:
		context.JSON(http.StatusOK, gin.H{
			NetworkAdapter: monitor_service.GetNetworkAdapterInfo(),
			Cpu:            monitor_service.GetCpuInfo(),
			Disk:           monitor_service.GetDiskInfo(),
		})
		break
	}
}
