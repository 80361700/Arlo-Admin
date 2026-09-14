package dto

import flowdto "arlo-admin/internal/modules/flow/dto"

// ListQuery 分页查询
type ListQuery struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=100"`
	Title    string `form:"title"`
	Status   *int8  `form:"status"`
}

// SaveRequest 新增（创建时绑定业务流程）
type SaveRequest struct {
	Title     string `json:"title" binding:"required,max=128"`
	Content   string `json:"content" binding:"required,max=512"`
	ProcessID uint64 `json:"processId" binding:"required"`
}

// LaunchRequest 发起审批（使用单据已绑定流程）
type LaunchRequest struct {
	FormData      map[string]interface{}      `json:"formData"`
	NodeAssignees map[string][]flowdto.IDName `json:"nodeAssignees"`
	NodeCCUsers   map[string][]flowdto.IDName `json:"nodeCcUsers"`
}

// ProcessOption 可选业务流程（创建绑定用：已启用的业务类型流程）
type ProcessOption struct {
	ProcessID   uint64 `json:"processId"`
	ProcessKey  string `json:"processKey"`
	ProcessName string `json:"processName"`
	ProcessType string `json:"processType"`
	Remark      string `json:"remark"`
}

// ProcessPreview 未发起时的流程设计预览
type ProcessPreview struct {
	ProcessID    uint64                 `json:"processId"`
	ProcessKey   string                 `json:"processKey"`
	ProcessName  string                 `json:"processName"`
	ModelContent map[string]interface{} `json:"modelContent"`
}

// OrderItem 列表项
type OrderItem struct {
	ID          uint64 `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Status      int8   `json:"status"`
	InstanceID  uint64 `json:"instanceId"`
	ProcessID   uint64 `json:"processId"`
	ProcessKey  string `json:"processKey"`
	ProcessName string `json:"processName"`
	CanLaunch   bool   `json:"canLaunch"` // 当前用户是否符合设计器「发起人」权限
	CreateBy    string `json:"createBy"`
	CreatedAt   string `json:"createdAt"`
}
