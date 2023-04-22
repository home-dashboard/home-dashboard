package internal

import "github.com/gin-gonic/gin"

const (
	// NotificationChannelConnectedEventType 通知通道已连接事件.
	NotificationChannelConnectedEventType = "NotificationChannelConnected"
)

// NotificationChannelConnectedEvent 通知通道已连接事件.
type NotificationChannelConnectedEvent struct {
	ThirdPartyEvent
}
type NotificationChannelConnectedEventData struct {
	Context *gin.Context
}
