package monitor_process_realtime

import "encoding/json"

type ProcessRealtimeStat struct {
	Pid           int32   `json:"pid"`
	Name          string  `json:"name"`
	Username      string  `json:"username"`
	CpuPercent    float64 `json:"cpuPercent"`
	MemoryPercent float32 `json:"memoryPercent"`
	ThreadSize    int32   `json:"ThreadSize"` // Deprecated: gopsutil 对线程个数查询的方法耗时较长, 且该属性目前没有任何用例, 因此弃用.
	CreateTime    int64   `json:"createTime"`
}

func (c ProcessRealtimeStat) String() string {
	marshal, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return string(marshal)
}

type ProcessNode struct {
	Pid    int32        `json:"pid"`
	Parent *ProcessNode `json:"parent"`
}
