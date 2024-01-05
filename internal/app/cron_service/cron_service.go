package cron_service

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/controller"
)

type CornServiceContext struct {
	context.Context
	Router *gin.RouterGroup
}

func Load(router *gin.RouterGroup) error {
	router.GET("/nodejs/version/list_proxy", controller.ListNodejsVersion)
	router.GET("/nodejs/install/:version", controller.InstallNodejs)

	return nil
}
