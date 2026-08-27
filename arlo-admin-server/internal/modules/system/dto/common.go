package dto

// PageRequest 通用分页请求
type PageRequest struct {
	Page     int `json:"page" form:"page" binding:"required,min=1" example:"1"`    // 当前页码
	PageSize int `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100" example:"10"` // 每页条数
}

// PageResponse 通用分页响应
type PageResponse struct {
	List     interface{} `json:"list"`         // 数据列表
	Total    int64       `json:"total" example:"100"` // 总记录数
	Page     int         `json:"page" example:"1"`    // 当前页码
	PageSize int         `json:"pageSize" example:"10"` // 每页条数
}
