package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/log/model"

	"github.com/gin-gonic/gin"
)

// responseWriter 包装 gin.ResponseWriter 以捕获状态码
type responseWriter struct {
	gin.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

// OperationLog 操作日志中间件
// 自动记录每个 API 请求的用户、方法、URL、耗时、状态码等信息
func OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查和 Swagger 请求
		path := c.Request.URL.Path
		if path == "/health" || strings.HasPrefix(path, "/swagger") {
			c.Next()
			return
		}

		startTime := time.Now()

		// 捕获请求体（需要先读取再放回去）
		var params string
		if c.Request.Method == "GET" {
			// GET 请求从 URL query 参数获取
			params = c.Request.URL.RawQuery
		} else if c.Request.Body != nil && c.Request.ContentLength > 0 {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				params = maskSensitiveParams(string(bodyBytes))
				if len(params) > 2000 {
					params = params[:2000] + "..."
				}
			}
		}

		// 包装 ResponseWriter 以捕获状态码
		rw := &responseWriter{ResponseWriter: c.Writer, statusCode: 200}
		c.Writer = rw

		c.Next()

		// 请求结束后计算耗时
		costTime := int(time.Since(startTime).Milliseconds())

		// 获取用户信息（可能为空，未登录接口没有）
		userID, _ := c.Get("userID")
		username, _ := c.Get("username")

		uid := uint64(0)
		uname := ""
		if v, ok := userID.(uint64); ok {
			uid = v
		}
		if v, ok := username.(string); ok {
			uname = v
		}

		// 如果上下文中没有用户名，尝试从 POST JSON 请求体中解析（如登录接口）
		if uname == "" && params != "" {
			if u := extractUsernameFromBody(params); u != "" {
				uname = u
			}
		}

		// 状态：HTTP 4xx/5xx 视为失败
		status := int8(1)
		errMsg := ""
		if rw.statusCode >= 400 {
			status = 0

			// 尝试获取错误信息（业务层通过 c.Set 设置的）
			if errStr, exists := c.Get("error_msg"); exists {
				if s, ok := errStr.(string); ok {
					errMsg = s
				}
			}
		}

		// 模块和操作映射
		module := extractModule(path)
		action := extractAction(c.Request.Method)

		opLog := &model.OperationLog{
			UserID:    uid,
			Username:  uname,
			Module:    module,
			Action:    action,
			Method:    c.Request.Method,
			URL:       path,
			IP:        c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Params:    params,
			CostTime:  costTime,
			Status:    status,
			ErrorMsg:  errMsg,
			CreatedAt: time.Now(),
		}

		// 异步入库，不阻塞请求响应（无 DB 时跳过，避免 debug 降级启动后 panic）
		if database.DB == nil {
			return
		}
		go func() {
			_ = database.DB.Create(opLog).Error
		}()
	}
}

// extractUsernameFromBody 从 JSON 请求体中解析 username 字段
// 用于未登录接口（如登录）记录操作人
func extractUsernameFromBody(body string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return ""
	}
	if v, ok := data["username"]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

// 操作日志脱敏字段（大小写不敏感）
var sensitiveParamKeys = map[string]struct{}{
	"password":        {},
	"oldpassword":     {},
	"newpassword":     {},
	"confirmpassword": {},
	"pwd":             {},
	"oldpwd":          {},
	"newpwd":          {},
	"captcha":         {},
	"captchacode":     {},
	"code":            {}, // 短信验证码等
	"token":           {},
	"accesstoken":     {},
	"refreshtoken":    {},
	"secret":          {},
	"accesskeysecret": {},
}

// maskSensitiveParams 对 JSON 请求体中的敏感字段替换为 ***
func maskSensitiveParams(body string) string {
	body = strings.TrimSpace(body)
	if body == "" || body[0] != '{' {
		return body
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return body
	}
	changed := false
	for k := range data {
		if _, ok := sensitiveParamKeys[strings.ToLower(k)]; ok {
			data[k] = "***"
			changed = true
		}
	}
	if !changed {
		return body
	}
	b, err := json.Marshal(data)
	if err != nil {
		return body
	}
	return string(b)
}

func extractModule(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	// /api/v1/{module}/{sub}/... → module/sub
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "v1" {
		if len(parts) >= 4 {
			return parts[2] + "/" + parts[3]
		}
		return parts[2]
	}
	return path
}

// extractAction 从 HTTP 方法映射操作类型
func extractAction(method string) string {
	switch method {
	case "GET":
		return "查询"
	case "POST":
		return "新增"
	case "PUT":
		return "修改"
	case "DELETE":
		return "删除"
	case "PATCH":
		return "修改"
	default:
		return method
	}
}
