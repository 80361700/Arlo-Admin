package dto

// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
	ParentID   uint64 `json:"parentId" example:"0"`                          // 父菜单ID（0=顶级菜单）
	Name       string `json:"name" binding:"required,max=32" example:"用户管理"` // 菜单名称
	Type       int8   `json:"type" binding:"required,min=1,max=3" example:"2"` // 类型（1=目录 2=菜单 3=按钮）
	Path       string `json:"path" example:"/system/user"`                   // 路由路径
	Component  string `json:"component" example:"system/user/index"`         // 前端组件路径
	Icon       string `json:"icon" example:"user"`                           // 菜单图标
	Sort       int    `json:"sort" example:"1"`                              // 排序
	Permission string `json:"permission" example:"sys:user:list"`            // 权限标识
	Status     int8   `json:"status" example:"1"`                            // 状态（0=禁用 1=启用）
	Visible    int8   `json:"visible" example:"1"`                           // 是否可见（0=隐藏 1=显示）
	KeepAlive  int8   `json:"keepAlive" example:"1"`                         // 是否缓存（0=否 1=是）
}

// UpdateMenuRequest 更新菜单请求
type UpdateMenuRequest struct {
	ID         uint64 `json:"id" binding:"required" example:"1"`              // 菜单ID
	ParentID   uint64 `json:"parentId" example:"0"`                           // 父菜单ID
	Name       string `json:"name" binding:"required,max=32" example:"用户管理"`  // 菜单名称
	Type       int8   `json:"type" binding:"required,min=1,max=3" example:"2"` // 类型（1=目录 2=菜单 3=按钮）
	Path       string `json:"path" example:"/system/user"`                    // 路由路径
	Component  string `json:"component" example:"system/user/index"`          // 前端组件路径
	Icon       string `json:"icon" example:"user"`                            // 菜单图标
	Sort       int    `json:"sort" example:"1"`                               // 排序
	Permission string `json:"permission" example:"sys:user:list"`             // 权限标识
	Status     int8   `json:"status" example:"1"`                             // 状态（0=禁用 1=启用）
	Visible    int8   `json:"visible" example:"1"`                            // 是否可见（0=隐藏 1=显示）
	KeepAlive  int8   `json:"keepAlive" example:"1"`                          // 是否缓存（0=否 1=是）
}

// MenuTreeResponse 菜单树节点响应
type MenuTreeResponse struct {
	ID         uint64              `json:"id" example:"1"`                    // 菜单ID
	ParentID   uint64              `json:"parentId" example:"0"`              // 父菜单ID
	Name       string              `json:"name" example:"用户管理"`              // 菜单名称
	Type       int8                `json:"type" example:"2"`                  // 类型（1=目录 2=菜单 3=按钮）
	Path       string              `json:"path" example:"/system/user"`       // 路由路径
	Component  string              `json:"component" example:"system/user/index"` // 前端组件路径
	Icon       string              `json:"icon" example:"user"`               // 菜单图标
	Sort       int                 `json:"sort" example:"1"`                  // 排序
	Permission string              `json:"permission" example:"sys:user:list"` // 权限标识
	Status     int8                `json:"status" example:"1"`                // 状态（0=禁用 1=启用）
	Visible    int8                `json:"visible" example:"1"`               // 是否可见（0=隐藏 1=显示）
	KeepAlive  int8                `json:"keepAlive" example:"1"`             // 是否缓存（0=否 1=是）
	Children   []*MenuTreeResponse `json:"children"`                          // 子菜单列表
}
