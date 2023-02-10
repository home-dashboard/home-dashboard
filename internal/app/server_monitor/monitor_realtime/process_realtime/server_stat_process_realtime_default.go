//go:build !windows

package process_realtime

import psuProc "github.com/shirou/gopsutil/v3/process"

func ignoreProcess(proc *psuProc.Process) bool {
	return false
}
