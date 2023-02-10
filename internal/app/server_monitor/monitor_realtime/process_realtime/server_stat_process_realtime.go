package process_realtime

import (
	psuProc "github.com/shirou/gopsutil/v3/process"
	"github.com/teivah/broadcast"
	"log"
	"runtime"
	"sort"
	"time"
)

//var processMap = make(map[int32]*psuProc.Process)

var processStatList, relationship = getProcessRealtimeStatistic()

func getProcessRealtimeStatistic() ([]*ProcessRealtimeStat, map[int32]*ProcessNode) {
	relationshipMap := make(map[int32]*ProcessNode)

	processes, _ := psuProc.Processes()
	processStatList := make([]*ProcessRealtimeStat, 0, len(processes))

	for _, proc := range processes {
		var err error
		if ignoreProcess(proc) {
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

		fillProcessStat(stat, proc)

		processStatList = append(processStatList, stat)
	}

	return processStatList, relationshipMap
}

// fillProcessStat 填充 *ProcessRealtimeStat 指向的数据结构, 即使某个属性填充失败也不会中断执行.
// 填充失败的属性将保持该属性对应的默认值.
func fillProcessStat(stat *ProcessRealtimeStat, proc *psuProc.Process) {
	stat.Pid = proc.Pid

	if percent, err := proc.MemoryPercent(); err != nil {
		log.Printf("process get memory percent failed, %s\n", err)
	} else {
		stat.MemoryPercent = percent
	}

	if percent, err := proc.Percent(0); err != nil {
		log.Printf("process get cpu percent failed, %s\n", err)
	} else {
		count := runtime.NumCPU()
		stat.CpuPercent = percent / float64(count)
	}

	if size, err := proc.NumThreads(); err != nil {
		log.Printf("process get thread size failed, %s\n", err)
	} else {
		stat.ThreadSize = size
	}

	if name, err := proc.Name(); err != nil {
		log.Printf("process get name failed, %s\n", err)
	} else {
		stat.Name = name
	}

	if createTime, err := proc.CreateTime(); err != nil {
		log.Printf("process get create time failed, %s\n", err)
	} else {
		stat.CreateTime = createTime
	}

	if username, err := proc.Username(); err != nil {
		log.Printf("process get username failed, %s\n", err)
	} else {
		stat.Username = username
	}
}

func GetRealtimeStat() ([]*ProcessRealtimeStat, map[int32]*ProcessNode) {
	return processStatList, relationship
}

func SortByMemoryUsage() ([]*ProcessRealtimeStat, map[int32]*ProcessNode) {
	copied := make([]*ProcessRealtimeStat, len(processStatList))
	copy(copied, processStatList)

	sort.SliceStable(copied, func(i, j int) bool {
		return copied[i].MemoryPercent > copied[j].MemoryPercent
	})

	return copied, relationship
}

func SortByCpuUsage() ([]*ProcessRealtimeStat, map[int32]*ProcessNode) {
	copied := make([]*ProcessRealtimeStat, len(processStatList))
	copy(copied, processStatList)

	sort.SliceStable(copied, func(i, j int) bool {
		return copied[i].CpuPercent > copied[j].CpuPercent
	})

	return copied, relationship
}

var done = make(chan bool)
var relay = broadcast.NewRelay[*map[int32]*ProcessNode]()

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
				relay.Notify(&relationship)
			}
		}
	}()
}

func StopRealtimeLoop() {
	relay.Close()
	done <- true
}

func GetListener() *broadcast.Listener[*map[int32]*ProcessNode] {
	return relay.Listener(1)
}
