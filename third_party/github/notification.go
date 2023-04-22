package github

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v50/github"
	"github.com/siaikin/home-dashboard/third_party/internal"
	"path/filepath"
	"strconv"
	"time"
)

const (
	MessageType = "github"
)

var latestNotifications = make([]*github.Notification, 0)

// 开始轮询获取 GitHub 通知.
func startFetchNotificationLoop(context context.Context) {
	if timer != nil {
		return
	}

	timer = time.NewTimer(time.Nanosecond)

	go func() {
		for {
			select {
			case <-context.Done():
				timer.Stop()
				logger.Info("stop fetch notification loop")
				return
			case <-timer.C:
				var delay int64 = 30

				if _notifications, r, err := httpClient.ListNotifications(context, &github.NotificationListOptions{All: true}); err != nil {
					logger.Warn("fetch notification failed, %v", err)
				} else {
					latestNotifications = _notifications

					sendNotification(&latestNotifications)

					// 根据 X-Poll-Interval 延迟时间重置定时器.
					// https://docs.github.com/en/rest/activity/notifications?apiVersion=2022-11-28
					if delay, err = strconv.ParseInt(r.Header.Get("X-Poll-Interval"), 10, 0); err != nil {
						logger.Warn("parse X-Poll-Interval failed, %v", err)
					}
				}

				timer.Reset(time.Duration(delay) * time.Second)
			}
		}
	}()
}

// sendNotification 通过 [notification.Send] 发送 GitHub 通知.
func sendNotification(notifications *[]*github.Notification) {
	simpleNotifications := make([]*map[string]any, 0)

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

		//suffixPath, _ := filepath.Rel("https://api.github.com/repos/", notificationInfo.GetSubject().GetURL())

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

		simpleNotifications = append(simpleNotifications, &simpleNotification)
	}

	internal.SendNotificationMessage(MessageType, map[string]any{"notifications": simpleNotifications})
}

// getNotifications 获取最新获取的 GitHub 通知.
func getCachedNotifications() *[]*github.Notification {
	return &latestNotifications
}

func resetNotificationRequestHeader(c *gin.Context) {
	httpClient.ResetListNotificationsLastModified()
}
