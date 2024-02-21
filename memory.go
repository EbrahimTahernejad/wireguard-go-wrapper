package WG

import (
	"time"

	runtimeDebug "runtime/debug"
)

func forceFree(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			runtimeDebug.FreeOSMemory()
		}
	}()
}

func LimitMemory(maxMemory int64, refreshInterval int) {
	runtimeDebug.SetGCPercent(10)
	runtimeDebug.SetMemoryLimit(maxMemory)
	if refreshInterval > 0 {
		duration := time.Duration(refreshInterval) * time.Second
		forceFree(duration)
	}
}