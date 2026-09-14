package dto

// CategoryListItem 分类（含流程列表，供左侧树）
type CategoryListItem struct {
	CategoryID     uint64          `json:"categoryId"`
	CategoryName   string          `json:"categoryName"`
	CategoryRemark string          `json:"categoryRemark"`
	CategorySort   int             `json:"categorySort"`
	ProcessList    []ProcessBrief  `json:"processList"`
}

// CategoryOption 下拉
type CategoryOption struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// CreateCategoryRequest 创建分类
type CreateCategoryRequest struct {
	Name   string `json:"name" binding:"required,max=64"`
	Sort   int    `json:"sort"`
	Remark string `json:"remark" binding:"omitempty,max=255"`
}

// UpdateCategoryRequest 更新分类
type UpdateCategoryRequest struct {
	ID     uint64 `json:"id" binding:"required"`
	Name   string `json:"name" binding:"required,max=64"`
	Sort   int    `json:"sort"`
	Remark string `json:"remark" binding:"omitempty,max=255"`
}

// ProcessOption 子流程等下拉选项
type ProcessOption struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

// ProcessBrief 列表简要
type ProcessBrief struct {
	ProcessID      uint64 `json:"processId"`
	CategoryID     uint64 `json:"categoryId"`
	ProcessKey     string `json:"processKey"`
	ProcessName    string `json:"processName"`
	ProcessIcon    string `json:"processIcon"`
	ProcessType    string `json:"processType"`
	ProcessVersion int    `json:"processVersion"`
	ProcessState   int8   `json:"processState"`
	Remark         string `json:"remark"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// ProcessDetail 详情（设计器用）
type ProcessDetail struct {
	ProcessID             uint64                   `json:"processId"`
	CategoryID            uint64                   `json:"categoryId"`
	ProcessKey            string                   `json:"processKey"`
	ProcessName           string                   `json:"processName"`
	ProcessIcon           string                   `json:"processIcon"`
	ProcessBgcolor        string                   `json:"processBgcolor"`
	ProcessType           string                   `json:"processType"`
	ProcessVersion        int                      `json:"processVersion"`
	ProcessState          int8                     `json:"processState"`
	Remark                string                   `json:"remark"`
	ModelContent          map[string]interface{}   `json:"modelContent"`
	ProcessForm           map[string]interface{}   `json:"processForm"`
	ProcessSetting        map[string]interface{}   `json:"processSetting"`
	ProcessPermissionList []map[string]interface{} `json:"processPermissionList"`
	CreatedBy             uint64                   `json:"createdBy"`
	CreatedAt             string                   `json:"createdAt"`
	UpdatedAt             string                   `json:"updatedAt"`
}

// SaveProcessRequest 创建/更新流程（与设计器表单对齐）
type SaveProcessRequest struct {
	ProcessID             *uint64                  `json:"processId"`
	CategoryID            uint64                   `json:"categoryId" binding:"required"`
	ProcessKey            string                   `json:"processKey" binding:"omitempty,max=64"`
	ProcessName           string                   `json:"processName" binding:"required,max=128"`
	ProcessIcon           string                   `json:"processIcon"` // 可为纯图标名，或 JSON
	ProcessBgcolor        string                   `json:"processBgcolor"`
	ProcessType           string                   `json:"processType"`
	Remark                string                   `json:"remark" binding:"omitempty,max=255"`
	ModelContent          interface{}              `json:"modelContent" binding:"required"`
	ProcessForm           interface{}              `json:"processForm"`
	ProcessSetting        interface{}              `json:"processSetting"`
	ProcessPermissionList []map[string]interface{} `json:"processPermissionList"`
}

// UpdateProcessStateRequest 启停
type UpdateProcessStateRequest struct {
	State int8 `json:"state" binding:"oneof=0 1"`
}

// CheckProcessResult 启用前配置体检
type CheckProcessResult struct {
	Ok      bool     `json:"ok"`
	Message string   `json:"message"`
	Issues  []string `json:"issues,omitempty"`
}

// ---------- 表单模板 ----------

// FormCategoryListItem 表单分类（含模板列表）
type FormCategoryListItem struct {
	CategoryID     uint64      `json:"categoryId"`
	CategoryName   string      `json:"categoryName"`
	CategoryRemark string      `json:"categoryRemark"`
	CategorySort   int         `json:"categorySort"`
	FormList       []FormBrief `json:"formList"`
}

// FormCategoryOption 表单分类下拉
type FormCategoryOption struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// CreateFormCategoryRequest 创建表单分类
type CreateFormCategoryRequest struct {
	Name   string `json:"name" binding:"required,max=64"`
	Sort   int    `json:"sort"`
	Remark string `json:"remark" binding:"omitempty,max=255"`
}

// UpdateFormCategoryRequest 更新表单分类
type UpdateFormCategoryRequest struct {
	ID     uint64 `json:"id" binding:"required"`
	Name   string `json:"name" binding:"required,max=64"`
	Sort   int    `json:"sort"`
	Remark string `json:"remark" binding:"omitempty,max=255"`
}

// FormBrief 列表简要
type FormBrief struct {
	FormID     uint64 `json:"formId"`
	CategoryID uint64 `json:"categoryId"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	FormType   int8   `json:"formType"` // 1设计 2系统
	PcURL      string `json:"pcUrl"`
	Status     int8   `json:"status"`
	Sort       int    `json:"sort"`
	Remark     string `json:"remark"`
	Bound      bool   `json:"bound"` // 是否被流程节点引用
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// FormDetail 详情（含 schema）
type FormDetail struct {
	FormID     uint64                 `json:"formId"`
	CategoryID uint64                 `json:"categoryId"`
	Name       string                 `json:"name"`
	Code       string                 `json:"code"`
	FormType   int8                   `json:"formType"`
	PcURL      string                 `json:"pcUrl"`
	Status     int8                   `json:"status"`
	Sort       int                    `json:"sort"`
	Remark     string                 `json:"remark"`
	FormSchema map[string]interface{} `json:"formSchema"`
	Bound      bool                   `json:"bound"`
	CreatedAt  string                 `json:"createdAt"`
	UpdatedAt  string                 `json:"updatedAt"`
}

// FormOption 节点选子表单等下拉
type FormOption struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	FormType int8   `json:"formType"`
	PcURL    string `json:"pcUrl"`
}

// SaveFormRequest 创建/更新表单元数据（不含 schema 大字段时可空）
type SaveFormRequest struct {
	FormID     *uint64 `json:"formId"`
	CategoryID uint64  `json:"categoryId" binding:"required"`
	Name       string  `json:"name" binding:"required,max=128"`
	Code       string  `json:"code" binding:"required,max=64"`
	FormType   int8    `json:"formType" binding:"omitempty,oneof=1 2"` // 缺省 1
	PcURL      string  `json:"pcUrl" binding:"omitempty,max=255"`
	Status     *int8   `json:"status" binding:"omitempty,oneof=0 1"`
	Sort       int     `json:"sort"`
	Remark     string  `json:"remark" binding:"omitempty,max=255"`
}

// UpdateFormStateRequest 启停
type UpdateFormStateRequest struct {
	Status int8 `json:"status" binding:"oneof=0 1"`
}

// SaveFormSchemaRequest 保存设计器 JSON
type SaveFormSchemaRequest struct {
	FormSchema interface{} `json:"formSchema" binding:"required"`
}

// ---------- 流程历史版本 ----------

// ProcessHistoryBrief 历史列表项
type ProcessHistoryBrief struct {
	HistoryID      uint64 `json:"historyId"`
	ProcessID      uint64 `json:"processId"`
	ProcessName    string `json:"processName"`
	ProcessIcon    string `json:"processIcon"`
	ProcessVersion int    `json:"processVersion"`
	Remark         string `json:"remark"`
	CreatedAt      string `json:"createdAt"`
}

// ProcessHistoryDetail 历史详情（预览/迁出用）
type ProcessHistoryDetail struct {
	HistoryID             uint64                   `json:"historyId"`
	ProcessID             uint64                   `json:"processId"`
	CategoryID            uint64                   `json:"categoryId"`
	ProcessKey            string                   `json:"processKey"`
	ProcessName           string                   `json:"processName"`
	ProcessIcon           string                   `json:"processIcon"`
	ProcessBgcolor        string                   `json:"processBgcolor"`
	ProcessType           string                   `json:"processType"`
	ProcessVersion        int                      `json:"processVersion"`
	ProcessState          int8                     `json:"processState"`
	Remark                string                   `json:"remark"`
	ModelContent          map[string]interface{}   `json:"modelContent"`
	ProcessForm           map[string]interface{}   `json:"processForm"`
	ProcessSetting        map[string]interface{}   `json:"processSetting"`
	ProcessPermissionList []map[string]interface{} `json:"processPermissionList"`
	CreatedAt             string                   `json:"createdAt"`
}
