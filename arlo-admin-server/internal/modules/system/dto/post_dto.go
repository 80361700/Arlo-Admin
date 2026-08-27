package dto

// CreatePostRequest 创建岗位请求
type CreatePostRequest struct {
	Code   string `json:"code" binding:"required,max=32" example:"ceo"`  // 岗位编码
	Name   string `json:"name" binding:"required,max=64" example:"总经理"`  // 岗位名称
	Sort   int    `json:"sort" example:"1"`                               // 排序
	Status int8   `json:"status" example:"1"`                             // 状态（0=禁用 1=启用）
	Remark string `json:"remark" example:"公司负责人"`                         // 备注
}

// UpdatePostRequest 更新岗位请求
type UpdatePostRequest struct {
	ID     uint64 `json:"id" binding:"required" example:"1"`          // 岗位ID
	Code   string `json:"code" binding:"required,max=32" example:"ceo"` // 岗位编码
	Name   string `json:"name" binding:"required,max=64" example:"总经理"` // 岗位名称
	Sort   int    `json:"sort" example:"1"`                            // 排序
	Status int8   `json:"status" example:"1"`                          // 状态（0=禁用 1=启用）
	Remark string `json:"remark" example:"公司负责人"`                      // 备注
}

// PostListRequest 岗位列表查询
type PostListRequest struct {
	PageRequest
	Code   string `json:"code" form:"code" example:"ceo"` // 岗位编码（模糊查询）
	Name   string `json:"name" form:"name" example:"经理"`  // 岗位名称（模糊查询）
	Status *int8  `json:"status" form:"status" example:"1"` // 状态（0=禁用 1=启用）
}

// PostResponse 岗位响应
type PostResponse struct {
	ID        uint64 `json:"id" example:"1"`                               // 岗位ID
	Code      string `json:"code" example:"ceo"`                           // 岗位编码
	Name      string `json:"name" example:"总经理"`                           // 岗位名称
	Sort      int    `json:"sort" example:"1"`                             // 排序
	Status    int8   `json:"status" example:"1"`                           // 状态（0=禁用 1=启用）
	Remark    string `json:"remark" example:"公司负责人"`                      // 备注
	CreatedAt string `json:"createdAt" example:"2026-07-10 12:00:00"`      // 创建时间
}
