package monitor_process_realtime

import (
	psuProc "github.com/shirou/gopsutil/v3/process"
)

// ignoredProcess 在 windows 平台上忽略的进程
// 0: 系统空闲进程
// 4: 系统进程
// 8: 系统进程
// from https://en.wikipedia.org/wiki/Process_identifier#Microsoft_Windows
func ignoredProcess(proc *psuProc.Process) bool {
	switch proc.Pid {
	case 0, 4:
		return true
	}

	return false
}
