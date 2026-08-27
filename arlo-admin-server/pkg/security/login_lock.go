package security

import (
	"context"
	"fmt"
	"time"

	"arlo-admin/internal/database"

	"github.com/redis/go-redis/v9"
)

const (
	failKeyPrefix = "login:fail:"
	lockKeyPrefix = "login:lock:"
)

// IsLocked 账号是否处于锁定中；若锁定返回剩余秒数
func IsLocked(ctx context.Context, username string) (bool, time.Duration) {
	if database.RDB == nil || username == "" {
		return false, 0
	}
	ttl, err := database.RDB.TTL(ctx, lockKeyPrefix+username).Result()
	if err != nil || ttl <= 0 {
		return false, 0
	}
	return true, ttl
}

// RecordLoginFail 记录一次失败；达到阈值则锁定。返回是否已锁定、剩余可试次数说明
func RecordLoginFail(ctx context.Context, username string, maxRetry, lockMinutes int) (locked bool, msg string) {
	if database.RDB == nil || username == "" || maxRetry <= 0 {
		return false, ""
	}
	if lockMinutes <= 0 {
		lockMinutes = 30
	}
	failKey := failKeyPrefix + username
	n, err := database.RDB.Incr(ctx, failKey).Result()
	if err != nil {
		return false, ""
	}
	// 失败计数窗口与锁定时长一致
	_ = database.RDB.Expire(ctx, failKey, time.Duration(lockMinutes)*time.Minute).Err()

	remain := maxRetry - int(n)
	if n >= int64(maxRetry) {
		_ = database.RDB.Set(ctx, lockKeyPrefix+username, "1", time.Duration(lockMinutes)*time.Minute).Err()
		_ = database.RDB.Del(ctx, failKey).Err()
		return true, fmt.Sprintf("登录失败次数过多，账号已锁定 %d 分钟", lockMinutes)
	}
	if remain > 0 {
		return false, fmt.Sprintf("用户名或密码错误，还可尝试 %d 次", remain)
	}
	return false, "用户名或密码错误"
}

// ClearLoginFail 登录成功后清除失败计数与锁定
func ClearLoginFail(ctx context.Context, username string) {
	if database.RDB == nil || username == "" {
		return
	}
	database.RDB.Del(ctx, failKeyPrefix+username, lockKeyPrefix+username)
}

// UnlockAccount 管理员解锁（可选扩展）
func UnlockAccount(ctx context.Context, username string) error {
	if database.RDB == nil {
		return redis.Nil
	}
	return database.RDB.Del(ctx, failKeyPrefix+username, lockKeyPrefix+username).Err()
}
