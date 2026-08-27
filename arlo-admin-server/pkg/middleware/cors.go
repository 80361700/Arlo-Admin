package middleware

import (
	"net/http"
	"strings"

	"arlo-admin/internal/config"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowOrigin := resolveAllowOrigin(origin)
		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			if allowOrigin != "*" {
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func resolveAllowOrigin(requestOrigin string) string {
	origins := []string{}
	if config.GlobalConfig != nil {
		origins = config.GlobalConfig.Server.CorsOrigins
	}
	// 未配置：开发兼容 *
	if len(origins) == 0 {
		return "*"
	}
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if o == "*" {
			return "*"
		}
		if requestOrigin != "" && strings.EqualFold(o, requestOrigin) {
			return requestOrigin
		}
	}
	// 有白名单但不匹配：不回写 Allow-Origin（浏览器拦截）
	return ""
}
