package dto

import (
	"arlo-admin/internal/modules/system/dto"
)

// SendMessageRequest 发送站内信请求
// ReceiverIDs 为空或省略 = 广播全部用户；非空 = 指定接收者批量发送
type SendMessageRequest struct {
	Title       string   `json:"title" binding:"required" example:"系统通知"`   // 消息标题
	Content     string   `json:"content" binding:"required" example:"您的申请已通过"` // 消息内容
	Type        int8     `json:"type" example:"1"`                        // 类型（1=系统消息 2=通知 3=私信）
	ReceiverIDs []uint64 `json:"receiverIds"`                               // 接收者ID列表；空=广播
}

// MessageListQuery 站内信查询参数
type MessageListQuery struct {
	IsRead    *int8 `form:"isRead" example:"0"`    // 是否已读（0=未读 1=已读）
	Direction *int8 `form:"direction" example:"0"` // 消息范围（0=全部 1=我收到的 2=我发送的）
	dto.PageRequest
}

// MessageResponse 站内信响应
type MessageResponse struct {
	ID            uint64 `json:"id" example:"1"`                           // 消息ID
	Title         string `json:"title" example:"系统通知"`                    // 消息标题
	Content       string `json:"content" example:"您的申请已通过"`              // 消息内容
	Type          int8   `json:"type" example:"1"`                         // 类型（1=系统消息 2=通知 3=私信）
	SenderID      uint64 `json:"senderId" example:"0"`                     // 发送者ID
	Sender        string `json:"sender" example:"系统"`                       // 发送者
	ReceiverID    uint64 `json:"receiverId" example:"2"`                     // 接收者ID
	ReceiverCount int    `json:"receiverCount" example:"1"`                // 接收者数量（发送记录聚合用）
	ReceiverName  string `json:"receiverName" example:"张三"`                // 接收者名称
	IsRead        int8   `json:"isRead" example:"0"`                         // 是否已读（0=未读 1=已读）
	ReadAt        string `json:"readAt" example:""`                        // 阅读时间
	CreatedAt     string `json:"createdAt" example:"2026-07-10 12:00:00"`   // 发送时间
}

// MessageListResponse 站内信分页响应
type MessageListResponse struct {
	List     []MessageResponse `json:"list"`     // 消息列表
	Total    int64             `json:"total" example:"20"`    // 总记录数
	Page     int               `json:"page" example:"1"`      // 当前页码
	PageSize int               `json:"pageSize" example:"10"` // 每页条数
}

// UnreadCountResponse 未读消息数响应
type UnreadCountResponse struct {
	Count int64 `json:"count" example:"5"` // 未读消息数
}
