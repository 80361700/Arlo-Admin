package dto

// --- 请求 DTO ---

// CreateConfigRequest 创建配置
type CreateConfigRequest struct {
	Name   string `json:"name" binding:"required" example:"系统名称"`            // 配置名称
	Key    string `json:"key" binding:"required" example:"sys.name"`           // 配置键（唯一）
	Value  string `json:"value" example:"Arlo Admin"`                          // 配置值（图片可空；开关默认 false）
	Type   int8   `json:"type" binding:"required,oneof=1 2 3 4" example:"1"`   // 类型: 1=文本,2=JSON,3=开关,4=图片
	Remark string `json:"remark" example:"系统显示名称"`                             // 备注
}

// UpdateConfigRequest 更新配置
type UpdateConfigRequest struct {
	ID     uint64 `json:"id" binding:"required" example:"1"`                   // 配置ID
	Name   string `json:"name" binding:"required" example:"系统名称"`              // 配置名称
	Key    string `json:"key" binding:"required" example:"sys.name"`           // 配置键（唯一）
	Value  string `json:"value" example:"Arlo Admin"`                          // 配置值
	Type   int8   `json:"type" binding:"required,oneof=1 2 3 4" example:"1"`   // 类型: 1=文本,2=JSON,3=开关,4=图片
	Remark string `json:"remark" example:"系统显示名称"`                             // 备注
}

// PublicConfigResponse 公开配置（登录页等免鉴权场景）
type PublicConfigResponse struct {
	Name    string `json:"name" example:"Arlo Admin"` // 系统名称 sys.name
	Captcha bool   `json:"captcha" example:"false"`   // 是否开启验证码 sys.captcha
	Logo    string `json:"logo" example:""`           // 系统 Logo sys.logo
	Version string `json:"version" example:"1.0.0"`   // 系统版本 sys.version
}

// ConfigListQuery 配置列表查询
type ConfigListQuery struct {
	Name string `form:"name" example:"系统"`   // 配置名称（模糊搜索）
	Key  string `form:"key" example:"sys"`    // 配置键（模糊搜索）
	Type int8   `form:"type" example:"1"`     // 配置类型
}

// --- 响应 DTO ---

// ConfigResponse 配置详情
type ConfigResponse struct {
	ID        uint64 `json:"id" example:"1"`                              // 配置ID
	Name      string `json:"name" example:"系统名称"`                         // 配置名称
	Key       string `json:"key" example:"sys.name"`                      // 配置键
	Value     string `json:"value" example:"Arlo Admin"`                  // 配置值
	Type      int8   `json:"type" example:"1"`                            // 类型: 1=文本,2=JSON,3=开关,4=图片
	Remark    string `json:"remark" example:"系统显示名称"`                     // 备注
	CreatedAt string `json:"createdAt"`                                   // 创建时间
	UpdatedAt string `json:"updatedAt"`                                   // 更新时间
}
