package model

import "github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"

// Project 表示一个项目
type Project struct {
	monitor_model.Model
	Name           string                 `json:"name" binding:"required"`
	DisplayName    string                 `json:"displayName"`
	Description    string                 `json:"description"`
	ExecuteRecords []ProjectExecuteRecord `json:"executeRecords"`
}

// Repository 表示 Project 的 git 仓库
type Repository struct {
	ProjectID uint   `json:"projectId"`
	Name      string `json:"name"`
	URL       string `json:"url"`
}

type ProjectExecuteRecord struct {
	monitor_model.Model
	ProjectID uint   `json:"projectId"`
	Branch    string `json:"branch"`
	Hash      string `json:"hash"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
}
