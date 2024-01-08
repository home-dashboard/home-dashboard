package model

import "github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"

// Project 表示一个项目
type Project struct {
	monitor_model.Model
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// Repository 表示 Project 的 git 仓库
type Repository struct {
	ProjectID uint   `json:"projectId"`
	Name      string `json:"name"`
	URL       string `json:"url"`
}
