package internal

import "github.com/siaikin/home-dashboard/internal/pkg/notification"

const (
	UserNotificationKindInfo    = notification.UserNotificationKindInfo
	UserNotificationKindWarning = notification.UserNotificationKindWarning
	UserNotificationKindError   = notification.UserNotificationKindError
	UserNotificationKindSuccess = notification.UserNotificationKindSuccess

	UserNotificationOriginMain   = notification.UserNotificationOriginMain
	UserNotificationOriginWakapi = notification.UserNotificationOriginWakapi
	UserNotificationOriginGithub = notification.UserNotificationOriginGithub
)

type UserNotification = notification.UserNotification

// SendUserNotifications 代理 [notification.SendUserNotifications] 函数. 不同之处在于, 它将会给每个通知的 uniqueId 添加一个前缀, 以防止不同模块之间发生 uniqueId 冲突.
func SendUserNotifications(module ThirdPartyModule, notifications []*UserNotification) {
	_notifications := make([]UserNotification, len(notifications))
	for i, _notification := range notifications {
		_notification.UniqueId = module.GetName() + "-" + _notification.UniqueId
		_notifications[i] = *_notification
	}

	notification.SendUserNotifications(_notifications)
}

// SyncUserNotificationsUnreadState 代理 [notification.SyncUserNotificationsUnreadState] 函数. 不同之处在于, 它将会给每个通知的 uniqueId 添加一个前缀, 以防止不同模块之间发生 uniqueId 冲突.
// origin 是一个字符串, 用于标识通知的来源.
func SyncUserNotificationsUnreadState(module ThirdPartyModule, uniqueIds []string, origin string) {
	_uniqueIds := make([]string, len(uniqueIds))
	for i, uniqueId := range uniqueIds {
		_uniqueIds[i] = module.GetName() + "-" + uniqueId
	}

	notification.SyncUserNotificationsUnreadState(_uniqueIds, origin)
}
