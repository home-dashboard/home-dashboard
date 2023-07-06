package github

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v50/github"
	"github.com/siaikin/home-dashboard/internal/pkg/notification"
	"github.com/siaikin/home-dashboard/third_party/internal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var latestNotifications = make([]*github.Notification, 0)

// 开始轮询获取 GitHub 通知.
func startFetchNotificationLoop(context context.Context) {
	timer := time.NewTimer(time.Nanosecond)
	// 检查通知已读状态的定时器
	checkReadTimer := time.NewTimer(time.Minute * 10)

	go func() {
		for {
			select {
			case <-context.Done():
				timer.Stop()
				logger.Info("stop fetch notification loop\n")
				return
			case <-timer.C:
				var delay int64 = 30

				if _notifications, r, err := httpClient.ListNotificationsByLastModified(context, &github.NotificationListOptions{}); err != nil {
					logger.Warn("fetch notification failed, %v\n", err)
				} else {
					latestNotifications = _notifications

					sendNotifications(&latestNotifications)

					// 根据 X-Poll-Interval 延迟时间重置定时器.
					// https://docs.github.com/en/rest/activity/notifications?apiVersion=2022-11-28
					if delay, err = strconv.ParseInt(r.Header.Get("X-Poll-Interval"), 10, 0); err != nil {
						logger.Warn("parse X-Poll-Interval failed, %v\n", err)
					}

					logger.Info("update %d notification, delay %d seconds\n", len(latestNotifications), delay)
				}

				timer.Reset(time.Duration(delay) * time.Second)

				break
			case <-checkReadTimer.C:
				// 获取所有未读通知, 并同步本地通知已读状态.
				if _notifications, _, err := httpClient.Activity.ListNotifications(context, &github.NotificationListOptions{}); err != nil {
					logger.Warn("fetch all unread notification failed, %v\n", err)
				} else {
					latestNotifications = _notifications

					syncNotificationsUnreadState(&latestNotifications)

					logger.Info("update %d unread notification\n", len(latestNotifications))
				}

				checkReadTimer.Reset(time.Minute * 10)
				break
			}
		}
	}()
}

// sendNotifications 通过 [notification.SendUserNotification] 发送 GitHub 通知.
func sendNotifications(notifications *[]*github.Notification) {
	userNotifications := make([]*notification.UserNotification, 0, len(*notifications))

	for _, notificationInfo := range *notifications {
		id := filepath.Base(notificationInfo.GetSubject().GetURL())

		simpleNotification := map[string]any{
			"id": id,
			"repository": map[string]any{
				"name": notificationInfo.GetRepository().GetName(),
				"url":  notificationInfo.GetRepository().GetHTMLURL(),
			},
			"unread":    notificationInfo.GetUnread(),
			"reason":    notificationInfo.GetReason(),
			"updatedAt": notificationInfo.GetUpdatedAt().UnixMilli(),
			"title":     notificationInfo.GetSubject().GetTitle(),
			"type":      notificationInfo.GetSubject().GetType(),
		}

		var url string
		switch notificationInfo.GetSubject().GetType() {
		case "Issue":
			url = notificationInfo.GetRepository().GetHTMLURL() + "/issues/" + id
			break
		case "PullRequest":
			url = notificationInfo.GetRepository().GetHTMLURL() + "/pull/" + id
			break
		case "Commit":
			url = notificationInfo.GetRepository().GetHTMLURL() + "/commit/" + id
			break
		case "Release":
			url = notificationInfo.GetRepository().GetHTMLURL() + "/releases/" + id
			break
		case "RepositoryInvitation":
			url = notificationInfo.GetRepository().GetHTMLURL()
			break
		case "SecurityAlert":
			url = notificationInfo.GetRepository().GetHTMLURL() + "/network/alerts"
			break
		case "Discussion":
			url = notificationInfo.GetRepository().GetHTMLURL() + "/discussions/" + id
			break
		}

		simpleNotification["url"] = url

		// 从 [github.NotificationSubject.LatestCommentURL] 中提取最新评论的 URL 的 id.
		// 并以 [#issuecomment-{id}] 的形式拼接到 url 末尾.
		commentId := filepath.Base(notificationInfo.GetSubject().GetLatestCommentURL())
		simpleNotification["lastCommentUrl"] = url + "#issuecomment-" + commentId

		userNotifications = append(userNotifications, &notification.UserNotification{
			UniqueId:       notificationInfo.GetID(),
			Unread:         notificationInfo.GetUnread(),
			Title:          notificationInfo.GetRepository().GetFullName(),
			Caption:        notificationInfo.GetSubject().GetTitle(),
			Link:           url + "#issuecomment-" + commentId,
			Kind:           internal.UserNotificationKindInfo,
			Origin:         internal.UserNotificationOriginGithub,
			OriginCreateAt: notificationInfo.GetUpdatedAt().UnixMilli(),
		})

		internal.SendUserNotifications(inst, userNotifications)
	}
}

// syncNotificationsUnreadState 标记用户的 GitHub 通知为已读.
func syncNotificationsUnreadState(notifications *[]*github.Notification) {
	uniqueIds := make([]string, 0, len(*notifications))

	for _, notificationInfo := range *notifications {
		uniqueIds = append(uniqueIds, notificationInfo.GetID())
	}

	internal.SyncUserNotificationsUnreadState(inst, uniqueIds, internal.UserNotificationOriginGithub)
}

// getNotifications 获取最新获取的 GitHub 通知.
func getCachedNotifications() *[]*github.Notification {
	return &latestNotifications
}

func resetNotificationRequestHeader(c *gin.Context) {
	httpClient.ResetListNotificationsLastModified()
}

func markNotificationAsRead(uniqueId string) error {
	if _, err := httpClient.Activity.MarkThreadRead(context.Background(), strings.SplitN(uniqueId, "-", 2)[1]); err != nil {
		return err
	}

	logger.Info("mark notification as read, %s\n", uniqueId)
	return nil
}
