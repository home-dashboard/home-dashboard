package overseer

import (
	"encoding/gob"
	"github.com/jinzhu/copier"
)

type StatusType int

type Status struct {
	Type StatusType `json:"type,omitempty"`
	Text string     `json:"text,omitempty"`
	// Extra 用于存储额外的信息, 例如进度等.
	// 1. 在 StatusTypeUpgrading 状态时, 会存储 written, total 表示已下载的更新文件大小和更新文件总大小.
	Extra map[string]any `json:"extra,omitempty"`
}

// 服务运行状态
const (
	// StatusTypeRunning 正常运行中
	StatusTypeRunning StatusType = 1 + iota
	StatusTypeUpgrading
	// StatusTypeRestarting 正在重启 worker
	StatusTypeRestarting
	// StatusTypeDestroyed 已销毁
	StatusTypeDestroyed
)

// 服务运行状态预设对象, 包括状态类型和状态文本
var (
	StatusRunning    = Status{Type: StatusTypeRunning, Text: "Running"}
	StatusUpgrading  = Status{Type: StatusTypeUpgrading, Text: "Upgrading"}
	StatusRestarting = Status{Type: StatusTypeRestarting, Text: "Restarting"}
	StatusDestroyed  = Status{Type: StatusTypeDestroyed, Text: "Destroyed"}
)

// StatusMessageType 服务运行状态变化时发送的通知消息的消息类型. 通过 [github.com/siaikin/home-dashboard/internal/app/server_monitor/notification] 的 sse 连接发送.
const StatusMessageType = "overseer.StatusMessageType"

// NewVersionMessageType 有新版本可用时发送的通知消息的消息类型. 通过 [github.com/siaikin/home-dashboard/internal/app/server_monitor/notification] 的 sse 连接发送.
const NewVersionMessageType = "overseer.NewVersionMessageType"

func init() {
	gob.Register(Status{})
}

func newStatus(statusType StatusType) Status {
	status := Status{}

	var defaultStatus = StatusRunning
	switch statusType {
	case StatusTypeRunning:
		defaultStatus = StatusRunning
	case StatusTypeUpgrading:
		defaultStatus = StatusUpgrading
	case StatusTypeRestarting:
		defaultStatus = StatusRestarting
	case StatusTypeDestroyed:
		defaultStatus = StatusDestroyed
	}

	_ = copier.Copy(&status, &defaultStatus)

	status.Extra = make(map[string]any)

	return status
}
