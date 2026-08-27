package dto

import (
	"arlo-admin/internal/modules/system/dto"
)

// CreateNoticeRequest 创建通知公告请求
type CreateNoticeRequest struct {
	Title   string `json:"title" binding:"required" example:"系统升级通知"`     // 标题
	Content string `json:"content" binding:"required" example:"系统将于今晚升级..."` // 内容
	Type    int8   `json:"type" example:"1"`                               // 类型（1=通知 2=公告）
	Level   int8   `json:"level" example:"1"`                              // 级别（1=普通 2=重要 3=紧急）
}

// UpdateNoticeRequest 更新通知公告请求
type UpdateNoticeRequest struct {
	Title   string `json:"title" binding:"required" example:"系统升级通知"`     // 标题
	Content string `json:"content" binding:"required" example:"系统将于今晚升级..."` // 内容
	Type    int8   `json:"type" example:"1"`                               // 类型（1=通知 2=公告）
	Level   int8   `json:"level" example:"1"`                              // 级别（1=普通 2=重要 3=紧急）
}

// NoticeListQuery 通知公告查询参数
type NoticeListQuery struct {
	Title  string `form:"title" example:"系统升级"` // 标题（模糊查询）
	Status *int8  `form:"status" example:"1"`   // 状态（0=草稿 1=已发布 2=已撤回）
	dto.PageRequest
}

// NoticeResponse 通知公告响应
type NoticeResponse struct {
	ID          uint64 `json:"id" example:"1"`                        // 公告ID
	Title       string `json:"title" example:"系统升级通知"`               // 标题
	Content     string `json:"content" example:"系统将于今晚升级..."`        // 内容
	Type        int8   `json:"type" example:"1"`                      // 类型（1=通知 2=公告）
	Level       int8   `json:"level" example:"1"`                     // 级别（1=普通 2=重要 3=紧急）
	Status      int8   `json:"status" example:"1"`                    // 状态（0=草稿 1=已发布 2=已撤回）
	PublisherID uint64 `json:"publisherId" example:"1"`               // 发布人ID
	Publisher   string `json:"publisher" example:"admin"`             // 发布人
	CreatedAt   string `json:"createdAt" example:"2026-07-10 12:00:00"` // 创建时间
	UpdatedAt   string `json:"updatedAt" example:"2026-07-10 12:00:00"` // 更新时间
}

// NoticeListResponse 通知公告分页响应
type NoticeListResponse struct {
	List     []NoticeResponse `json:"list"`     // 公告列表
	Total    int64            `json:"total" example:"50"`    // 总记录数
	Page     int              `json:"page" example:"1"`      // 当前页码
	PageSize int              `json:"pageSize" example:"10"` // 每页条数
}
