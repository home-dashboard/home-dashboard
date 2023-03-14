package notification

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_process_realtime"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_realtime"
	"github.com/siaikin/home-dashboard/internal/pkg/authority"
	"github.com/siaikin/home-dashboard/internal/pkg/cache"
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

var collectConfigCache = cache.New("collectConfig")

func Notification(c *gin.Context) {
	log.Println("receive notification connection")

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	c.SSEvent("message", "notification channel connected")

	session := sessions.Default(c)

	var listener = monitor_realtime.GetListener()
	var listenerCh = listener.Ch()
	defer func() {
		listener.Close()
		log.Println("listener close complete")
	}()

	var processListener = monitor_process_realtime.GetListener()
	var processListenerCh = processListener.Ch()
	defer func() {
		processListener.Close()
		log.Println("process listener close complete")
	}()

	c.Stream(func(w io.Writer) bool {
		defer func() {
			err := recover()
			if err != nil {
				log.Printf("system stat send failed, %s\n", err)
			}
		}()

		message := message{}

		select {
		case realtimeStat := <-listenerCh:
			collectConfig := getCollectStatConfig(session)

			if collectConfig.System.Enable {
				c.SSEvent(RealtimeStat, gin.H{RealtimeStat: realtimeStat})
			}

			return true
		case processRealtimeStat := <-processListenerCh:
			collectConfig := getCollectStatConfig(session)

			if collectConfig.Process.Enable {
				message.Type = ProcessRealtimeStat
				message.Data = map[string]any{
					"sortField":   collectConfig.Process.SortField,
					"sortOrder":   collectConfig.Process.SortOrder,
					"max":         collectConfig.Process.Max,
					"total":       len(*processRealtimeStat),
					"cpuUsage":    monitor_realtime.GetCpuPercent(),
					"memoryUsage": monitor_realtime.GetMemoryPercent(),
				}

				switch collectConfig.Process.SortField {
				case sortByCpuUsage:
					sortedProcesses, _ := monitor_process_realtime.SortByCpuUsage(collectConfig.Process.Max)
					message.Data["processes"] = sortedProcesses
					break
				case sortByMemoryUsage:
					sortedProcesses, _ := monitor_process_realtime.SortByMemoryUsage(collectConfig.Process.Max)
					message.Data["processes"] = sortedProcesses
					break
				case normal:
				default:
					processes, _ := monitor_process_realtime.GetRealtimeStat(collectConfig.Process.Max)
					message.Data["processes"] = processes
				}

				c.SSEvent(ProcessRealtimeStat, message.Data)
			}

			return true
		}
	})

	log.Println("notification connection finish")
}

// ModifyCollectStat 用于控制 Notification 收集的实时统计数据的类型.
// 接受的数据格式见 CollectStatConfig
func ModifyCollectStat(context *gin.Context) {
	var body CollectStatConfig

	if err := context.ShouldBindJSON(&body); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		panic(err)
	}

	session := sessions.Default(context)

	statConfig := getCollectStatConfig(session)
	user := getAuthInfo(session)

	if err := copier.CopyWithOption(&statConfig, &body, copier.Option{
		DeepCopy: true,
	}); err != nil {
		log.Printf("collect stat config merge failed, %s\n", err)
	}

	collectConfigCache.Set(user.Username, statConfig)
}

func GetCollectStat(context *gin.Context) {
	session := sessions.Default(context)

	statConfig := getCollectStatConfig(session)

	context.JSON(http.StatusOK, statConfig)
}

func getCollectStatConfig(session sessions.Session) CollectStatConfig {
	user := getAuthInfo(session)

	cachedConfig, ok := collectConfigCache.Get(user.Username)
	if ok == false {
		cachedConfig = DefaultCollectStatConfig()
		collectConfigCache.Set(user.Username, cachedConfig)
	}

	collectConfig, _ := cachedConfig.(CollectStatConfig)

	return collectConfig
}

func getAuthInfo(session sessions.Session) monitor_model.User {
	config, _ := session.Get(authority.InfoKey).(monitor_model.User)

	return config
}
