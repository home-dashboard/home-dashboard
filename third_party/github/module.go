package github

import (
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/third_party/internal"
)

var logger *comfy_log.Logger

type module struct {
}

func (m module) GetName() string {
	return "GitHub"
}
func (m module) Load(context *internal.ThirdPartyModuleLoadContext) error {
	logger = context.Logger
	return load(context, context.Router)
}
func (m module) Unload() error {
	return unload()
}
func (m module) DispatchEvent(event internal.ThirdPartyEvent) error {
	return dispatchEvent(event)
}

var inst *module

// Get 获取 GitHub 模块实例. 重复调用返回同一个实例.
func Get() internal.ThirdPartyModule {
	if inst != nil {
		return inst
	}

	inst = &module{}

	return inst
}
