package dto

// OnlineQuery 在线用户查询
type OnlineQuery struct {
	Username string `form:"username"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

// OnlineSessionItem 在线会话项
type OnlineSessionItem struct {
	UserID     uint64 `json:"userId"`
	Username   string `json:"username"`
	RefreshJTI string `json:"sessionId"` // 会话 ID（refresh jti）
	IP         string `json:"ip"`
	Browser    string `json:"browser"`
	OS         string `json:"os"`
	LoginAt    string `json:"loginAt"`
}

// OnlineListResponse 在线用户列表
type OnlineListResponse struct {
	List     []OnlineSessionItem `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}

// KickRequest 强制下线请求
type KickRequest struct {
	UserID    uint64 `json:"userId" binding:"required"`
	SessionID string `json:"sessionId"` // 空则踢该用户全部会话
}
