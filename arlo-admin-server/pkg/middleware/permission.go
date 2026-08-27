package middleware

import (
	"fmt"

	casbinpkg "arlo-admin/pkg/casbin"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// CasbinAuth Casbin 权限校验中间件
// 在 JWTAuth 之后使用，检查当前用户是否有权限访问该 API
func CasbinAuth(enforcer *casbinpkg.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// enforcer 未初始化时拒绝受保护接口，避免 nil panic 与「未鉴权即放行」
		if enforcer == nil {
			response.Error(c, apperrors.Internal, "权限组件未就绪，请检查服务启动日志")
			c.Abort()
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			response.Error(c, apperrors.Unauthorized, "未登录")
			c.Abort()
			return
		}

		uid, ok := userID.(uint64)
		if !ok {
			response.Error(c, apperrors.Internal, "用户标识错误")
			c.Abort()
			return
		}

		allowed, err := enforcer.Enforce(
			fmt.Sprintf("%d", uid),
			c.Request.URL.Path,
			c.Request.Method,
		)
		if err != nil {
			response.Error(c, apperrors.Internal, "权限校验失败")
			c.Abort()
			return
		}

		if !allowed {
			response.Error(c, apperrors.Forbidden, apperrors.GetMsg(apperrors.Forbidden))
			c.Abort()
			return
		}

		c.Next()
	}
}
