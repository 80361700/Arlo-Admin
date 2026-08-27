package dto

// ServerMonitorResponse 服务监控总览
type ServerMonitorResponse struct {
	CPU   *CPUInfo    `json:"cpu"`
	Mem   *MemInfo    `json:"mem"`
	Disk  []DiskInfo  `json:"disk"`
	Sys   *SysInfo    `json:"sys"`
	Go    *GoInfo     `json:"go"`
	App   *AppInfo    `json:"app"`
	DB    *DBInfo     `json:"db"`
	Redis *RedisInfo  `json:"redis"`
}

type CPUInfo struct {
	Cores     int       `json:"cores"`
	UsagePct  float64   `json:"usagePct"`  // 0-100，采样约 300ms
	Load1     float64   `json:"load1"`
	Load5     float64   `json:"load5"`
	Load15    float64   `json:"load15"`
}

type MemInfo struct {
	Total     uint64  `json:"total"`     // bytes
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	UsagePct  float64 `json:"usagePct"`
}

type DiskInfo struct {
	Mount     string  `json:"mount"`
	FSType    string  `json:"fsType"`
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Free      uint64  `json:"free"`
	UsagePct  float64 `json:"usagePct"`
}

type SysInfo struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Hostname string `json:"hostname"`
	Uptime   int64  `json:"uptime"` // 系统启动以来秒数，拿不到则为 0
}

type GoInfo struct {
	Version        string  `json:"version"`
	Goroutines     int     `json:"goroutines"`
	GOMAXPROCS     int     `json:"gomaxprocs"`
	HeapAlloc      uint64  `json:"heapAlloc"`
	HeapSys        uint64  `json:"heapSys"`
	HeapInuse      uint64  `json:"heapInuse"`
	GCCPUFraction  float64 `json:"gcCPUFraction"`
	NumGC          uint32  `json:"numGC"`
	LastGC         string  `json:"lastGC"` // 格式化时间，无则空
}

type AppInfo struct {
	Name      string `json:"name"`
	Mode      string `json:"mode"`
	StartTime string `json:"startTime"`
	RunSeconds int64 `json:"runSeconds"`
	GoVersion string `json:"goVersion"`
}

type DBInfo struct {
	Status  string  `json:"status"` // up / down / unknown
	PingMs  float64 `json:"pingMs"` // 支持亚毫秒
	Open    int     `json:"open"`
	InUse   int     `json:"inUse"`
	Idle    int     `json:"idle"`
	MaxOpen int     `json:"maxOpen"`
}

type RedisInfo struct {
	Status string  `json:"status"`
	PingMs float64 `json:"pingMs"`
}
