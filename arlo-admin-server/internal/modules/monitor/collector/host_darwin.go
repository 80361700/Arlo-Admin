//go:build darwin

package collector

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"arlo-admin/internal/modules/monitor/dto"

	"golang.org/x/sys/unix"
)

func CollectCPU() *dto.CPUInfo {
	info := &dto.CPUInfo{Cores: runtime.NumCPU()}
	info.Load1, info.Load5, info.Load15 = darwinLoadAvg()

	// 新版 macOS 已无 kern.cp_time，优先用进程 %cpu 汇总；再退回负载估算
	if pct, ok := darwinCPUFromTicks(); ok {
		info.UsagePct = pct
	} else if pct, ok := darwinCPUFromPS(); ok {
		info.UsagePct = pct
	} else if info.Cores > 0 && info.Load1 > 0 {
		info.UsagePct = info.Load1 / float64(info.Cores) * 100
		if info.UsagePct > 100 {
			info.UsagePct = 100
		}
	}
	return info
}

func darwinLoadAvg() (l1, l5, l15 float64) {
	raw, err := unix.SysctlRaw("vm.loadavg")
	if err != nil {
		return
	}
	type loadavg struct {
		LdAvg  [3]uint32
		Fscale int64
	}
	if len(raw) < int(unsafe.Sizeof(loadavg{})) {
		return
	}
	lav := (*loadavg)(unsafe.Pointer(&raw[0]))
	scale := float64(lav.Fscale)
	if scale == 0 {
		return
	}
	return float64(lav.LdAvg[0]) / scale, float64(lav.LdAvg[1]) / scale, float64(lav.LdAvg[2]) / scale
}

// darwinCPUFromTicks 旧版 macOS 的 kern.cp_time（新系统 oid 不存在则失败）
func darwinCPUFromTicks() (float64, bool) {
	t1, i1, err1 := darwinCPUTimes()
	if err1 != nil {
		return 0, false
	}
	time.Sleep(400 * time.Millisecond)
	t2, i2, err2 := darwinCPUTimes()
	if err2 != nil || t2 <= t1 {
		return 0, false
	}
	idleDelta := float64(i2 - i1)
	totalDelta := float64(t2 - t1)
	if totalDelta <= 0 {
		return 0, false
	}
	usage := (1 - idleDelta/totalDelta) * 100
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	return usage, true
}

func darwinCPUTimes() (total, idle uint64, err error) {
	// USER=0 SYSTEM=1 IDLE=2 NICE=3
	raw, err := unix.SysctlRaw("kern.cp_time")
	if err != nil || len(raw) < 16 {
		return 0, 0, unix.ENOENT
	}
	vals, err := parseCPUTicks(raw)
	if err != nil || len(vals) < 3 {
		return 0, 0, err
	}
	for _, v := range vals {
		total += v
	}
	idle = vals[2]
	return total, idle, nil
}

func parseCPUTicks(raw []byte) ([]uint64, error) {
	switch {
	case len(raw) >= 32 && len(raw)%8 == 0:
		n := len(raw) / 8
		if n > 8 {
			n = 8
		}
		out := make([]uint64, n)
		for i := 0; i < n; i++ {
			out[i] = *(*uint64)(unsafe.Pointer(&raw[i*8]))
		}
		return out, nil
	case len(raw) >= 16:
		n := len(raw) / 4
		if n > 8 {
			n = 8
		}
		out := make([]uint64, n)
		for i := 0; i < n; i++ {
			out[i] = uint64(*(*uint32)(unsafe.Pointer(&raw[i*4])))
		}
		return out, nil
	default:
		return nil, unix.EINVAL
	}
}

// darwinCPUFromPS 汇总所有进程 %cpu（相对单核），再除以核心数得到整机大致利用率
func darwinCPUFromPS() (float64, bool) {
	out, err := exec.Command("ps", "-A", "-o", "%cpu=").Output()
	if err != nil {
		return 0, false
	}
	var sum float64
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		v, err := strconv.ParseFloat(line, 64)
		if err != nil {
			continue
		}
		sum += v
	}
	cores := float64(runtime.NumCPU())
	if cores <= 0 {
		return 0, false
	}
	usage := sum / cores
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	return usage, true
}

func sysctlPageCount(name string) uint64 {
	if v, err := unix.SysctlUint32(name); err == nil {
		return uint64(v)
	}
	if v, err := unix.SysctlUint64(name); err == nil {
		return v
	}
	return 0
}

func CollectMem() *dto.MemInfo {
	total, _ := unix.SysctlUint64("hw.memsize")
	pageSize := uint64(unix.Getpagesize())
	free := sysctlPageCount("vm.page_free_count")
	inactive := sysctlPageCount("vm.page_inactive_count")
	speculative := sysctlPageCount("vm.page_speculative_count")
	purgeable := sysctlPageCount("vm.page_purgeable_count")

	available := (free + inactive + speculative + purgeable) * pageSize
	if available > total {
		available = total
	}
	used := uint64(0)
	if total > available {
		used = total - available
	}
	return &dto.MemInfo{
		Total:     total,
		Used:      used,
		Available: available,
		UsagePct:  pct(used, total),
	}
}

func CollectDisk() []dto.DiskInfo {
	// macOS APFS 上多个路径常是同一卷，只展示根分区避免重复
	var st unix.Statfs_t
	if err := unix.Statfs("/", &st); err != nil {
		return nil
	}
	total := st.Blocks * uint64(st.Bsize)
	free := st.Bavail * uint64(st.Bsize)
	if total == 0 {
		return nil
	}
	used := total - free
	return []dto.DiskInfo{{
		Mount:    "/",
		FSType:   "local",
		Total:    total,
		Used:     used,
		Free:     free,
		UsagePct: pct(used, total),
	}}
}

func CollectSys() *dto.SysInfo {
	host, _ := os.Hostname()
	uptime := int64(0)
	if tv, err := unix.SysctlTimeval("kern.boottime"); err == nil && tv != nil {
		uptime = int64(time.Now().Unix() - int64(tv.Sec))
		if uptime < 0 {
			uptime = 0
		}
	}
	return &dto.SysInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: host,
		Uptime:   uptime,
	}
}
