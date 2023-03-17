//go:build !windows

package monitor_process_realtime

import psuProc "github.com/shirou/gopsutil/v3/process"

func ignoredProcess(proc *psuProc.Process) bool {
	return false
}
