package monitor_process_realtime

import (
	psuProc "github.com/shirou/gopsutil/v3/process"
)

func ignoreProcess(proc *psuProc.Process) bool {
	switch proc.Pid {
	case 0, 4:
		return true
	}

	return false
}
