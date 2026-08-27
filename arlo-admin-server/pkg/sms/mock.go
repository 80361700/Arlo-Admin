package sms

import (
	"context"

	"arlo-admin/pkg/logger"

	"go.uber.org/zap"
)

// MockSender 开发用：不调网关，仅打日志
type MockSender struct{}

func (s *MockSender) Name() string { return "mock" }

func (s *MockSender) Send(_ context.Context, phone, code string) error {
	logger.Logger.Info("sms mock send",
		zap.String("phone", phone),
		zap.String("code", code),
	)
	return nil
}
