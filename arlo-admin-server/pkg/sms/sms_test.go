package sms

import (
	"strings"
	"testing"

	"arlo-admin/internal/config"
)

func TestNewMock(t *testing.T) {
	s, err := New(&config.SMSConfig{Provider: "mock"})
	if err != nil || s.Name() != "mock" {
		t.Fatalf("mock: %v %v", s, err)
	}
}

func TestNewAliyunRequiresKeys(t *testing.T) {
	_, err := New(&config.SMSConfig{Provider: "aliyun"})
	if err == nil || !strings.Contains(err.Error(), "accessKey") {
		t.Fatalf("expected accessKey error, got %v", err)
	}
}

func TestNewTencentRequiresKeys(t *testing.T) {
	_, err := New(&config.SMSConfig{Provider: "tencent"})
	if err == nil || !strings.Contains(err.Error(), "secret") {
		t.Fatalf("expected secret error, got %v", err)
	}
}

func TestPercentEncode(t *testing.T) {
	if got := percentEncode("a b"); got != "a%20b" {
		t.Fatalf("got %s", got)
	}
}
