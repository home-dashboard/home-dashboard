package third_party

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"github.com/siaikin/home-dashboard/third_party/github"
	"github.com/siaikin/home-dashboard/third_party/internal"
	"github.com/siaikin/home-dashboard/third_party/wakapi"
)

var ctx context.Context
var cancel context.CancelFunc

var modules = map[string]internal.ThirdPartyModule{}

// Load 启用第三方服务.
func Load(router *gin.RouterGroup) error {
	wakapiConfig := configuration.Get().ServerMonitor.ThirdParty.Wakapi
	githubConfig := configuration.Get().ServerMonitor.ThirdParty.GitHub

	if wakapiConfig.Enable {
		if err := wakapi.Use(router.Group("wakapi")); err != nil {
			return err
		}
	}

	ctx, cancel = context.WithCancel(context.Background())

	if githubConfig.Enable {
		githubModule := github.Get()

		if err := githubModule.Load(&internal.ThirdPartyModuleLoadContext{
			Context: ctx,
			Router:  router.Group("github"),
			Logger:  comfy_log.New("[THIRD-PARTY " + githubModule.GetName() + "]"),
		}); err != nil {
			return err
		}

		modules[githubModule.GetName()] = githubModule
	}

	return nil
}

// Unload 停用第三方服务.
func Unload() error {
	// 通知所有第三方服务停止.
	cancel()

	// 卸载所有第三方服务.
	for s, module := range modules {
		if err := module.Unload(); err != nil {
			return err
		}
		delete(modules, s)
	}

	return nil
}

// DispatchEvent 将事件通知给第三方服务.
func DispatchEvent(event ThirdPartyEventImpl) error {
	for _, module := range modules {
		if err := module.DispatchEvent(event); err != nil {
			return err
		}
	}

	return nil
}
