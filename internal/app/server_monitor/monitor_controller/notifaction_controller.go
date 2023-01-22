package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"io"
	"log"
	"time"
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

	messageNotify := make(chan *NotificationMessage, 1)
	defer func() {
		close(messageNotify)
		log.Println("messageNotify channel close")
	}()

	done := make(chan bool, 1)
	defer func() {
		close(done)
		log.Println("done channel close")
	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				log.Println("realtime loop finish")
				return
			case <-ticker.C:
				messageNotify <- &NotificationMessage{
					Type: RealtimeStat,
					Data: gin.H{RealtimeStat: monitor_service.GetSystemRealtimeStat()},
				}
			}
		}
	}()

	context.Stream(func(w io.Writer) bool {
		message, ok := <-messageNotify
		if !ok {
			log.Println("message send loop finish")
			return false
		}

		defer func() {
			err := recover()
			if err != nil {
				log.Println(err)
			}
		}()
		context.SSEvent(message.Type, message.Data)

		return true
	})
}
