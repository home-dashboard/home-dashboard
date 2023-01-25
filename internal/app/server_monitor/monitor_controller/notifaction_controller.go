package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_realtime"
	"io"
	"log"
)

const (
	RealtimeStat = "realtimeStat"
)

type NotificationMessage struct {
	Type string
	Data map[string]any
}

func Notification(context *gin.Context) {
	log.Println("receive notification connection")
	defer log.Println("notification connection finish")

	context.Writer.Header().Set("Content-Type", "text/event-stream")
	context.Writer.Header().Set("Cache-Control", "no-cache")
	context.Writer.Header().Set("Connection", "keep-alive")
	context.Writer.Header().Set("Transfer-Encoding", "chunked")

	context.SSEvent("message", "notification channel connected")

	listener := monitor_realtime.GetListener()
	context.Stream(func(w io.Writer) bool {
		defer func() {
			err := recover()
			if err != nil {
				log.Printf("sse disconnected %s\n", err)
			}
		}()

		message := &NotificationMessage{
			Type: RealtimeStat,
			Data: gin.H{RealtimeStat: <-listener.Ch()},
		}
		context.SSEvent(message.Type, message.Data)
		return true
	})
}
