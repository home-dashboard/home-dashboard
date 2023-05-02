package internal

import "github.com/gin-gonic/gin"

const (
	// NotificationChannelConnectedEventType 通知通道已连接事件.
	NotificationChannelConnectedEventType = "NotificationChannelConnected"
	// UserNotificationMarkedAsReadEventType 通知被标记为已读事件.
	UserNotificationMarkedAsReadEventType = "NotificationMarkedAsRead"
)

// NotificationChannelConnectedEvent 通知通道已连接事件.
type NotificationChannelConnectedEvent struct {
	ThirdPartyEvent
}
type NotificationChannelConnectedEventData struct {
	Context *gin.Context
}

// UserNotificationMarkedAsReadEvent 通知被标记为已读事件
type UserNotificationMarkedAsReadEvent struct {
	ThirdPartyEvent
}
type UserNotificationMarkedAsReadEventData struct {
	Context      *gin.Context
	Notification UserNotification
}
