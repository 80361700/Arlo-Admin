package sms

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"arlo-admin/internal/config"
	"arlo-admin/pkg/logger"

	"go.uber.org/zap"
)

// Sender 短信发送器
type Sender interface {
	Send(ctx context.Context, phone, code string) error
	Name() string
}

var (
	mu      sync.RWMutex
	current Sender = &MockSender{}
)

// Init 按配置初始化全局发送器
func Init(cfg *config.SMSConfig) error {
	if cfg == nil {
		cfg = &config.SMSConfig{Provider: "mock"}
	}
	s, err := New(cfg)
	if err != nil {
		return err
	}
	mu.Lock()
	current = s
	mu.Unlock()
	logger.Logger.Info("sms sender ready", zap.String("provider", s.Name()))
	return nil
}

// Send 使用当前配置的渠道发送验证码短信
func Send(ctx context.Context, phone, code string) error {
	mu.RLock()
	s := current
	mu.RUnlock()
	if s == nil {
		return fmt.Errorf("短信发送器未初始化")
	}
	return s.Send(ctx, phone, code)
}

// New 根据 provider 创建发送器
func New(cfg *config.SMSConfig) (Sender, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	switch provider {
	case "", "mock", "log", "dev":
		return &MockSender{}, nil
	case "aliyun", "ali", "dysms":
		return newAliyunSender(&cfg.Aliyun)
	case "tencent", "qcloud", "tc":
		return newTencentSender(&cfg.Tencent)
	default:
		return nil, fmt.Errorf("不支持的短信渠道: %s（可用 mock / aliyun / tencent）", cfg.Provider)
	}
}
