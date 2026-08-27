package middleware

import (
	"strings"

	"arlo-admin/pkg/jwt"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/onlinesession"
	"arlo-admin/pkg/response"
	"arlo-admin/pkg/tokenblacklist"

	"github.com/gin-gonic/gin"
)

func extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

func applyAccessClaims(c *gin.Context, tokenString string, claims *jwt.Claims) bool {
	if claims.Subject != "" && claims.Subject != "access" {
		return false
	}
	if tokenblacklist.IsBlacklisted(c.Request.Context(), tokenString, claims) {
		return false
	}
	if claims.IssuedAt != nil && onlinesession.IsKicked(c.Request.Context(), claims.UserID, claims.IssuedAt.Time) {
		return false
	}
	c.Set("userID", claims.UserID)
	c.Set("username", claims.Username)
	c.Set("accessToken", tokenString)
	return true
}

// JWTAuth JWT 认证中间件
// 从 Authorization: Bearer 提取并校验 access token
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractBearerToken(c)
		if tokenString == "" {
			response.Error(c, apperrors.Unauthorized, "未提供认证令牌")
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			response.Error(c, apperrors.Unauthorized, err.Error())
			c.Abort()
			return
		}

		if !applyAccessClaims(c, tokenString, claims) {
			response.Error(c, apperrors.Unauthorized, "登录已失效，请重新登录")
			c.Abort()
			return
		}

		c.Next()
	}
}

// TryJWTAuth 尝试解析 JWT（不强制）。成功时注入用户上下文并返回 true。
func TryJWTAuth(c *gin.Context) bool {
	tokenString := extractBearerToken(c)
	if tokenString == "" {
		return false
	}
	claims, err := jwt.ParseToken(tokenString)
	if err != nil {
		return false
	}
	return applyAccessClaims(c, tokenString, claims)
}

// GetCurrentUser 从 Gin 上下文中获取当前登录用户信息
// 返回 userID 和 username，如果未认证则返回 0 和空字符串
func GetCurrentUser(c *gin.Context) (uint64, string) {
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	var uid uint64
	if v, ok := userID.(uint64); ok {
		uid = v
	}
	var uname string
	if v, ok := username.(string); ok {
		uname = v
	}
	return uid, uname
}

// MemberAuth 会员 JWT 认证中间件
// 从 Authorization Header 提取 Bearer token，验证会员身份后将 memberID 和 phone 注入上下文
func MemberAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, apperrors.Unauthorized, "未提供认证令牌")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, apperrors.Unauthorized, "认证令牌格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := jwt.ParseMemberToken(tokenString)
		if err != nil {
			response.Error(c, apperrors.Unauthorized, err.Error())
			c.Abort()
			return
		}

		if claims.Subject != "" && claims.Subject != "member-access" {
			response.Error(c, apperrors.Unauthorized, "非法的访问令牌类型")
			c.Abort()
			return
		}

		c.Set("memberID", claims.MemberID)
		c.Set("phone", claims.Phone)

		c.Next()
	}
}

// GetCurrentMember 从 Gin 上下文中获取当前登录会员信息
func GetCurrentMember(c *gin.Context) (uint64, string) {
	memberID, _ := c.Get("memberID")
	phone, _ := c.Get("phone")

	var mid uint64
	if v, ok := memberID.(uint64); ok {
		mid = v
	}
	var p string
	if v, ok := phone.(string); ok {
		p = v
	}
	return mid, p
}
