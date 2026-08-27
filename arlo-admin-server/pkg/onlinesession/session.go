package onlinesession

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"arlo-admin/internal/config"
	"arlo-admin/internal/database"
	"arlo-admin/pkg/logger"
	"arlo-admin/pkg/tokenblacklist"

	"go.uber.org/zap"
)

const (
	sessionKeyPrefix = "jwt:online:"
	kickKeyPrefix    = "jwt:kick:"
)

// Session 在线会话信息
type Session struct {
	UserID     uint64 `json:"userId"`
	Username   string `json:"username"`
	AccessJTI  string `json:"accessJti"`
	RefreshJTI string `json:"refreshJti"`
	IP         string `json:"ip"`
	Browser    string `json:"browser"`
	OS         string `json:"os"`
	LoginAt    string `json:"loginAt"`
}

func sessionKey(userID uint64, refreshJTI string) string {
	return fmt.Sprintf("%s%d:%s", sessionKeyPrefix, userID, refreshJTI)
}

func kickKey(userID uint64) string {
	return fmt.Sprintf("%s%d", kickKeyPrefix, userID)
}

func sessionTTL() time.Duration {
	if config.GlobalConfig != nil && config.GlobalConfig.JWT.RefreshExpire > 0 {
		return config.GlobalConfig.JWT.RefreshExpire
	}
	return 7 * 24 * time.Hour
}

// Register 登录成功后登记在线会话（以 refreshJTI 为会话主键）
func Register(ctx context.Context, s Session) {
	if database.RDB == nil || s.UserID == 0 || s.RefreshJTI == "" {
		return
	}
	raw, err := json.Marshal(s)
	if err != nil {
		return
	}
	if err := database.RDB.Set(ctx, sessionKey(s.UserID, s.RefreshJTI), raw, sessionTTL()).Err(); err != nil {
		logger.Logger.Warn("online session register failed", zap.Error(err))
	}
}

// TouchAccess 刷新令牌后更新 accessJTI
func TouchAccess(ctx context.Context, userID uint64, refreshJTI, accessJTI string) {
	if database.RDB == nil || userID == 0 || refreshJTI == "" || accessJTI == "" {
		return
	}
	key := sessionKey(userID, refreshJTI)
	val, err := database.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return
	}
	var s Session
	if err := json.Unmarshal([]byte(val), &s); err != nil {
		return
	}
	s.AccessJTI = accessJTI
	raw, err := json.Marshal(s)
	if err != nil {
		return
	}
	_ = database.RDB.Set(ctx, key, raw, sessionTTL()).Err()
}

// UnregisterByAccess 登出时按 accessJTI 移除会话
func UnregisterByAccess(ctx context.Context, userID uint64, accessJTI string) {
	if database.RDB == nil || userID == 0 || accessJTI == "" {
		return
	}
	for _, s := range listByUser(ctx, userID) {
		if s.AccessJTI == accessJTI {
			_ = database.RDB.Del(ctx, sessionKey(userID, s.RefreshJTI)).Err()
			return
		}
	}
}

// UnregisterByRefresh 按 refreshJTI 移除会话
func UnregisterByRefresh(ctx context.Context, userID uint64, refreshJTI string) {
	if database.RDB == nil || userID == 0 || refreshJTI == "" {
		return
	}
	_ = database.RDB.Del(ctx, sessionKey(userID, refreshJTI)).Err()
}

// List 列出全部在线会话（SCAN，管理端规模可接受）
func List(ctx context.Context, username string) ([]Session, error) {
	if database.RDB == nil {
		return []Session{}, nil
	}
	var out []Session
	var cursor uint64
	pattern := sessionKeyPrefix + "*"
	for {
		keys, next, err := database.RDB.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			val, err := database.RDB.Get(ctx, key).Result()
			if err != nil || val == "" {
				continue
			}
			var s Session
			if err := json.Unmarshal([]byte(val), &s); err != nil {
				continue
			}
			if username != "" && !strings.Contains(strings.ToLower(s.Username), strings.ToLower(username)) {
				continue
			}
			out = append(out, s)
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return out, nil
}

func listByUser(ctx context.Context, userID uint64) []Session {
	list, err := List(ctx, "")
	if err != nil {
		return nil
	}
	var out []Session
	for _, s := range list {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	return out
}

// KickSession 强制下线单个会话（按 refreshJTI）
func KickSession(ctx context.Context, userID uint64, refreshJTI string) error {
	if database.RDB == nil {
		return fmt.Errorf("redis 未就绪")
	}
	key := sessionKey(userID, refreshJTI)
	val, err := database.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return fmt.Errorf("会话不存在或已下线")
	}
	var s Session
	if err := json.Unmarshal([]byte(val), &s); err != nil {
		return err
	}
	ttl := sessionTTL()
	if s.AccessJTI != "" {
		_ = tokenblacklist.AddJTI(ctx, s.AccessJTI, ttl)
	}
	if s.RefreshJTI != "" {
		_ = tokenblacklist.AddJTI(ctx, s.RefreshJTI, ttl)
	}
	_ = database.RDB.Del(ctx, key).Err()
	return nil
}

// KickUser 强制下线该用户全部会话（踢出标记 + 清理 online key）
func KickUser(ctx context.Context, userID uint64) error {
	if database.RDB == nil {
		return fmt.Errorf("redis 未就绪")
	}
	ttl := sessionTTL()
	sessions := listByUser(ctx, userID)
	for _, s := range sessions {
		if s.AccessJTI != "" {
			_ = tokenblacklist.AddJTI(ctx, s.AccessJTI, ttl)
		}
		if s.RefreshJTI != "" {
			_ = tokenblacklist.AddJTI(ctx, s.RefreshJTI, ttl)
		}
		_ = database.RDB.Del(ctx, sessionKey(userID, s.RefreshJTI)).Err()
	}
	// 踢出时间戳：作废该时刻之前签发的一切 token（含未登记的旧会话）
	now := strconv.FormatInt(time.Now().Unix(), 10)
	if err := database.RDB.Set(ctx, kickKey(userID), now, ttl).Err(); err != nil {
		return err
	}
	return nil
}

// IsKicked 判断 token 签发时间是否早于强制下线时间
func IsKicked(ctx context.Context, userID uint64, issuedAt time.Time) bool {
	if database.RDB == nil || userID == 0 || issuedAt.IsZero() {
		return false
	}
	val, err := database.RDB.Get(ctx, kickKey(userID)).Result()
	if err != nil || val == "" {
		return false
	}
	ts, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return false
	}
	return issuedAt.Unix() < ts
}
