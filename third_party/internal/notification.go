package internal

import "github.com/siaikin/home-dashboard/internal/pkg/notification"

const (
	notificationMessageType = "thirdParty:notification"
)

func SendNotificationMessage(_type string, data map[string]any) {
	notification.Send(notificationMessageType, map[string]any{
		"type": _type,
		"data": data,
	})
}
