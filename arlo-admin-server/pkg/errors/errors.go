package errors

import "fmt"

// 通用错误码
const (
	OK           = 200
	BadRequest   = 400
	Unauthorized = 401
	Forbidden    = 403
	NotFound     = 404
	Internal     = 500
)

// 认证模块 [1000-1999]
const (
	ErrLoginFailed       = 1001
	ErrTokenExpired      = 1002
	ErrTokenInvalid      = 1003
	ErrUserDisabled      = 1004
	ErrCaptchaInvalid    = 1005
	ErrPasswordWrong     = 1006
	ErrUserNotFound      = 1007
	ErrOldPasswordWrong  = 1008
	ErrAccountLocked     = 1009
	ErrPasswordWeak      = 1010
	ErrMustChangePwd     = 1011
)

// 系统模块 [2000-2999]
const (
	ErrUserExists        = 2001
	ErrRoleExists        = 2002
	ErrDeptExists        = 2003
	ErrMenuExists        = 2004
	ErrDictTypeExists    = 2005
	ErrHasChildren       = 2006
	ErrRoleAssigned      = 2007
)

// 日志模块 [3000-3999]
const (
	ErrLogNotFound = 3001
)

// 消息模块 [4000-4999]
const (
	ErrMessageNotFound = 4001
)

// 配置模块 [5000-5999]
const (
	ErrConfigNotFound = 5001
	ErrUploadFailed   = 5002
	ErrFileTooLarge   = 5003
	ErrFileTypeInvalid = 5004
)

// 会员模块 [6000-6999]
const (
	ErrCodeSendFailed     = 6001
	ErrCodeInvalid        = 6002
	ErrMemberDisabled     = 6003
	ErrPhoneInvalid       = 6004
)

type AppError struct {
	Code int
	Msg  string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("code=%d, msg=%s", e.Code, e.Msg)
}

func New(code int, msg string) *AppError {
	return &AppError{Code: code, Msg: msg}
}

var errorMessages = map[int]string{
	OK:             "操作成功",
	BadRequest:     "请求参数错误",
	Unauthorized:   "未授权，请先登录",
	Forbidden:      "无权限访问",
	NotFound:       "资源不存在",
	Internal:       "服务器内部错误",
	ErrLoginFailed: "登录失败",
	ErrTokenExpired:  "Token已过期",
	ErrTokenInvalid:  "Token无效",
	ErrUserDisabled:  "用户已被禁用",
	ErrCaptchaInvalid:"验证码错误",
	ErrPasswordWrong: "密码错误",
	ErrUserNotFound:  "用户不存在",
	ErrOldPasswordWrong: "原密码错误",
	ErrAccountLocked:    "账号已锁定，请稍后再试",
	ErrPasswordWeak:     "密码不符合安全策略",
	ErrMustChangePwd:    "请先修改密码",
	ErrUserExists:    "用户已存在",
	ErrRoleExists:    "角色已存在",
	ErrDeptExists:    "部门已存在",
	ErrMenuExists:    "菜单已存在",
	ErrHasChildren:   "存在子节点，无法删除",
	ErrRoleAssigned:  "角色已被分配，无法删除",
	ErrCodeSendFailed: "验证码发送失败",
	ErrCodeInvalid:    "验证码错误",
	ErrMemberDisabled: "账号已被禁用",
	ErrPhoneInvalid:   "手机号格式错误",
}

func GetMsg(code int) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
