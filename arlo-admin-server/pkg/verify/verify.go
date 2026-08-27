package verify

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"arlo-admin/internal/config"
	"arlo-admin/internal/database"
	"arlo-admin/pkg/logger"
	"arlo-admin/pkg/sms"

	"go.uber.org/zap"
)

const (
	prefix      = "sms:code:"
	defaultTTL  = 5 * time.Minute
	defaultWait = 60 * time.Second
	codeLen     = 6
)

// SendCode 生成验证码写入 Redis，并按配置渠道发送短信
func SendCode(ctx context.Context, phone string) error {
	if database.RDB == nil {
		return fmt.Errorf("Redis 未就绪，无法发送验证码")
	}

	ttl, resend := codePolicy()

	lockKey := prefix + phone + ":lock"
	exists, err := database.RDB.Exists(ctx, lockKey).Result()
	if err != nil {
		return fmt.Errorf("检查发送频率失败: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("验证码发送过于频繁，请 %d 秒后再试", int(resend/time.Second))
	}

	code, err := randomDigits(codeLen)
	if err != nil {
		return fmt.Errorf("生成验证码失败: %w", err)
	}

	codeKey := prefix + phone
	if err := database.RDB.Set(ctx, codeKey, code, ttl).Err(); err != nil {
		return fmt.Errorf("验证码存储失败: %w", err)
	}
	if err := database.RDB.Set(ctx, lockKey, "1", resend).Err(); err != nil {
		_ = database.RDB.Del(ctx, codeKey).Err()
		return fmt.Errorf("设置发送间隔失败: %w", err)
	}

	if err := sms.Send(ctx, phone, code); err != nil {
		_ = database.RDB.Del(ctx, codeKey, lockKey).Err()
		logger.Logger.Error("短信发送失败", zap.String("phone", phone), zap.Error(err))
		return fmt.Errorf("验证码发送失败: %w", err)
	}

	logger.Logger.Info("验证码已发送",
		zap.String("phone", phone),
		zap.Duration("ttl", ttl),
	)
	return nil
}

// VerifyCode 校验手机验证码，验证通过后删除
func VerifyCode(ctx context.Context, phone string, code string) bool {
	if database.RDB == nil || code == "" {
		return false
	}
	codeKey := prefix + phone
	stored, err := database.RDB.Get(ctx, codeKey).Result()
	if err != nil {
		return false
	}
	if stored == code {
		database.RDB.Del(ctx, codeKey)
		return true
	}
	return false
}

func codePolicy() (ttl, resend time.Duration) {
	ttl, resend = defaultTTL, defaultWait
	if config.GlobalConfig == nil {
		return
	}
	cfg := config.GlobalConfig.SMS
	if cfg.CodeTTL > 0 {
		ttl = cfg.CodeTTL
	}
	if cfg.ResendInterval > 0 {
		resend = cfg.ResendInterval
	}
	return
}

func randomDigits(n int) (string, error) {
	max := big.NewInt(10)
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		v, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = byte('0' + v.Int64())
	}
	return string(out), nil
}
