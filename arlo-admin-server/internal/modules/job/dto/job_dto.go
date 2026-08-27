package dto

// JobListQuery 任务列表
type JobListQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Name     string `form:"name"`
	Handler  string `form:"handler"`
	Status   *int8  `form:"status"`
}

// CreateJobRequest 创建
type CreateJobRequest struct {
	Name    string `json:"name" binding:"required,max=64"`
	Handler string `json:"handler" binding:"required,max=64"`
	Cron    string `json:"cron" binding:"required,max=64"`
	Params  string `json:"params" binding:"omitempty,max=512"`
	Status  *int8  `json:"status"`
	Remark  string `json:"remark" binding:"omitempty,max=255"`
}

// UpdateJobRequest 更新
type UpdateJobRequest struct {
	Name   string `json:"name" binding:"required,max=64"`
	Cron   string `json:"cron" binding:"required,max=64"`
	Params string `json:"params" binding:"omitempty,max=512"`
	Remark string `json:"remark" binding:"omitempty,max=255"`
}

// UpdateJobStatusRequest 启停
type UpdateJobStatusRequest struct {
	Status int8 `json:"status" binding:"oneof=0 1"`
}

// JobLogQuery 执行日志
type JobLogQuery struct {
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PageSize    int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	JobID       *uint64 `form:"jobId"`
	Status      *int8  `form:"status"`
	TriggerType *int8  `form:"triggerType"`
}

// JobResponse 任务
type JobResponse struct {
	ID         uint64  `json:"id"`
	Name       string  `json:"name"`
	Handler    string  `json:"handler"`
	Cron       string  `json:"cron"`
	Params     string  `json:"params"`
	Status     int8    `json:"status"`
	Remark     string  `json:"remark"`
	LastRunAt  string  `json:"lastRunAt"`
	LastStatus *int8   `json:"lastStatus"`
	NextRunAt  string  `json:"nextRunAt"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
}

// JobListResponse 列表
type JobListResponse struct {
	List     []JobResponse `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

// JobLogResponse 日志
type JobLogResponse struct {
	ID          uint64 `json:"id"`
	JobID       uint64 `json:"jobId"`
	JobName     string `json:"jobName"`
	Handler     string `json:"handler"`
	TriggerType int8   `json:"triggerType"`
	Status      int8   `json:"status"`
	Result      string `json:"result"`
	ErrorMsg    string `json:"errorMsg"`
	DurationMs  int    `json:"durationMs"`
	CreatedAt   string `json:"createdAt"`
}

// JobLogListResponse 日志列表
type JobLogListResponse struct {
	List     []JobLogResponse `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}
