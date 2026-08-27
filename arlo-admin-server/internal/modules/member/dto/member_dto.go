package dto

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
}

// LoginRequest 会员登录请求
type LoginRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
	Code  string `json:"code" binding:"required,len=6"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// RefreshRequest 刷新 Token 请求
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// MemberInfoResponse 会员信息响应
type MemberInfoResponse struct {
	ID        uint64 `json:"id"`
	Phone     string `json:"phone"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Gender    int8   `json:"gender"`
	Source    string `json:"source"`
	Status    int8   `json:"status"`
	CreatedAt string `json:"createdAt"`
}

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"required,max=32"`
	Avatar   string `json:"avatar" binding:"max=255"`
	Gender   int8   `json:"gender" binding:"oneof=0 1 2"`
}

// PageRequest 分页请求（管理员查看会员列表）
type PageRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"pageSize" binding:"min=1,max=100"`
	Phone    string `form:"phone"`
	Nickname string `form:"nickname"`
	Source   string `form:"source"`
	Status   *int8  `form:"status"`
}

// MemberItem 会员列表项（管理员视角）
type MemberItem struct {
	ID        uint64 `json:"id"`
	Phone     string `json:"phone"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Gender    int8   `json:"gender"`
	Source    string `json:"source"`
	Status    int8   `json:"status"`
	LastLogin string `json:"lastLogin"`
	CreatedAt string `json:"createdAt"`
}

// MemberDetailResponse 会员详情（管理员）
type MemberDetailResponse struct {
	ID        uint64 `json:"id"`
	Phone     string `json:"phone"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Gender    int8   `json:"gender"`
	Openid    string `json:"openid"`
	Unionid   string `json:"unionid"`
	MpOpenid  string `json:"mpOpenid"`
	Source    string `json:"source"`
	Status    int8   `json:"status"`
	LastLogin string `json:"lastLogin"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// AdminCreateMemberRequest 管理员手动录入会员
type AdminCreateMemberRequest struct {
	Phone    string `json:"phone" binding:"required,len=11"`
	Password string `json:"password" binding:"omitempty,min=6,max=32"`
	Nickname string `json:"nickname" binding:"required,max=32"`
	Avatar   string `json:"avatar" binding:"max=255"`
	Gender   int8   `json:"gender" binding:"oneof=0 1 2"`
	Source   string `json:"source" binding:"required,oneof=h5 mini oa"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

// AdminUpdateMemberRequest 管理员更新会员资料
type AdminUpdateMemberRequest struct {
	ID       uint64 `json:"id" binding:"required"`
	Nickname string `json:"nickname" binding:"required,max=32"`
	Avatar   string `json:"avatar" binding:"max=255"`
	Gender   int8   `json:"gender" binding:"oneof=0 1 2"`
	Source   string `json:"source" binding:"required,oneof=h5 mini oa"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

// UpdateMemberPasswordRequest 管理员重置会员密码
type UpdateMemberPasswordRequest struct {
	ID       uint64 `json:"id" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

// UpdateStatusRequest 管理员更新状态
type UpdateStatusRequest struct {
	Status int8 `json:"status" binding:"oneof=0 1"`
}
