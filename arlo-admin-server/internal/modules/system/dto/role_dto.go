package dto

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name   string `json:"name" binding:"required,max=32" example:"系统管理员"`          // 角色名称
	Code   string `json:"code" binding:"required,max=32" example:"admin"`              // 角色编码
	Sort   int    `json:"sort" example:"1"`                                            // 排序
	Status int8   `json:"status" example:"1"`                                          // 状态（0=禁用 1=启用）
	Remark    string `json:"remark" example:"超级管理员角色"`                                   // 备注
	DataScope int8   `json:"dataScope" example:"1"`                                          // 数据范围: 1=全部 2=自定义 3=本部门及以下 4=本部门 5=仅本人
	DeptIDs   []uint64 `json:"deptIds"`                                                     // 自定义数据时关联的部门ID列表
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	ID        uint64   `json:"id" binding:"required" example:"1"`                       // 角色ID
	Name      string   `json:"name" binding:"required,max=32" example:"系统管理员"`         // 角色名称
	Code      string   `json:"code" binding:"required,max=32" example:"admin"`          // 角色编码
	Sort      int      `json:"sort" example:"1"`                                        // 排序
	Status    int8     `json:"status" example:"1"`                                      // 状态（0=禁用 1=启用）
	Remark    string   `json:"remark" example:"超级管理员角色"`                               // 备注
	DataScope int8     `json:"dataScope" example:"1"`                                   // 数据范围: 1=全部 2=自定义 3=本部门及以下 4=本部门 5=仅本人
	DeptIDs   []uint64 `json:"deptIds"`                                                // 自定义数据时关联的部门ID列表
}

// RoleListRequest 角色列表查询
type RoleListRequest struct {
	PageRequest
	Name   string `json:"name" form:"name" example:"管理员"` // 角色名称（模糊查询）
	Code   string `json:"code" form:"code" example:"admin"` // 角色编码（模糊查询）
	Status *int8  `json:"status" form:"status" example:"1"` // 状态（0=禁用 1=启用）
}

// AssignRoleMenusRequest 分配角色菜单
type AssignRoleMenusRequest struct {
	RoleID  uint64   `json:"roleId" binding:"required" example:"1"` // 角色ID
	MenuIDs []uint64 `json:"menuIds"`                               // 菜单ID列表
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID        uint64   `json:"id" example:"1"`                               // 角色ID
	Name      string   `json:"name" example:"系统管理员"`                        // 角色名称
	Code      string   `json:"code" example:"admin"`                         // 角色编码
	Sort      int      `json:"sort" example:"1"`                             // 排序
	Status    int8     `json:"status" example:"1"`                           // 状态（0=禁用 1=启用）
	Remark    string   `json:"remark" example:"超级管理员角色"`                    // 备注
	DataScope int8     `json:"dataScope" example:"1"`                        // 数据范围: 1=全部 2=自定义 3=本部门及以下 4=本部门 5=仅本人
	DeptIDs   []uint64 `json:"deptIds"`                                     // 自定义数据时关联的部门ID列表
	CreatedAt string   `json:"createdAt" example:"2026-07-10 12:00:00"`      // 创建时间
}
