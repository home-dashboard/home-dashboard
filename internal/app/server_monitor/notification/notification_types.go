package notification

import (
	"encoding/gob"
	"encoding/json"
)

const CollectStatConfigSessionKey = "collectStatConfig"

type ProcessStatSortField = string

const (
	sortByCpuUsage    ProcessStatSortField = "cpuUsage"
	sortByMemoryUsage                      = "memoryUsage"
	normal                                 = "default"
)

type CollectStatConfig struct {
	System struct {
		Enable bool `form:"enable"`
	} `form:"system"`
	Process struct {
		Enable    bool                 `form:"enable"`
		SortField ProcessStatSortField `form:"field"`
		SortOrder bool                 `form:"order"`
		Max       int                  `form:"max"`
	}
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
			Enable bool `form:"enable"`
		}{
			Enable: true,
		},
		Process: struct {
			Enable    bool                 `form:"enable"`
			SortField ProcessStatSortField `form:"field"`
			SortOrder bool                 `form:"order"`
			Max       int                  `form:"max"`
		}{},
	}
}

func init() {
	gob.Register(CollectStatConfig{})
}
