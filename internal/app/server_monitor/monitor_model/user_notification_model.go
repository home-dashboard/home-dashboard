package monitor_model

import "github.com/siaikin/home-dashboard/internal/pkg/notification"

// StoredNotification 表示一个通知的存储结构.
type StoredNotification struct {
	Model
	notification.UserNotification
}
