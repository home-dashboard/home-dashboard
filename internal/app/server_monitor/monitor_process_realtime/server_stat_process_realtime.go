package monitor_process_realtime

import (
	"fmt"
	psuProc "github.com/shirou/gopsutil/v3/process"
	"github.com/teivah/broadcast"
	"log"
	"runtime"
	"sort"
	"time"
)

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
				log.Printf("create process instance failed, %s\n", err)
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
			log.Printf("process get ppid failed, %s\n", err)
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
		log.Printf("fill failed process ids: %d\n", fillFailProcessIds)
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
	} else {
		stat.MemoryPercent = percent
	}

	if percent, err := proc.Percent(0); err != nil {
		errs = append(errs, fmt.Errorf("process get cpu percent failed, %s\n", err))
	} else {
		count := runtime.NumCPU()
		stat.CpuPercent = percent / float64(count)
	}

	if size, err := proc.NumThreads(); err != nil {
		errs = append(errs, fmt.Errorf("process get thread size failed, %s\n", err))
	} else {
		stat.ThreadSize = size
	}

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

func GetRealtimeStat(max int) ([]*ProcessRealtimeStat, map[int32]*ProcessNode) {
	length := len(processStatList)

	if length < max {
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

var done = make(chan bool)
var relay = broadcast.NewRelay[*[]*ProcessRealtimeStat]()

func StartRealtimeLoop(d time.Duration) {
	ticker := time.NewTicker(d)

	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				processStatList, relationship = getProcessRealtimeStatistic()
				relay.Notify(&processStatList)
			}
		}
	}()
}

func StopRealtimeLoop() {
	relay.Close()
	done <- true
}

func GetListener() *broadcast.Listener[*[]*ProcessRealtimeStat] {
	return relay.Listener(1)
}
