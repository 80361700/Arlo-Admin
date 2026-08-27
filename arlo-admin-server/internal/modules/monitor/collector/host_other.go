//go:build !darwin && !linux

package collector

import (
	"os"
	"runtime"

	"arlo-admin/internal/modules/monitor/dto"
)

func CollectCPU() *dto.CPUInfo {
	return &dto.CPUInfo{Cores: runtime.NumCPU()}
}

func CollectMem() *dto.MemInfo {
	return &dto.MemInfo{}
}

func CollectDisk() []dto.DiskInfo {
	return nil
}

func CollectSys() *dto.SysInfo {
	host, _ := os.Hostname()
	return &dto.SysInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: host,
	}
}
