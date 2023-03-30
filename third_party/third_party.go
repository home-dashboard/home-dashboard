package third_party

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"github.com/siaikin/home-dashboard/third_party/wakapi"
)

// Use 启用第三方服务.
func Use(router *gin.RouterGroup) error {
	wakapiConfig := configuration.Get().ServerMonitor.ThirdParty.Wakapi

	if wakapiConfig.Enable {
		if err := wakapi.Use(router.Group("wakapi")); err != nil {
			return err
		}
	}

	return nil
}
