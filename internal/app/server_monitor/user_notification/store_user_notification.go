package user_notification

import (
	"context"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/notification"
)

var logger = comfy_log.New("notification")

func StartListenUserNotificationNotify(context context.Context) {
	go func() {
		var listener = notification.GetListener()
		var listenerCh = listener.Ch()
		defer func() {
			listener.Close()
			logger.Info("listener close complete\n")
		}()

		for {
			select {
			case <-context.Done():
				logger.Info("stop listen user notification notify\n")
				return
			case message, ok := <-listenerCh:
				if !ok {
					break
				}

				switch message.Type {
				case notification.UserNotificationReceivedMessageType:
					userNotifications, ok := message.Data["notifications"].([]notification.UserNotification)

					if !ok {
						logger.Error("invalid user notifications\n")
						continue
					}

					storedNotifications := make([]monitor_model.StoredNotification, len(userNotifications))
					for i, userNotification := range userNotifications {
						storedNotifications[i] = monitor_model.StoredNotification{
							UserNotification: userNotification,
						}
					}
					if err := monitor_service.CreateOrUpdateNotifications(storedNotifications); err != nil {
						logger.Error(comfy_errors.NewResponseError(comfy_errors.UnknownError, "store notification error, %w\n", err).Error())
						continue
					} else {
						logger.Info("store %d notifications success\n", len(storedNotifications))
						notification.Send("userNotification", map[string]any{})
					}

					break
				case notification.UserNotificationReadMessageType:
					unreadUniqueIds, ok := message.Data["unreadUniqueIds"].([]string)
					if !ok {
						logger.Error("invalid uniqueId list\n")
						continue
					}
					origin, ok := message.Data["origin"].(string)
					if !ok {
						logger.Error("invalid origin\n")
						continue
					}

					if updateIds, err := monitor_service.SyncNotificationsReadStateByUniqueIds(unreadUniqueIds, origin); err != nil {
						logger.Error(comfy_errors.NewResponseError(comfy_errors.UnknownError, "sync user notification read state error, %w\n", err).Error())
						continue
					} else {
						logger.Info("sync user notification read state success, update ids: %v\n", updateIds)

						if len(updateIds) > 0 {
							notification.Send("userNotification", map[string]any{})
						}
					}

					break
				}
			}
		}

	}()
}
