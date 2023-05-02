package notification

const (
	// UserNotificationReceivedMessageType 用户通知的类型.
	UserNotificationReceivedMessageType = "userNotificationReceived"
	// UserNotificationReadMessageType 用户通知已读的类型.
	UserNotificationReadMessageType = "userNotificationRead"
)

// UserNotificationKind 表示通知的类型, 一般也表示通知的重要程度.
type UserNotificationKind = string

const (
	UserNotificationKindError   UserNotificationKind = "error"
	UserNotificationKindInfo    UserNotificationKind = "info"
	UserNotificationKindWarning UserNotificationKind = "warning"
	UserNotificationKindSuccess UserNotificationKind = "success"
)

// UserNotificationOrigin 表示通知的来源.
type UserNotificationOrigin = string

const (
	// UserNotificationOriginMain 表示通知来自于主程序.
	UserNotificationOriginMain UserNotificationOrigin = "main"
	// UserNotificationOriginWakapi 表示通知来自于第三方 Wakapi 服务.
	UserNotificationOriginWakapi UserNotificationOrigin = "wakapi"
	// UserNotificationOriginGithub 表示通知来自于第三方 Github 服务.
	UserNotificationOriginGithub UserNotificationOrigin = "github"
)

type UserNotification struct {
	// UniqueId 通知的唯一标识. 用于去重.
	UniqueId string `json:"uniqueId"`
	// Unread 通知是否未读.
	Unread bool `json:"unread"`
	// Title 通知的标题.
	Title string `json:"title"`
	// Caption 通知的描述信息.
	Caption string `json:"caption"`
	// Link 通知的跳转链接. 为空则不显示跳转按钮. 通常用于跳转到外部页面.
	Link string `json:"link"`
	// UserNotificationKind 通知的重要程度.
	Kind UserNotificationKind `json:"kind"`
	// Origin 通知的来源. 用于区分不同的通知来源.
	Origin         UserNotificationOrigin `json:"origin"`
	OriginCreateAt int64                  `json:"originCreateAt"`
}

// SendUserNotifications 发送用户通知.
func SendUserNotifications(notifications []UserNotification) {
	relay.Notify(Message{Type: UserNotificationReceivedMessageType, Data: map[string]any{"notifications": notifications}})
}

// SyncUserNotificationsUnreadState 将 origin 下除了 unreadUniqueIds 中列出的用户通知设置为已读. unreadUniqueIds 是所有未读通知的 uniqueId 列表, origin 是通知的来源.
func SyncUserNotificationsUnreadState(unreadUniqueIds []string, origin UserNotificationOrigin) {
	relay.Notify(Message{Type: UserNotificationReadMessageType, Data: map[string]any{"unreadUniqueIds": unreadUniqueIds, "origin": origin}})
}
