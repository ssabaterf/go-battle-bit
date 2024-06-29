package metrics

import (
	"log/slog"
	"runtime"
	"time"
)

func MemTicker(breakChan chan struct{}, interval time.Duration) {
	ticker1s := time.NewTicker(interval)
	for {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		slog.Info("HeapAlloc in KB", "heap", memStats.HeapAlloc/1024)
		<-ticker1s.C
		select {
		case <-breakChan:
			{
				runtime.ReadMemStats(&memStats)
				slog.Info("HeapAlloc in KB", "heap", memStats.HeapAlloc/1024)
				return
			}
		default:
			continue
		}
	}
}