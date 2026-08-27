package dto

// CreateDeptRequest 创建部门请求
type CreateDeptRequest struct {
	ParentID uint64 `json:"parentId" example:"0"`                           // 父部门ID（0=顶级部门）
	Name     string `json:"name" binding:"required,max=64" example:"技术部"`   // 部门名称
	Sort     int    `json:"sort" example:"1"`                               // 排序
	Leader   string `json:"leader" example:"张三"`                            // 负责人
	Phone    string `json:"phone" example:"13800138000"`                    // 联系电话
	Email    string `json:"email" example:"tech@example.com"`               // 邮箱
	Status   int8   `json:"status" example:"1"`                             // 状态（0=禁用 1=启用）
}

// UpdateDeptRequest 更新部门请求
type UpdateDeptRequest struct {
	ID       uint64 `json:"id" binding:"required" example:"1"`              // 部门ID
	ParentID uint64 `json:"parentId" example:"0"`                           // 父部门ID
	Name     string `json:"name" binding:"required,max=64" example:"技术部"`   // 部门名称
	Sort     int    `json:"sort" example:"1"`                               // 排序
	Leader   string `json:"leader" example:"张三"`                            // 负责人
	Phone    string `json:"phone" example:"13800138000"`                    // 联系电话
	Email    string `json:"email" example:"tech@example.com"`               // 邮箱
	Status   int8   `json:"status" example:"1"`                             // 状态（0=禁用 1=启用）
}

// DeptTreeResponse 部门树节点响应
type DeptTreeResponse struct {
	ID       uint64             `json:"id" example:"1"`                   // 部门ID
	ParentID uint64             `json:"parentId" example:"0"`              // 父部门ID
	Name     string             `json:"name" example:"技术部"`               // 部门名称
	Sort     int                `json:"sort" example:"1"`                  // 排序
	Leader   string             `json:"leader" example:"张三"`              // 负责人
	Phone    string             `json:"phone" example:"13800138000"`       // 联系电话
	Email    string             `json:"email" example:"tech@example.com"` // 邮箱
	Status   int8               `json:"status" example:"1"`               // 状态（0=禁用 1=启用）
	Children []*DeptTreeResponse `json:"children"`                        // 子部门列表
}
