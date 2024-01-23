package monitor_controller

import (
	"encoding/gob"
	"encoding/json"
)

type ProcessStatSortField = string

const (
	sortByCpuUsage    ProcessStatSortField = "cpuUsage"
	sortByMemoryUsage                      = "memoryUsage"
	normal                                 = "default"
)

type CollectStatConfig struct {
	System struct {
		Enable bool `json:"enable" form:"enable"`
	} `json:"system" form:"system"`
	Process struct {
		Enable    bool                 `json:"enable" form:"enable"`
		SortField ProcessStatSortField `json:"sortField" form:"field"`
		SortOrder bool                 `json:"sortOrder" form:"order"`
		Max       int                  `json:"max" form:"max"`
	} `json:"process" form:"process"`
}

func (c CollectStatConfig) String() string {
	marshal, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return string(marshal)
}

func DefaultCollectStatConfig() CollectStatConfig {
	return CollectStatConfig{
		System: struct {
			Enable bool `json:"enable" form:"enable"`
		}{Enable: true},
		Process: struct {
			Enable    bool                 `json:"enable" form:"enable"`
			SortField ProcessStatSortField `json:"sortField" form:"field"`
			SortOrder bool                 `json:"sortOrder" form:"order"`
			Max       int                  `json:"max" form:"max"`
		}{},
	}
}

func init() {
	gob.Register(CollectStatConfig{})
}
