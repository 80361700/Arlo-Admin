package collector

import (
	"runtime"
	"time"

	"arlo-admin/internal/modules/monitor/dto"
	"arlo-admin/pkg/appruntime"
	"arlo-admin/pkg/utils"
)

// CollectGo 采集 Go 运行时指标
func CollectGo() *dto.GoInfo {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	lastGC := ""
	if ms.LastGC > 0 {
		lastGC = utils.FormatTime(time.Unix(0, int64(ms.LastGC)))
	}

	return &dto.GoInfo{
		Version:       runtime.Version(),
		Goroutines:    runtime.NumGoroutine(),
		GOMAXPROCS:    runtime.GOMAXPROCS(0),
		HeapAlloc:     ms.HeapAlloc,
		HeapSys:       ms.HeapSys,
		HeapInuse:     ms.HeapInuse,
		GCCPUFraction: ms.GCCPUFraction,
		NumGC:         ms.NumGC,
		LastGC:        lastGC,
	}
}

// CollectApp 应用进程信息
func CollectApp(mode string) *dto.AppInfo {
	start := appruntime.StartedAt
	return &dto.AppInfo{
		Name:       "arlo-admin",
		Mode:       mode,
		StartTime:  utils.FormatTime(start),
		RunSeconds: int64(time.Since(start).Seconds()),
		GoVersion:  runtime.Version(),
	}
}

func pct(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(used) / float64(total) * 100
}
