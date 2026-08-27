package captcha

import (
	"context"
	"errors"
	"time"

	"arlo-admin/internal/database"

	"github.com/mojocn/base64Captcha"
)

const (
	prefix = "captcha:"
	ttl    = 5 * time.Minute
)

// ErrStoreUnavailable Redis 未初始化或写入失败（避免对 nil client 调用导致 panic）
var ErrStoreUnavailable = errors.New("验证码服务不可用：Redis 未就绪")

// driver 数字验证码驱动：80x240，4位数字
var driver = base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)

// store Redis 存储（自动过期 + 一次性消费）
var store = newRedisStore()

// ── Redis 存储适配 base64Captcha.Store 接口 ──────────────────────────

type redisStore struct{}

func newRedisStore() *redisStore { return &redisStore{} }

func (s *redisStore) Set(id string, value string) error {
	if database.RDB == nil {
		return ErrStoreUnavailable
	}
	if err := database.RDB.Set(context.Background(), prefix+id, value, ttl).Err(); err != nil {
		return errors.Join(ErrStoreUnavailable, err)
	}
	return nil
}

func (s *redisStore) Get(id string, clear bool) (val string) {
	if database.RDB == nil {
		return ""
	}
	ctx := context.Background()
	key := prefix + id
	v, err := database.RDB.Get(ctx, key).Result()
	if err != nil {
		return ""
	}
	if clear {
		database.RDB.Del(ctx, key)
	}
	return v
}

func (s *redisStore) Verify(id string, answer string, clear bool) bool {
	return s.Get(id, clear) == answer
}

// ── 公开 API ────────────────────────────────────────────────────────

// Generate 生成验证码，返回 id + base64 图片
func Generate() (id string, b64s string, err error) {
	if database.RDB == nil {
		return "", "", ErrStoreUnavailable
	}
	c := base64Captcha.NewCaptcha(driver, store)
	id, b64s, _, err = c.Generate()
	return
}

// Verify 校验验证码（一次性消费：校验通过后自动删除）
func Verify(id string, code string) bool {
	return store.Verify(id, code, true)
}
