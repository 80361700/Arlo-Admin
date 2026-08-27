package collector

import (
	"context"
	"math"
	"time"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/monitor/dto"
)

func pingMs(start time.Time) float64 {
	us := float64(time.Since(start).Microseconds())
	ms := us / 1000
	// 保留 2 位小数，避免本地极快 ping 显示成 0
	return math.Round(ms*100) / 100
}

// CollectDB MySQL 健康与连接池
func CollectDB(ctx context.Context) *dto.DBInfo {
	info := &dto.DBInfo{Status: "unknown"}
	if database.DB == nil {
		info.Status = "down"
		return info
	}
	sqlDB, err := database.DB.DB()
	if err != nil {
		info.Status = "down"
		return info
	}
	start := time.Now()
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		info.Status = "down"
		info.PingMs = pingMs(start)
		return info
	}
	stats := sqlDB.Stats()
	info.Status = "up"
	info.PingMs = pingMs(start)
	info.Open = stats.OpenConnections
	info.InUse = stats.InUse
	info.Idle = stats.Idle
	info.MaxOpen = stats.MaxOpenConnections
	return info
}

// CollectRedis Redis 健康
func CollectRedis(ctx context.Context) *dto.RedisInfo {
	info := &dto.RedisInfo{Status: "unknown"}
	if database.RDB == nil {
		info.Status = "down"
		return info
	}
	start := time.Now()
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := database.RDB.Ping(pingCtx).Err(); err != nil {
		info.Status = "down"
		info.PingMs = pingMs(start)
		return info
	}
	info.Status = "up"
	info.PingMs = pingMs(start)
	return info
}
