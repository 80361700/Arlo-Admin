package dto

// LoginRequest 登录请求
type LoginRequest struct {
	Username    string `json:"username" binding:"required" example:"admin"`    // 用户名
	Password    string `json:"password" binding:"required" example:"123456"`   // 密码
	CaptchaID   string `json:"captchaId" example:"abc123"`                     // 验证码ID
	CaptchaCode string `json:"captchaCode" example:"1234"`                     // 验证码
	IP          string `json:"-"`                                              // 登录IP（由服务端填充）
	Browser     string `json:"-"`                                              // 浏览器（由服务端填充）
	OS          string `json:"-"`                                              // 操作系统（由服务端填充）
}

// CaptchaResponse 验证码响应
type CaptchaResponse struct {
	CaptchaID string `json:"captchaId" example:"abc123"`                                                  // 验证码ID
	Captcha   string `json:"captcha" example:"data:image/png;base64,iVBORw0KGgo..."`                     // 验证码图片（base64）
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIs..."`  // 访问令牌
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIs..."` // 刷新令牌
	ExpiresIn    int64  `json:"expiresIn" example:"7200"`                       // 过期时间（秒）
}

// RefreshRequest 刷新令牌请求
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required" example:"eyJhbGciOiJIUzI1NiIs..."` // 刷新令牌
}

// LogoutRequest 登出请求（可选带 refresh，一并作废）
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIs..."`
}

// RefreshResponse 刷新令牌响应
type RefreshResponse struct {
	AccessToken string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIs..."` // 访问令牌
	ExpiresIn   int64  `json:"expiresIn" example:"7200"`                      // 过期时间（秒）
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	ID          uint64   `json:"id" example:"1"`                   // 用户ID
	Username    string   `json:"username" example:"admin"`         // 用户名
	Nickname    string   `json:"nickname" example:"系统管理员"`         // 昵称
	Avatar      string   `json:"avatar" example:"1"`               // 头像（站内文件 ID，或外链 URL）
	Email       string   `json:"email" example:"admin@example.com"` // 邮箱
	Phone       string   `json:"phone" example:"13800138000"`      // 手机号
	Gender      int8     `json:"gender" example:"1"`               // 性别（0=未知 1=男 2=女）
	DeptID      uint64   `json:"deptId" example:"1"`               // 部门ID
	DeptName    string   `json:"deptName" example:"技术部"`           // 部门名称
	Status      int8     `json:"status" example:"1"`               // 状态（0=禁用 1=启用）
	Remark      string   `json:"remark"`                           // 备注
	RoleNames   []string `json:"roleNames"`                        // 角色名称列表
	PostNames   []string `json:"postNames"`                        // 岗位名称列表
	Permissions []string `json:"permissions"`                      // 权限标识列表
	DataScope     int8     `json:"dataScope" example:"1"`            // 数据范围: 1=全部 2=自定义 3=本部门及以下 4=本部门 5=仅本人
	DeptIDs       []uint64 `json:"deptIds"`                          // 自定义数据时可见的部门ID
	MustChangePwd bool     `json:"mustChangePwd"`                    // 是否强制改密
	PwdExpired    bool     `json:"pwdExpired"`                       // 密码是否已过期
}

// UpdateProfileRequest 更新个人资料请求（部门不可自改，仅展示）
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"required,max=32" example:"张三"`
	Gender   int8   `json:"gender" example:"1"`
	Phone    string `json:"phone" example:"13800138000"`
	Email    string `json:"email" example:"zhangsan@example.com"`
	Remark   string `json:"remark" example:""`
	Avatar   string `json:"avatar" example:""`
}

// ChangePasswordRequest 修改本人密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required" example:"oldpwd"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=32" example:"newpwd"`
}

// MenuTreeNode 当前用户可见菜单树节点（侧边栏/动态路由用）
type MenuTreeNode struct {
	ID         uint64          `json:"id"`
	ParentID   uint64          `json:"parentId"`
	Name       string          `json:"name"`
	Type       int8            `json:"type"` // 1=目录 2=菜单 3=按钮
	Path       string          `json:"path"`
	Component  string          `json:"component"`
	Icon       string          `json:"icon"`
	Sort       int             `json:"sort"`
	Permission string          `json:"permission"`
	Visible    int8            `json:"visible"`
	KeepAlive  int8            `json:"keepAlive"`
	Children   []*MenuTreeNode `json:"children,omitempty"`
}
