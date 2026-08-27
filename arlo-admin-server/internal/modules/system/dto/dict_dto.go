package dto

// CreateDictTypeRequest 创建字典类型请求
type CreateDictTypeRequest struct {
	Name   string `json:"name" binding:"required,max=64" example:"用户性别"`   // 字典名称
	Code   string `json:"code" binding:"required,max=64" example:"sys_gender"` // 字典编码
	Status int8   `json:"status" example:"1"`                              // 状态（0=禁用 1=启用）
	Remark string `json:"remark" example:"性别字典"`                           // 备注
}

// UpdateDictTypeRequest 更新字典类型请求
type UpdateDictTypeRequest struct {
	ID     uint64 `json:"id" binding:"required" example:"1"`              // 字典类型ID
	Name   string `json:"name" binding:"required,max=64" example:"用户性别"`  // 字典名称
	Code   string `json:"code" binding:"required,max=64" example:"sys_gender"` // 字典编码
	Status int8   `json:"status" example:"1"`                             // 状态（0=禁用 1=启用）
	Remark string `json:"remark" example:"性别字典"`                          // 备注
}

// DictTypeListRequest 字典类型列表查询
type DictTypeListRequest struct {
	PageRequest
	Name   string `json:"name" form:"name" example:"性别"`    // 字典名称（模糊查询）
	Code   string `json:"code" form:"code" example:"gender"` // 字典编码（模糊查询）
	Status *int8  `json:"status" form:"status" example:"1"`  // 状态（0=禁用 1=启用）
}

// DictTypeResponse 字典类型响应
type DictTypeResponse struct {
	ID        uint64 `json:"id" example:"1"`                               // 字典类型ID
	Name      string `json:"name" example:"用户性别"`                          // 字典名称
	Code      string `json:"code" example:"sys_gender"`                    // 字典编码
	Status    int8   `json:"status" example:"1"`                           // 状态（0=禁用 1=启用）
	Remark    string `json:"remark" example:"性别字典"`                        // 备注
	CreatedAt string `json:"createdAt" example:"2026-07-10 12:00:00"`      // 创建时间
}

// CreateDictDataRequest 创建字典数据请求
type CreateDictDataRequest struct {
	DictTypeID uint64 `json:"dictTypeId" binding:"required" example:"1"`    // 字典类型ID
	Label      string `json:"label" binding:"required,max=64" example:"男"`  // 字典标签
	Value      string `json:"value" binding:"required,max=64" example:"1"`  // 字典值
	Sort       int    `json:"sort" example:"1"`                             // 排序
	IsDefault  int8   `json:"isDefault" example:"0"`                        // 是否默认（0=否 1=是）
	Status     int8   `json:"status" example:"1"`                           // 状态（0=禁用 1=启用）
	Remark     string `json:"remark" example:""`                             // 备注
}

// UpdateDictDataRequest 更新字典数据请求
type UpdateDictDataRequest struct {
	ID         uint64 `json:"id" binding:"required" example:"1"`            // 字典数据ID
	DictTypeID uint64 `json:"dictTypeId" binding:"required" example:"1"`    // 字典类型ID
	Label      string `json:"label" binding:"required,max=64" example:"男"`  // 字典标签
	Value      string `json:"value" binding:"required,max=64" example:"1"`  // 字典值
	Sort       int    `json:"sort" example:"1"`                             // 排序
	IsDefault  int8   `json:"isDefault" example:"0"`                        // 是否默认（0=否 1=是）
	Status     int8   `json:"status" example:"1"`                           // 状态（0=禁用 1=启用）
	Remark     string `json:"remark" example:""`                             // 备注
}

// DictDataListRequest 字典数据列表查询
type DictDataListRequest struct {
	PageRequest
	DictTypeID *uint64 `json:"dictTypeId" form:"dictTypeId" example:"1"` // 字典类型ID
	Label      string  `json:"label" form:"label" example:"男"`           // 字典标签（模糊查询）
	Status     *int8   `json:"status" form:"status" example:"1"`         // 状态（0=禁用 1=启用）
}

// DictDataResponse 字典数据响应
type DictDataResponse struct {
	ID         uint64 `json:"id" example:"1"`                               // 字典数据ID
	DictTypeID uint64 `json:"dictTypeId" example:"1"`                       // 字典类型ID
	Label      string `json:"label" example:"男"`                            // 字典标签
	Value      string `json:"value" example:"1"`                            // 字典值
	Sort       int    `json:"sort" example:"1"`                             // 排序
	IsDefault  int8   `json:"isDefault" example:"0"`                        // 是否默认（0=否 1=是）
	Status     int8   `json:"status" example:"1"`                           // 状态（0=禁用 1=启用）
	Remark     string `json:"remark" example:""`                             // 备注
	CreatedAt  string `json:"createdAt" example:"2026-07-10 12:00:00"`      // 创建时间
}
