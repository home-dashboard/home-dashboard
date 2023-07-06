package monitor_process_realtime

import (
	"context"
	"fmt"
	psuProc "github.com/shirou/gopsutil/v3/process"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/notification"
	"runtime"
	"sort"
	"time"
)

var logger = comfy_log.New("[monitor_process_realtime]")

var processMap = make(map[int32]*psuProc.Process)

var processStatList, relationship = getProcessRealtimeStatistic()

func getProcessRealtimeStatistic() ([]*ProcessRealtimeStat, map[int32]*ProcessNode) {
	relationshipMap := make(map[int32]*ProcessNode)

	pids, _ := psuProc.Pids()
	processStatList := make([]*ProcessRealtimeStat, 0, len(pids))

	fillFailProcessIds := make([]int32, 0)

	for _, pid := range pids {
		var proc *psuProc.Process
		var err error

		if processMap[pid] != nil {
			proc = processMap[pid]
		} else {
			proc, err = psuProc.NewProcess(pid)
			processMap[pid] = proc

			if err != nil {
				logger.Warn("create process instance failed, %s\n", err)
				continue
			}
		}

		if running, err := proc.IsRunning(); !running || err != nil {
			delete(processMap, pid)
			continue
		}

		if ignoredProcess(proc) {
			continue
		}

		var ppid int32
		if ppid, err = proc.Ppid(); err != nil {
			logger.Warn("process get ppid failed, %s\n", err)
			continue
		}

		if relationshipMap[ppid] == nil {
			relationshipMap[ppid] = &ProcessNode{Pid: ppid}
		}

		if relationshipMap[proc.Pid] == nil {
			relationshipMap[proc.Pid] = &ProcessNode{Pid: proc.Pid}
		}

		node := relationshipMap[proc.Pid]
		node.Parent = relationshipMap[ppid]

		stat := &ProcessRealtimeStat{}

		errs := fillProcessStat(stat, proc)
		if len(errs) > 0 {
			fillFailProcessIds = append(fillFailProcessIds, stat.Pid)
		}

		processStatList = append(processStatList, stat)
	}

	if len(fillFailProcessIds) > 0 {
		logger.Warn("fill failed process ids: %d\n", fillFailProcessIds)
	}
	return processStatList, relationshipMap
}

// fillProcessStat 填充 *ProcessRealtimeStat 指向的数据结构, 即使某个属性填充失败也不会中断执行.
// 填充失败的属性将保持该属性对应的默认值.
func fillProcessStat(stat *ProcessRealtimeStat, proc *psuProc.Process) []error {
	errs := make([]error, 0)

	stat.Pid = proc.Pid

	if percent, err := proc.MemoryPercent(); err != nil {
		errs = append(errs, fmt.Errorf("process get memory percent failed, %s\n", err))
		errs = append(errs, fmt.Errorf("process get memory percent failed, %s\n", err))
	} else {
		stat.MemoryPercent = percent
	}

	if percent, err := proc.Percent(0); err != nil {
		errs = append(errs, fmt.Errorf("process get cpu percent failed, %s\n", err))
	} else {
		count := runtime.NumCPU()
		stat.CpuPercent = percent / float64(count)
	}

	//if size, err := proc.NumThreads(); err != nil {
	//	errs = append(errs, fmt.Errorf("process get thread size failed, %s\n", err))
	//} else {
	//	stat.ThreadSize = size
	//}

	if name, err := proc.Name(); err != nil {
		errs = append(errs, fmt.Errorf("process get name failed, %s\n", err))
	} else {
		stat.Name = name
	}

	if createTime, err := proc.CreateTime(); err != nil {
		errs = append(errs, fmt.Errorf("process get create time failed, %s\n", err))
	} else {
		stat.CreateTime = createTime
	}

	if username, err := proc.Username(); err != nil {
		errs = append(errs, fmt.Errorf("process get username failed, %s\n", err))
	} else {
		stat.Username = username
	}

	return errs
}

// GetRealtimeStat 获取最新的实时进程统计信息. 更新频率来自 StartRealtimeLoop 函数的调用参数.
// max 为返回的最大进程数, 如果 max 大于实际进程数或 max 小于 0, 则返回所有进程.
func GetRealtimeStat(max int) ([]*ProcessRealtimeStat, map[int32]*ProcessNode) {
	length := len(processStatList)

	if length < max || max < 0 {
		max = length
	}
	return processStatList[0:max], relationship
}

func SortByMemoryUsage(max int) ([]*ProcessRealtimeStat, map[int32]*ProcessNode) {
	length := len(processStatList)
	copied := make([]*ProcessRealtimeStat, length)
	copy(copied, processStatList)

	sort.SliceStable(copied, func(i, j int) bool {
		return copied[i].MemoryPercent > copied[j].MemoryPercent
	})

	if length < max {
		max = length
	}
	return copied[0:max], relationship
}

func SortByCpuUsage(max int) ([]*ProcessRealtimeStat, map[int32]*ProcessNode) {
	length := len(processStatList)
	copied := make([]*ProcessRealtimeStat, length)
	copy(copied, processStatList)

	sort.SliceStable(copied, func(i, j int) bool {
		return copied[i].CpuPercent > copied[j].CpuPercent
	})

	if length < max {
		max = length
	}
	return copied[0:max], relationship
}

const (
	MessageType = "processRealtimeStat"
)

func Loop(context context.Context, d time.Duration) {
	ticker := time.NewTicker(d)

	go func() {
		for {
			select {
			case <-context.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				processStatList, relationship = getProcessRealtimeStatistic()
				notification.Send(MessageType, map[string]any{MessageType: processStatList})
			}
		}
	}()
}
