package dto

import (
	"arlo-admin/internal/modules/system/dto"
)

// LoginLogQuery 登录日志查询参数
type LoginLogQuery struct {
	Username  string `form:"username" example:"admin"`                // 用户名
	Status    *int8  `form:"status" example:"1"`                      // 状态（0=失败 1=成功）
	StartTime string `form:"startTime" example:"2026-07-01 00:00:00"` // 开始时间
	EndTime   string `form:"endTime" example:"2026-07-20 23:59:59"`   // 结束时间
	dto.PageRequest
}

// LoginLogResponse 登录日志响应
type LoginLogResponse struct {
	ID        uint64 `json:"id" example:"1"`                          // 日志ID
	Username  string `json:"username" example:"admin"`                // 用户名
	IP        string `json:"ip" example:"127.0.0.1"`                  // 登录IP
	Location  string `json:"location" example:"本地"`                  // 登录地点
	Browser   string `json:"browser" example:"Chrome 120"`            // 浏览器
	OS        string `json:"os" example:"macOS"`                      // 操作系统
	Status    int8   `json:"status" example:"1"`                      // 状态（0=失败 1=成功）
	Msg       string `json:"msg" example:"登录成功"`                   // 提示消息
	CreatedAt string `json:"createdAt" example:"2026-07-10 12:00:00"` // 登录时间
}

// OperationLogQuery 操作日志查询参数
type OperationLogQuery struct {
	Username  string `form:"username" example:"admin"`                    // 操作人
	Module    string `form:"module" example:"系统管理"`                     // 操作模块
	URL       string `form:"url" example:"/api/v1/system/user"`           // 请求地址（模糊）
	StartTime string `form:"startTime" example:"2026-07-01 00:00:00"`     // 开始时间
	EndTime   string `form:"endTime" example:"2026-07-20 23:59:59"`       // 结束时间
	Status    *int8  `form:"status" example:"1"`                          // 状态（0=失败 1=成功）
	dto.PageRequest
}

// OperationLogResponse 操作日志响应
type OperationLogResponse struct {
	ID        uint64 `json:"id" example:"1"`                            // 日志ID
	UserID    uint64 `json:"userId" example:"1"`                        // 用户ID
	Username  string `json:"username" example:"admin"`                  // 用户名
	Module    string `json:"module" example:"用户管理"`                     // 操作模块
	Action    string `json:"action" example:"查询"`                       // 操作类型
	Method    string `json:"method" example:"GET"`                      // 请求方法
	URL       string `json:"url" example:"/api/v1/system/user/list"`    // 请求URL
	IP        string `json:"ip" example:"127.0.0.1"`                    // 操作IP
	UserAgent string `json:"userAgent" example:"Mozilla/5.0..."`        // 用户代理
	Params    string `json:"params" example:"{\"page\":1,\"pageSize\":10}"` // 请求参数
	Result    string `json:"result"`                                    // 响应摘要
	CostTime  int    `json:"costTime" example:"12"`                     // 耗时（毫秒）
	Status    int8   `json:"status" example:"1"`                        // 状态（0=失败 1=成功）
	ErrorMsg  string `json:"errorMsg" example:""`                       // 错误信息
	CreatedAt string `json:"createdAt" example:"2026-07-10 12:00:00"`   // 操作时间
}

// LoginLogListResponse 登录日志分页响应
type LoginLogListResponse struct {
	List     []LoginLogResponse `json:"list"`     // 登录日志列表
	Total    int64              `json:"total" example:"100"`    // 总记录数
	Page     int                `json:"page" example:"1"`       // 当前页码
	PageSize int                `json:"pageSize" example:"10"`  // 每页条数
}

// OperationLogListResponse 操作日志分页响应
type OperationLogListResponse struct {
	List     []OperationLogResponse `json:"list"`     // 操作日志列表
	Total    int64                  `json:"total" example:"500"`    // 总记录数
	Page     int                    `json:"page" example:"1"`       // 当前页码
	PageSize int                    `json:"pageSize" example:"10"`  // 每页条数
}
