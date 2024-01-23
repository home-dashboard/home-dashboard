package monitor_controller

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_process_realtime"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_realtime"
	"github.com/siaikin/home-dashboard/internal/pkg/authority"
	"github.com/siaikin/home-dashboard/internal/pkg/cache"
	"github.com/siaikin/home-dashboard/internal/pkg/notification"
	"github.com/siaikin/home-dashboard/internal/pkg/overseer"
	"github.com/siaikin/home-dashboard/third_party"
	"io"
	"math/rand"
	"net/http"
	"strconv"
)

var collectConfigCache = cache.New("collectConfig")

// Notification 用于向客户端发送实时统计信息
// @Summary 用于向客户端发送实时统计信息
// @Description 用于向客户端发送实时统计信息
// @Tags Notification
// @Accept json
// @Produce json
// @Success 200 {string} string "ok"
// @Router /notification [get]
func Notification(c *gin.Context) {
	randomKey := strconv.Itoa(rand.Int())
	logger.Info("[%s] notification channel connected, remote ip: %s, client ip: %s\n", randomKey, c.RemoteIP(), c.ClientIP())
	defer logger.Info("[%s] notification channel disconnected, remote ip: %s, client ip: %s\n", randomKey, c.RemoteIP(), c.ClientIP())

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	c.SSEvent("message", "notification channel connected")

	session := sessions.Default(c)

	// 通知信道连接成功时, 立即发送一次实时统计信息. 以便客户端能够立即显示统计信息.
	// 立即发送的实时统计信息包含系统实时统计信息, 进程实时统计信息以及第三方模块的实时统计信息.
	var collectStatConfig = getCollectStatConfig(session)
	sendSystemRealtimeStatMessage(c, collectStatConfig, notification.Message{
		Type: monitor_realtime.MessageType,
		Data: map[string]interface{}{
			monitor_realtime.MessageType: monitor_realtime.GetCachedSystemRealtimeStat(),
		},
	})
	processes, _ := monitor_process_realtime.GetRealtimeStat(-1)
	sendProcessRealtimeStatMessage(c, collectStatConfig, notification.Message{
		Type: monitor_process_realtime.MessageType,
		Data: map[string]interface{}{
			monitor_process_realtime.MessageType: processes,
		},
	})

	// 检查是否有新版本可用并发送更新通知
	if overseerInst, err := overseer.Get(); err != nil {
		logger.Warn("not found overseer instance, %s\n", err)
	} else {
		if versionInfo, err := overseerInst.LatestVersionInfo(); err != nil {
			logger.Info("check latest version info failed, %s\n", err)
		} else {
			c.SSEvent(overseer.NewVersionMessageType, versionInfo)
		}
	}

	// 通知所有第三方模块, 通知消息信道已连接
	if err := third_party.DispatchEvent(third_party.NewNotificationChannelConnectedEvent(c)); err != nil {
		logger.Info("dispatch notification channel connected event failed, %s\n", err)
	}

	var listener = notification.GetListener()
	var listenerCh = listener.Ch()
	defer listener.Close()
	c.Stream(func(w io.Writer) bool {
		defer func() {
			err := recover()
			if err != nil {
				logger.Error("notification send failed, %s\n", err)
			}
		}()

		// 1. 从监听器中获取消息
		message, ok := <-listenerCh
		if !ok {
			return false
		}

		// 2. 获取保存在 session 消息通知配置
		collectStatConfig = getCollectStatConfig(session)

		// 3. 根据获取到的消息的类型, 发送对应的实时统计信息
		switch message.Type {
		case monitor_realtime.MessageType:
			if collectStatConfig.System.Enable {
				sendSystemRealtimeStatMessage(c, collectStatConfig, message)
			}
			break
		case monitor_process_realtime.MessageType:
			if collectStatConfig.Process.Enable {
				sendProcessRealtimeStatMessage(c, collectStatConfig, message)
			}
			break
		case "userNotification":
		case overseer.StatusMessageType:
			c.SSEvent(message.Type, message.Data)
			break
		}

		return true
	})
}

// ModifyCollectStat 用于控制 Notification 收集的实时统计数据的类型.
// 接受的数据格式见 CollectStatConfig
func ModifyCollectStat(context *gin.Context) {
	var body CollectStatConfig

	if err := context.ShouldBindJSON(&body); err != nil {
		respondEntityValidationError(context, err.Error())
		return
	}

	session := sessions.Default(context)

	statConfig := getCollectStatConfig(session)
	user := getAuthInfo(session)

	if err := copier.CopyWithOption(&statConfig, &body, copier.Option{
		DeepCopy: true,
	}); err != nil {
		respondUnknownError(context, err.Error())
		return
	}

	collectConfigCache.Set(user.Username, statConfig)

	context.JSON(http.StatusOK, gin.H{})
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
	if !ok {
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

// 发送系统实时统计信息
func sendSystemRealtimeStatMessage(c *gin.Context, collectConfig CollectStatConfig, message notification.Message) {
	c.SSEvent(message.Type, message.Data)
}

// 发送进程实时统计信息
func sendProcessRealtimeStatMessage(c *gin.Context, collectConfig CollectStatConfig, message notification.Message) {
	processStatList := message.Data[message.Type].([]*monitor_process_realtime.ProcessRealtimeStat)

	// 从 message 创建新的 map 对象, 以避免并发读写冲突
	restructureMessage := map[string]any{
		"sortField":   collectConfig.Process.SortField,
		"sortOrder":   collectConfig.Process.SortOrder,
		"max":         collectConfig.Process.Max,
		"total":       len(processStatList),
		"cpuUsage":    monitor_realtime.GetCpuPercent(),
		"memoryUsage": monitor_realtime.GetMemoryPercent(),
	}

	switch collectConfig.Process.SortField {
	case sortByCpuUsage:
		sortedProcesses, _ := monitor_process_realtime.SortByCpuUsage(collectConfig.Process.Max)
		restructureMessage["processes"] = sortedProcesses
		break
	case sortByMemoryUsage:
		sortedProcesses, _ := monitor_process_realtime.SortByMemoryUsage(collectConfig.Process.Max)
		restructureMessage["processes"] = sortedProcesses
		break
	case normal:
	default:
		processes, _ := monitor_process_realtime.GetRealtimeStat(collectConfig.Process.Max)
		restructureMessage["processes"] = processes
	}

	c.SSEvent(message.Type, restructureMessage)
}
