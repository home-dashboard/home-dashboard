package notification

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_process_realtime"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_realtime"
	"github.com/siaikin/home-dashboard/internal/pkg/authority"
	"github.com/siaikin/home-dashboard/internal/pkg/cache"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/notification"
	"github.com/siaikin/home-dashboard/third_party"
	"io"
	"net/http"
)

var logger = comfy_log.New("[notification]")

var collectConfigCache = cache.New("collectConfig")

func Notification(c *gin.Context) {
	logger.Info("receive notification connection\n")

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	c.SSEvent("message", "notification channel connected")

	session := sessions.Default(c)

	var _collectConfig = getCollectStatConfig(session)

	// 发送系统实时统计信息
	sendSystemRealtimeStatMessage := func(c *gin.Context, collectConfig CollectStatConfig, message notification.Message) {
		c.SSEvent(message.Type, message.Data)
	}
	// 发送进程实时统计信息
	sendProcessRealtimeStatMessage := func(c *gin.Context, collectConfig CollectStatConfig, message notification.Message) {
		processStatList := message.Data[message.Type].([]*monitor_process_realtime.ProcessRealtimeStat)
		delete(message.Data, message.Type)

		message.Data["sortField"] = collectConfig.Process.SortField
		message.Data["sortOrder"] = collectConfig.Process.SortOrder
		message.Data["max"] = collectConfig.Process.Max
		message.Data["total"] = len(processStatList)
		message.Data["cpuUsage"] = monitor_realtime.GetCpuPercent()
		message.Data["memoryUsage"] = monitor_realtime.GetMemoryPercent()

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

		c.SSEvent(message.Type, message.Data)
	}

	// 通知信道连接成功时, 立即发送一次实时统计信息. 以便客户端能够立即显示统计信息.
	// 立即发送的实时统计信息包含系统实时统计信息, 进程实时统计信息以及第三方模块的实时统计信息.
	sendSystemRealtimeStatMessage(c, _collectConfig, notification.Message{
		Type: monitor_realtime.MessageType,
		Data: map[string]interface{}{
			monitor_realtime.MessageType: monitor_realtime.GetCachedSystemRealtimeStat(),
		},
	})

	processes, _ := monitor_process_realtime.GetRealtimeStat(-1)
	sendProcessRealtimeStatMessage(c, _collectConfig, notification.Message{
		Type: monitor_process_realtime.MessageType,
		Data: map[string]interface{}{
			monitor_process_realtime.MessageType: processes,
		},
	})
	if err := third_party.DispatchEvent(third_party.NewNotificationChannelConnectedEvent(c)); err != nil {
		logger.Info("dispatch notification channel connected event failed, %s\n", err)
	}

	var listener = notification.GetListener()
	var listenerCh = listener.Ch()
	defer func() {
		listener.Close()
		logger.Info("listener close complete\n")
	}()
	c.Stream(func(w io.Writer) bool {
		defer func() {
			err := recover()
			if err != nil {
				logger.Info("system stat send failed, %s\n", err)
			}
		}()

		message, ok := <-listenerCh
		if !ok {
			return false
		}

		_collectConfig = getCollectStatConfig(session)

		switch message.Type {
		case monitor_realtime.MessageType:
			if _collectConfig.System.Enable {
				sendSystemRealtimeStatMessage(c, _collectConfig, message)
			}
			break
		case monitor_process_realtime.MessageType:
			if _collectConfig.Process.Enable {
				sendProcessRealtimeStatMessage(c, _collectConfig, message)
			}
			break
		case "userNotification":
			c.SSEvent(message.Type, message.Data)
			break
		}

		return true
	})

	logger.Info("notification connection finish\n")
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
		logger.Info("collect stat config merge failed, %s\n", err)
	}

	collectConfigCache.Set(user.Username, statConfig)
}

func GetCollectStat(context *gin.Context) {
	session := sessions.Default(context)

	statConfig := getCollectStatConfig(session)

	context.JSON(http.StatusOK, statConfig)
}

// getCollectStatConfig 通过 session 中的 username 获取统计数据收集配置
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

func getAuthInfo(session sessions.Session) authority.User {
	config, _ := session.Get(authority.InfoKey).(authority.User)

	return config
}
