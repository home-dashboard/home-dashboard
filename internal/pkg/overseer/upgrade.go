package overseer

import "encoding/gob"

type StatusType int

type Status struct {
	Type StatusType `json:"type,omitempty"`
	Text string     `json:"text,omitempty"`
}

const (
	// StatusTypeRunning 正常运行中
	StatusTypeRunning StatusType = 1 + iota
	StatusTypeUpgrading
	// StatusTypeRestarting 正在重启 worker
	StatusTypeRestarting
	// StatusTypeDestroyed 已销毁
	StatusTypeDestroyed
)

var (
	StatusRunning    = Status{Type: StatusTypeRunning, Text: "Running"}
	StatusUpgrading  = Status{Type: StatusTypeUpgrading, Text: "Upgrading"}
	StatusRestarting = Status{Type: StatusTypeRestarting, Text: "Restarting"}
	StatusDestroyed  = Status{Type: StatusTypeDestroyed, Text: "Destroyed"}
)

const StatusMessageType = "overseer.StatusMessageType"

func init() {
	gob.Register(Status{})
}
