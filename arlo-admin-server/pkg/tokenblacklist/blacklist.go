package tokenblacklist

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"arlo-admin/internal/database"
	"arlo-admin/pkg/jwt"
	"arlo-admin/pkg/logger"

	"go.uber.org/zap"
)

const keyPrefix = "jwt:blacklist:"

// Add 将 token 加入黑名单，TTL 为剩余有效期（过期后自动清除）
// Redis 不可用时降级为空操作，不阻断登出流程
func Add(ctx context.Context, tokenString string, claims *jwt.Claims) error {
	if database.RDB == nil || claims == nil || tokenString == "" {
		return nil
	}
	ttl := remainingTTL(claims)
	if ttl <= 0 {
		return nil
	}
	key := keyPrefix + tokenKey(tokenString, claims)
	if err := database.RDB.Set(ctx, key, "1", ttl).Err(); err != nil {
		logger.Logger.Warn("token blacklist add failed", zap.Error(err))
		return err
	}
	return nil
}

// AddJTI 按 jti 写入黑名单（强制下线等场景，无完整 token 字符串时）
func AddJTI(ctx context.Context, jti string, ttl time.Duration) error {
	if database.RDB == nil || jti == "" {
		return nil
	}
	if ttl < time.Second {
		return nil
	}
	key := keyPrefix + jti
	if err := database.RDB.Set(ctx, key, "1", ttl).Err(); err != nil {
		logger.Logger.Warn("token blacklist add jti failed", zap.Error(err))
		return err
	}
	return nil
}

// IsBlacklisted 判断 token 是否已登出作废
// Redis 不可用时返回 false（降级：不拦业务，与现有 captcha/缓存策略一致）
func IsBlacklisted(ctx context.Context, tokenString string, claims *jwt.Claims) bool {
	if database.RDB == nil || claims == nil || tokenString == "" {
		return false
	}
	key := keyPrefix + tokenKey(tokenString, claims)
	n, err := database.RDB.Exists(ctx, key).Result()
	if err != nil {
		logger.Logger.Warn("token blacklist check failed", zap.Error(err))
		return false
	}
	return n > 0
}

// tokenKey 优先用 JWT ID（jti）；旧 token 无 jti 时回退为 token 哈希
func tokenKey(tokenString string, claims *jwt.Claims) string {
	if claims.ID != "" {
		return claims.ID
	}
	sum := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(sum[:])
}

func remainingTTL(claims *jwt.Claims) time.Duration {
	if claims.ExpiresAt == nil {
		return 0
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl < time.Second {
		return 0
	}
	return ttl
}
