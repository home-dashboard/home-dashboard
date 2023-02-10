package process_realtime

import "encoding/json"

type ProcessRealtimeStat struct {
	Pid           int32   `json:"pid"`
	Name          string  `json:"name"`
	Username      string  `json:"username"`
	CpuPercent    float64 `json:"cpuPercent"`
	MemoryPercent float32 `json:"memoryPercent"`
	ThreadSize    int32   `json:"ThreadSize"`
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
