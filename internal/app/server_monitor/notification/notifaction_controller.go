package notification

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_process_realtime"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_realtime"
	"github.com/teivah/broadcast"
	"io"
	"log"
	"net/http"
)

const (
	RealtimeStat        = "realtimeStat"
	ProcessRealtimeStat = "processRealtimeStat"
)

type message struct {
	Type string
	Data map[string]any
}

func Notification(context *gin.Context) {
	log.Println("receive notification connection")

	context.Writer.Header().Set("Content-Type", "text/event-stream")
	context.Writer.Header().Set("Cache-Control", "no-cache")
	context.Writer.Header().Set("Connection", "keep-alive")
	context.Writer.Header().Set("Transfer-Encoding", "chunked")

	context.SSEvent("message", "notification channel connected")

	session := sessions.Default(context)
	statConfig := getCollectStatConfig(session)

	processConfig := statConfig.Process
	systemConfig := statConfig.System

	var listener *broadcast.Listener[*monitor_realtime.SystemRealtimeStat]
	var listenerCh <-chan *monitor_realtime.SystemRealtimeStat
	if systemConfig.Enable {
		listener = monitor_realtime.GetListener()
		defer func() {
			listener.Ch()
			listener.Close()
			log.Println("listener close complete")
		}()

		listenerCh = listener.Ch()
	}

	var processStatListener *broadcast.Listener[*[]*monitor_process_realtime.ProcessRealtimeStat]
	var processListenerCh <-chan *[]*monitor_process_realtime.ProcessRealtimeStat
	if processConfig.Enable {
		processStatListener = monitor_process_realtime.GetListener()
		defer func() {
			processStatListener.Close()
			log.Println("processStatListener close complete")
		}()

		processListenerCh = processStatListener.Ch()
	}

	context.Stream(func(w io.Writer) bool {
		defer func() {
			err := recover()
			if err != nil {
				log.Printf("sse disconnected %s\n", err)
			}
		}()

		message := message{}

		select {
		case realtimeStat := <-listenerCh:
			message.Type = RealtimeStat
			message.Data = gin.H{RealtimeStat: realtimeStat}
		case <-processListenerCh:
			message.Type = ProcessRealtimeStat
			switch processConfig.SortField {
			case sortByCpuUsage:
				sortedProcesses, _ := monitor_process_realtime.SortByCpuUsage(processConfig.Max)
				message.Data = gin.H{ProcessRealtimeStat: sortedProcesses}
			case sortByMemoryUsage:
				sortedProcesses, _ := monitor_process_realtime.SortByMemoryUsage(processConfig.Max)
				message.Data = gin.H{ProcessRealtimeStat: sortedProcesses}
			case normal:
			default:
				processes, _ := monitor_process_realtime.GetRealtimeStat(processConfig.Max)
				message.Data = gin.H{ProcessRealtimeStat: processes}
			}
		}

		context.SSEvent(message.Type, message.Data)
		return true
	})

	log.Println("notification connection finish")
}

// CollectStat 用于控制 Notification 收集的实时统计数据的类型.
// 接受的数据格式见 CollectStatConfig
func CollectStat(context *gin.Context) {
	var body CollectStatConfig

	if err := context.ShouldBindJSON(&body); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		panic(err)
	}

	session := sessions.Default(context)

	oldStatConfig := getCollectStatConfig(session)

	if err := copier.CopyWithOption(&oldStatConfig, &body, copier.Option{
		DeepCopy: true,
	}); err != nil {
		log.Printf("collect stat config merge failed, %s\n", err)
	}

	session.Set(CollectStatConfigSessionKey, oldStatConfig)

	if err := session.Save(); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		panic(err)
	}
}

func getCollectStatConfig(session sessions.Session) CollectStatConfig {
	config, _ := session.Get(CollectStatConfigSessionKey).(CollectStatConfig)

	return config
}
