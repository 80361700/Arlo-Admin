package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/sysconfig/dto"
	"arlo-admin/internal/modules/sysconfig/model"
	"arlo-admin/internal/modules/sysconfig/repository"
	"arlo-admin/pkg/utils"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 公开可读的配置键（登录页等免鉴权场景）
const (
	KeySysName    = "sys.name"
	KeySysCaptcha = "sys.captcha"
	KeySysLogo    = "sys.logo"
	KeySysVersion = "sys.version"
)

// 账号安全配置键
const (
	KeyLoginMaxRetry     = "sys.login.max_retry"
	KeyLoginLockMinutes  = "sys.login.lock_minutes"
	KeyPwdMinLength      = "sys.pwd.min_length"
	KeyPwdComplexity     = "sys.pwd.require_complexity"
	KeyPwdExpireDays     = "sys.pwd.expire_days"
	KeyInitPwd           = "sys.init_pwd"
)

const cacheKeyPrefix = "sys:config:"
const cacheTTL = 30 * time.Minute

// ConfigService 配置服务
type ConfigService struct {
	repo *repository.ConfigRepository
}

// NewConfigService 创建 ConfigService
func NewConfigService(repo *repository.ConfigRepository) *ConfigService {
	return &ConfigService{repo: repo}
}

// List 配置列表
func (s *ConfigService) List(ctx context.Context, req *dto.ConfigListQuery) ([]dto.ConfigResponse, error) {
	query := map[string]interface{}{
		"name": req.Name,
		"key":  req.Key,
		"type": req.Type,
	}
	configs, err := s.repo.FindAll(ctx, query)
	if err != nil {
		return nil, err
	}

	list := make([]dto.ConfigResponse, 0, len(configs))
	for _, c := range configs {
		list = append(list, toResponse(&c))
	}
	return list, nil
}

// GetByKey 按 key 获取配置（Redis 缓存）
func (s *ConfigService) GetByKey(ctx context.Context, key string) (*dto.ConfigResponse, error) {
	// 1. 先查 Redis 缓存
	cacheKey := cacheKeyPrefix + key
	if database.RDB != nil {
		cached, err := database.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			var resp dto.ConfigResponse
			if json.Unmarshal([]byte(cached), &resp) == nil {
				return &resp, nil
			}
		} else if !errors.Is(err, redis.Nil) {
			// Redis 报错不阻塞，降级查 DB
		}
	}

	// 2. 穿透查 DB
	config, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("配置不存在: %s", key)
		}
		return nil, err
	}

	resp := toResponse(config)

	// 3. 回填缓存
	if database.RDB != nil {
		data, _ := json.Marshal(resp)
		database.RDB.Set(ctx, cacheKey, data, cacheTTL)
	}

	return &resp, nil
}

// GetPublic 获取公开配置（免鉴权）
func (s *ConfigService) GetPublic(ctx context.Context) *dto.PublicConfigResponse {
	resp := &dto.PublicConfigResponse{
		Name:    "Arlo Admin",
		Captcha: true, // 缺省开启，更安全
		Logo:    "",
		Version: "1.0.0",
	}
	if cfg, err := s.GetByKey(ctx, KeySysName); err == nil && cfg.Value != "" {
		resp.Name = cfg.Value
	}
	if cfg, err := s.GetByKey(ctx, KeySysCaptcha); err == nil {
		resp.Captcha = IsTruthy(cfg.Value)
	}
	if cfg, err := s.GetByKey(ctx, KeySysLogo); err == nil {
		resp.Logo = cfg.Value
	}
	if cfg, err := s.GetByKey(ctx, KeySysVersion); err == nil && cfg.Value != "" {
		resp.Version = cfg.Value
	}
	return resp
}

// IsCaptchaEnabled 登录验证码是否开启
func (s *ConfigService) IsCaptchaEnabled(ctx context.Context) bool {
	cfg, err := s.GetByKey(ctx, KeySysCaptcha)
	if err != nil {
		return true
	}
	return IsTruthy(cfg.Value)
}

// GetInt 读取整数配置，失败或不存在时返回默认值
func (s *ConfigService) GetInt(ctx context.Context, key string, def int) int {
	cfg, err := s.GetByKey(ctx, key)
	if err != nil || cfg.Value == "" {
		return def
	}
	n, err := strconv.Atoi(strings.TrimSpace(cfg.Value))
	if err != nil {
		return def
	}
	return n
}

// GetBool 读取开关配置
func (s *ConfigService) GetBool(ctx context.Context, key string, def bool) bool {
	cfg, err := s.GetByKey(ctx, key)
	if err != nil {
		return def
	}
	return IsTruthy(cfg.Value)
}

// GetString 读取字符串配置
func (s *ConfigService) GetString(ctx context.Context, key string, def string) string {
	cfg, err := s.GetByKey(ctx, key)
	if err != nil || cfg.Value == "" {
		return def
	}
	return cfg.Value
}

// LoginMaxRetry 登录最大失败次数（0=不限制）
func (s *ConfigService) LoginMaxRetry(ctx context.Context) int {
	return s.GetInt(ctx, KeyLoginMaxRetry, 5)
}

// LoginLockMinutes 锁定分钟数
func (s *ConfigService) LoginLockMinutes(ctx context.Context) int {
	return s.GetInt(ctx, KeyLoginLockMinutes, 30)
}

// PwdMinLength 密码最小长度
func (s *ConfigService) PwdMinLength(ctx context.Context) int {
	return s.GetInt(ctx, KeyPwdMinLength, 6)
}

// PwdRequireComplexity 是否要求字母+数字
func (s *ConfigService) PwdRequireComplexity(ctx context.Context) bool {
	return s.GetBool(ctx, KeyPwdComplexity, false)
}

// PwdExpireDays 密码有效天数（0=永不过期）
func (s *ConfigService) PwdExpireDays(ctx context.Context) int {
	return s.GetInt(ctx, KeyPwdExpireDays, 0)
}

// IsTruthy 解析开关类配置值
func IsTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true", "1", "on", "yes", "开启", "开":
		return true
	default:
		return false
	}
}

// normalizeValue 按类型规范化并校验配置值
func normalizeValue(typ int8, value string) (string, error) {
	value = strings.TrimSpace(value)
	switch typ {
	case 2: // JSON
		if value == "" {
			return "{}", nil
		}
		if !json.Valid([]byte(value)) {
			return "", fmt.Errorf("JSON 格式不正确")
		}
		return value, nil
	case 3: // 开关
		if value == "" {
			return "false", nil
		}
		if IsTruthy(value) {
			return "true", nil
		}
		return "false", nil
	case 4: // 图片：允许空
		return value, nil
	default: // 文本
		if value == "" {
			return "", fmt.Errorf("配置值不能为空")
		}
		return value, nil
	}
}

// Create 创建配置
func (s *ConfigService) Create(ctx context.Context, req *dto.CreateConfigRequest) (*dto.ConfigResponse, error) {
	exists, err := s.repo.ExistsByKey(ctx, req.Key, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("配置键 %s 已存在", req.Key)
	}

	value, err := normalizeValue(req.Type, req.Value)
	if err != nil {
		return nil, err
	}

	config := &model.SysConfig{
		Name:   req.Name,
		Key:    req.Key,
		Value:  value,
		Type:   req.Type,
		Remark: req.Remark,
	}
	if err := s.repo.Create(ctx, config); err != nil {
		return nil, err
	}

	resp := toResponse(config)
	return &resp, nil
}

// Update 更新配置
func (s *ConfigService) Update(ctx context.Context, req *dto.UpdateConfigRequest) (*dto.ConfigResponse, error) {
	config, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("配置不存在")
		}
		return nil, err
	}

	oldKey := config.Key
	if req.Key != oldKey {
		exists, err := s.repo.ExistsByKey(ctx, req.Key, req.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("配置键 %s 已存在", req.Key)
		}
	}

	value, err := normalizeValue(req.Type, req.Value)
	if err != nil {
		return nil, err
	}

	config.Name = req.Name
	config.Key = req.Key
	config.Value = value
	config.Type = req.Type
	config.Remark = req.Remark
	if err := s.repo.Update(ctx, config); err != nil {
		return nil, err
	}

	// 更新后清缓存（键变更时旧键也要清）
	s.invalidateCache(ctx, oldKey)
	if req.Key != oldKey {
		s.invalidateCache(ctx, req.Key)
	}

	resp := toResponse(config)
	return &resp, nil
}

// Delete 删除配置
func (s *ConfigService) Delete(ctx context.Context, id uint64) error {
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("配置不存在")
		}
		return err
	}

	s.invalidateCache(ctx, config.Key)
	return s.repo.Delete(ctx, id)
}

// invalidateCache 清除 Redis 缓存
func (s *ConfigService) invalidateCache(ctx context.Context, key string) {
	if database.RDB != nil {
		database.RDB.Del(ctx, cacheKeyPrefix+key)
	}
}

// toResponse model → dto
func toResponse(c *model.SysConfig) dto.ConfigResponse {
	return dto.ConfigResponse{
		ID:        c.ID,
		Name:      c.Name,
		Key:       c.Key,
		Value:     c.Value,
		Type:      c.Type,
		Remark:    c.Remark,
		CreatedAt: utils.FormatTime(c.CreatedAt),
		UpdatedAt: utils.FormatTime(c.UpdatedAt),
	}
}
