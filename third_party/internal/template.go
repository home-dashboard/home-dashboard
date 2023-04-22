package internal

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
)

// ThirdPartyModule 是第三方服务接口, 所有第三方服务都必须实现该接口.
type ThirdPartyModule interface {
	// GetName 获取第三方服务名称.
	GetName() string
	// Load 加载第三方服务.
	Load(context *ThirdPartyModuleLoadContext) error
	// Unload 卸载第三方服务.
	Unload() error
	// DispatchEvent 将事件通知给第三方服务.
	DispatchEvent(event ThirdPartyEvent) error
}

type ThirdPartyModuleLoadContext struct {
	context.Context
	Router *gin.RouterGroup
	Logger *comfy_log.Logger
}

// ThirdPartyEvent 是第三方事件接口, 所有第三方事件都必须实现该接口.
type ThirdPartyEvent interface {
	GetType() string
	GetData() any
}
