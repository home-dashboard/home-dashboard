package monitor_process_realtime

import (
	"testing"
	"time"
)

func BenchmarkGetProcessRealtimeStatistic(b *testing.B) {
	getProcessRealtimeStatistic()
}

func TestProcessRealtimeStatLoop(t *testing.T) {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-done:
			ticker.Stop()
			return
		case <-ticker.C:
			processStatList, relationship = getProcessRealtimeStatistic()
		}
	}
}
