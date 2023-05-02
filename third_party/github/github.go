package github

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"github.com/siaikin/home-dashboard/third_party/internal"
)

// 启用 GitHub 服务.
func load(context context.Context, router *gin.RouterGroup) error {
	githubConfig := configuration.Get().ServerMonitor.ThirdParty.GitHub

	// 校验参数
	if !githubConfig.Enable {
		return fmt.Errorf("third party service are used, but not enabled in configuration file")
	} else if len(githubConfig.PersonalAccessToken) <= 0 {
		return fmt.Errorf("personal access token is empty")
	}

	logger.Info("load with apiKey: \"%s\"\n", utils.MaskToken(githubConfig.PersonalAccessToken))

	if err := httpClientInitial(); err != nil {
		return err
	}

	router.GET("user", GetUserInfo)
	router.PATCH("notification/reset", resetNotificationRequestHeader)

	startFetchNotificationLoop(context)
	startFetchUserInfoLoop(context)

	return nil
}

func unload() error {
	return nil
}

func dispatchEvent(event internal.ThirdPartyEvent) error {
	switch event.GetType() {
	case internal.NotificationChannelConnectedEventType:
		//sendNotifications(getCachedNotifications())
		return nil
	case internal.UserNotificationMarkedAsReadEventType:
		data := event.GetData().(internal.UserNotificationMarkedAsReadEventData)
		return markNotificationAsRead(data.Notification.UniqueId)
	default:
		return nil
	}
}
