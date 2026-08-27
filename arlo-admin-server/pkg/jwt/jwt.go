package jwt

import (
	"errors"
	"time"

	"arlo-admin/internal/config"

	"github.com/google/uuid"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token已过期")
	ErrTokenInvalid = errors.New("token无效")
)

// Claims 管理员 JWT Claims
type Claims struct {
	UserID   uint64 `json:"userId"`
	Username string `json:"username"`
	jwtv5.RegisteredClaims
}

// MemberClaims C端会员 JWT Claims
type MemberClaims struct {
	MemberID uint64 `json:"memberId"`
	Phone    string `json:"phone"`
	jwtv5.RegisteredClaims
}

// GenerateAccessToken 生成访问令牌（短期）
func GenerateAccessToken(userID uint64, username string) (string, int64, error) {
	cfg := config.GlobalConfig
	now := time.Now()
	expiresAt := now.Add(cfg.JWT.AccessExpire)

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ID:        uuid.NewString(), // jti，供登出黑名单精确作废
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
			IssuedAt:  jwtv5.NewNumericDate(now),
			Issuer:    cfg.JWT.Issuer,
			Subject:   "access",
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return "", 0, err
	}

	return signedToken, int64(cfg.JWT.AccessExpire.Seconds()), nil
}

// GenerateRefreshToken 生成刷新令牌（长期）
func GenerateRefreshToken(userID uint64, username string) (string, error) {
	cfg := config.GlobalConfig
	now := time.Now()
	expiresAt := now.Add(cfg.JWT.RefreshExpire)

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
			IssuedAt:  jwtv5.NewNumericDate(now),
			Issuer:    cfg.JWT.Issuer,
			Subject:   "refresh",
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ParseToken 解析并验证管理员 Token
// 返回 Claims；如果 token 过期或无效则返回 error
func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.GlobalConfig

	token, err := jwtv5.ParseWithClaims(tokenString, &Claims{}, func(t *jwtv5.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := t.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwtv5.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// GenerateMemberAccessToken 生成会员访问令牌（短期）
func GenerateMemberAccessToken(memberID uint64, phone string) (string, int64, error) {
	cfg := config.GlobalConfig
	now := time.Now()
	expiresAt := now.Add(cfg.JWT.AccessExpire)

	claims := MemberClaims{
		MemberID: memberID,
		Phone:    phone,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
			IssuedAt:  jwtv5.NewNumericDate(now),
			Issuer:    cfg.JWT.Issuer,
			Subject:   "member-access",
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return "", 0, err
	}

	return signedToken, int64(cfg.JWT.AccessExpire.Seconds()), nil
}

// GenerateMemberRefreshToken 生成会员刷新令牌（长期）
func GenerateMemberRefreshToken(memberID uint64, phone string) (string, error) {
	cfg := config.GlobalConfig
	now := time.Now()
	expiresAt := now.Add(cfg.JWT.RefreshExpire)

	claims := MemberClaims{
		MemberID: memberID,
		Phone:    phone,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
			IssuedAt:  jwtv5.NewNumericDate(now),
			Issuer:    cfg.JWT.Issuer,
			Subject:   "member-refresh",
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ParseMemberToken 解析并验证会员 Token
func ParseMemberToken(tokenString string) (*MemberClaims, error) {
	cfg := config.GlobalConfig

	token, err := jwtv5.ParseWithClaims(tokenString, &MemberClaims{}, func(t *jwtv5.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwtv5.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*MemberClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}
