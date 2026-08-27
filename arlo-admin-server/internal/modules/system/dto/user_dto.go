package dto

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string  `json:"username" binding:"required,min=2,max=32" example:"zhangsan"`   // 用户名
	Password string  `json:"password" binding:"required,min=6,max=32" example:"123456"`     // 密码
	Nickname string  `json:"nickname" binding:"required,max=32" example:"张三"`              // 昵称
	Avatar   string  `json:"avatar" example:"1"`               // 头像（站内文件 ID，或外链 URL）
	Email    string  `json:"email" example:"zhangsan@example.com"`                          // 邮箱
	Phone    string  `json:"phone" example:"13800138000"`                                   // 手机号
	Gender   int8    `json:"gender" example:"1"`                                            // 性别（0=未知 1=男 2=女）
	DeptID   uint64  `json:"deptId" example:"1"`                                            // 部门ID
	Status   int8    `json:"status" example:"1"`                                            // 状态（0=禁用 1=启用）
	Remark   string  `json:"remark" example:"新员工"`                                          // 备注
	RoleIDs  []uint64 `json:"roleIds"`                                                      // 角色ID列表
	PostIDs  []uint64 `json:"postIds"`                                                      // 岗位ID列表
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	ID       uint64  `json:"id" binding:"required" example:"1"`                         // 用户ID
	Nickname string  `json:"nickname" binding:"required,max=32" example:"张三"`          // 昵称
	Avatar   string  `json:"avatar" example:"1"`           // 头像（站内文件 ID，或外链 URL）
	Email    string  `json:"email" example:"zhangsan@example.com"`                      // 邮箱
	Phone    string  `json:"phone" example:"13800138000"`                               // 手机号
	Gender   int8    `json:"gender" example:"1"`                                        // 性别（0=未知 1=男 2=女）
	DeptID   uint64  `json:"deptId" example:"1"`                                        // 部门ID
	Status   int8    `json:"status" example:"1"`                                        // 状态（0=禁用 1=启用）
	Remark   string  `json:"remark" example:"更新备注"`                                    // 备注
	RoleIDs  []uint64 `json:"roleIds"`                                                  // 角色ID列表
	PostIDs  []uint64 `json:"postIds"`                                                  // 岗位ID列表
}

// UpdateUserPasswordRequest 修改密码请求
type UpdateUserPasswordRequest struct {
	ID       uint64 `json:"id" binding:"required" example:"1"`                        // 用户ID
	Password string `json:"password" binding:"required,min=6,max=32" example:"newpwd"` // 新密码
}

// UserListRequest 用户列表查询
type UserListRequest struct {
	PageRequest
	Username string `json:"username" form:"username" example:"admin"` // 用户名（模糊查询）
	Nickname string `json:"nickname" form:"nickname" example:"管理员"`   // 昵称（模糊查询）
	Phone    string `json:"phone" form:"phone" example:"138"`         // 手机号（模糊查询）
	Status   *int8  `json:"status" form:"status" example:"1"`         // 状态（0=禁用 1=启用）
	DeptID   *uint64 `json:"deptId" form:"deptId" example:"1"`        // 部门ID
}

// UserResponse 用户列表项响应
type UserResponse struct {
	ID        uint64  `json:"id" example:"1"`                    // 用户ID
	Username  string  `json:"username" example:"admin"`          // 用户名
	Nickname  string  `json:"nickname" example:"系统管理员"`         // 昵称
	Avatar    string  `json:"avatar"`                            // 头像（站内文件 ID，或外链 URL）
	Email     string  `json:"email" example:"admin@example.com"` // 邮箱
	Phone     string  `json:"phone" example:"13800138000"`       // 手机号
	Gender    int8    `json:"gender" example:"1"`                // 性别（0=未知 1=男 2=女）
	DeptID    uint64  `json:"deptId" example:"1"`                // 部门ID
	DeptName  string  `json:"deptName" example:"技术部"`            // 部门名称
	Status    int8    `json:"status" example:"1"`                // 状态（0=禁用 1=启用）
	Remark    string  `json:"remark"`                            // 备注
	RoleIDs   []uint64 `json:"roleIds"`                          // 角色ID列表
	RoleNames []string `json:"roleNames"`                        // 角色名称列表
	PostIDs   []uint64 `json:"postIds"`                          // 岗位ID列表
	PostNames []string `json:"postNames"`                        // 岗位名称列表
	CreatedAt string  `json:"createdAt" example:"2026-07-10 12:00:00"` // 创建时间
}
