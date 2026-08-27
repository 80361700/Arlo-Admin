package middleware

import (
	"bytes"
	"io"
	"runtime/debug"
	"time"

	"arlo-admin/pkg/logger"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fields := []zap.Field{
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.ByteString("stack", debug.Stack()),
				}
				if rid := c.GetString("RequestID"); rid != "" {
					fields = append(fields, zap.String("request_id", rid))
				}
				logger.Logger.Error("panic recovered", fields...)

				c.AbortWithStatusJSON(500, gin.H{
					"code": 500,
					"msg":  "服务器内部错误",
					"data": nil,
				})
			}
		}()
		c.Next()
	}
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		logger.Logger.Info("request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
		)
	}
}
