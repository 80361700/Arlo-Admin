//go:build linux

package collector

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"arlo-admin/internal/modules/monitor/dto"

	"golang.org/x/sys/unix"
)

func CollectCPU() *dto.CPUInfo {
	info := &dto.CPUInfo{Cores: runtime.NumCPU()}
	info.Load1, info.Load5, info.Load15 = linuxLoadAvg()

	t1, i1, err1 := linuxCPUTimes()
	time.Sleep(300 * time.Millisecond)
	t2, i2, err2 := linuxCPUTimes()
	if err1 == nil && err2 == nil && t2 > t1 {
		idleDelta := float64(i2 - i1)
		totalDelta := float64(t2 - t1)
		if totalDelta > 0 {
			info.UsagePct = (1 - idleDelta/totalDelta) * 100
			if info.UsagePct < 0 {
				info.UsagePct = 0
			}
			if info.UsagePct > 100 {
				info.UsagePct = 100
			}
		}
	}
	return info
}

func linuxLoadAvg() (l1, l5, l15 float64) {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return
	}
	fields := strings.Fields(string(b))
	if len(fields) < 3 {
		return
	}
	l1, _ = strconv.ParseFloat(fields[0], 64)
	l5, _ = strconv.ParseFloat(fields[1], 64)
	l15, _ = strconv.ParseFloat(fields[2], 64)
	return
}

func linuxCPUTimes() (total, idle uint64, err error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		return 0, 0, os.ErrInvalid
	}
	fields := strings.Fields(sc.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0, os.ErrInvalid
	}
	var vals []uint64
	for _, s := range fields[1:] {
		v, e := strconv.ParseUint(s, 10, 64)
		if e != nil {
			continue
		}
		vals = append(vals, v)
		total += v
	}
	if len(vals) > 3 {
		idle = vals[3]
	}
	return total, idle, nil
}

func CollectMem() *dto.MemInfo {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return &dto.MemInfo{}
	}
	defer f.Close()

	kv := map[string]uint64{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if len(fields) == 0 {
			continue
		}
		v, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			continue
		}
		kv[key] = v * 1024 // kB → bytes
	}

	total := kv["MemTotal"]
	available := kv["MemAvailable"]
	if available == 0 {
		available = kv["MemFree"] + kv["Buffers"] + kv["Cached"]
	}
	if available > total {
		available = total
	}
	used := total - available
	return &dto.MemInfo{
		Total:     total,
		Used:      used,
		Available: available,
		UsagePct:  pct(used, total),
	}
}

func CollectDisk() []dto.DiskInfo {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return collectDiskMounts([]string{"/"})
	}
	defer f.Close()

	var mounts []string
	seenDev := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 3 {
			continue
		}
		dev, mount, fs := fields[0], fields[1], fields[2]
		if !strings.HasPrefix(dev, "/dev/") {
			continue
		}
		switch fs {
		case "ext4", "ext3", "xfs", "btrfs", "vfat", "ntfs", "fuseblk":
		default:
			continue
		}
		if seenDev[dev] {
			continue
		}
		seenDev[dev] = true
		mounts = append(mounts, mount)
	}
	if len(mounts) == 0 {
		mounts = []string{"/"}
	}
	return collectDiskMounts(mounts)
}

func collectDiskMounts(mounts []string) []dto.DiskInfo {
	var out []dto.DiskInfo
	seen := map[string]bool{}
	for _, m := range mounts {
		var st unix.Statfs_t
		if err := unix.Statfs(m, &st); err != nil {
			continue
		}
		total := st.Blocks * uint64(st.Bsize)
		free := st.Bavail * uint64(st.Bsize)
		if total == 0 || seen[m] {
			continue
		}
		seen[m] = true
		used := total - free
		out = append(out, dto.DiskInfo{
			Mount:    m,
			FSType:   "local",
			Total:    total,
			Used:     used,
			Free:     free,
			UsagePct: pct(used, total),
		})
	}
	return out
}

func CollectSys() *dto.SysInfo {
	host, _ := os.Hostname()
	uptime := int64(0)
	if b, err := os.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(b))
		if len(fields) > 0 {
			if v, err := strconv.ParseFloat(fields[0], 64); err == nil {
				uptime = int64(v)
			}
		}
	}
	return &dto.SysInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: host,
		Uptime:   uptime,
	}
}
