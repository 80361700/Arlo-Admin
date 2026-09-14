package model

import "time"

// ---------- 运行时状态常量 ----------

const (
	InstDraft     int8 = -1 // 暂存待审
	InstActive    int8 = 0  // 审批中
	InstComplete  int8 = 1  // 通过
	InstReject    int8 = 2  // 拒绝
	InstRevoke    int8 = 3  // 撤销
	InstTerminate int8 = 4  // 终止
	InstTimeout   int8 = 5  // 超时
)

const (
	TaskMajor  int8 = 0
	TaskCC     int8 = 4
	TaskDelay  int8 = 5
	TaskTrig   int8 = 6
	TaskSub    int8 = 7
)

const (
	TaskActive     int8 = 0
	TaskComplete   int8 = 1
	TaskReject     int8 = 2
	TaskRevoke     int8 = 3
	TaskTerminate  int8 = 4
	TaskSkip       int8 = 5
)

const (
	ActorPending  int8 = 0
	ActorAgree    int8 = 1
	ActorRefuse   int8 = 2
	ActorTransfer int8 = 3
	ActorSkip     int8 = 4
)

// 参与人类型
const (
	ActorTypeUser  int8 = 0 // 普通审批人
	ActorTypeAppend int8 = 1 // 加签
	ActorTypeRole  int8 = 2 // 角色认领池（groupStrategy=0）
)

// FlowInstance 流程实例
type FlowInstance struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProcessID        uint64     `gorm:"not null;default:0;index" json:"processId"`
	ProcessKey       string     `gorm:"size:64;not null;default:''" json:"processKey"`
	ProcessName      string     `gorm:"size:128;not null;default:''" json:"processName"`
	ProcessVersion   int        `gorm:"not null;default:1" json:"processVersion"`
	ProcessType      string     `gorm:"size:32;not null;default:main" json:"processType"`
	ModelContent     string     `gorm:"type:longtext" json:"modelContent"`
	ProcessForm      string     `gorm:"type:longtext" json:"processForm"`
	FormData         string     `gorm:"type:longtext" json:"formData"`
	CurrentNodeKey   string     `gorm:"size:64;not null;default:''" json:"currentNodeKey"`
	CurrentNodeName  string     `gorm:"size:128;not null;default:''" json:"currentNodeName"`
	InstanceState    int8       `gorm:"not null;default:0;index" json:"instanceState"`
	CreateID         uint64     `gorm:"not null;default:0;index" json:"createId"`
	CreateBy         string     `gorm:"size:64;not null;default:''" json:"createBy"`
	CreateDeptID     uint64     `gorm:"not null;default:0" json:"createDeptId"`
	ParentInstanceID uint64     `gorm:"not null;default:0;index" json:"parentInstanceId"`
	ParentNodeKey    string     `gorm:"size:64;not null;default:''" json:"parentNodeKey"`
	JoinGateKey      string     `gorm:"size:64;not null;default:''" json:"joinGateKey"`
	Variable         string     `gorm:"type:text" json:"variable"`
	FinishTime       *time.Time `json:"finishTime"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func (FlowInstance) TableName() string { return "flow_instance" }

// FlowTask 活动任务
type FlowTask struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	InstanceID   uint64     `gorm:"not null;default:0;index" json:"instanceId"`
	NodeKey      string     `gorm:"size:64;not null;default:''" json:"nodeKey"`
	NodeName     string     `gorm:"size:128;not null;default:''" json:"nodeName"`
	NodeType     int        `gorm:"not null;default:0" json:"nodeType"`
	TaskType     int8       `gorm:"not null;default:0" json:"taskType"`
	TaskState    int8       `gorm:"not null;default:0;index" json:"taskState"`
	ExamineMode  int8       `gorm:"not null;default:1" json:"examineMode"`
	ParentTaskID uint64     `gorm:"not null;default:0" json:"parentTaskId"`
	FromNodeKey  string     `gorm:"size:64;not null;default:''" json:"fromNodeKey"`
	GateToken    string     `gorm:"size:64;not null;default:''" json:"gateToken"`
	ExpireTime   *time.Time `json:"expireTime"`
	Payload      string     `gorm:"type:text" json:"payload"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (FlowTask) TableName() string { return "flow_task" }

// FlowTaskActor 任务参与人
type FlowTaskActor struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID     uint64     `gorm:"not null;default:0;index" json:"taskId"`
	InstanceID uint64     `gorm:"not null;default:0;index" json:"instanceId"`
	ActorID    uint64     `gorm:"not null;default:0;index" json:"actorId"`
	ActorName  string     `gorm:"size:64;not null;default:''" json:"actorName"`
	ActorType  int8       `gorm:"not null;default:0" json:"actorType"`
	ActorState int8       `gorm:"not null;default:0" json:"actorState"`
	Weight     int        `gorm:"not null;default:0" json:"weight"`
	AgentID    uint64     `gorm:"not null;default:0" json:"agentId"`
	Opinion    string     `gorm:"size:1000;not null;default:''" json:"opinion"`
	FinishTime *time.Time `json:"finishTime"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

func (FlowTaskActor) TableName() string { return "flow_task_actor" }

// FlowHisTask 历史任务
type FlowHisTask struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID      uint64     `gorm:"not null;default:0;index" json:"taskId"`
	InstanceID  uint64     `gorm:"not null;default:0;index" json:"instanceId"`
	NodeKey     string     `gorm:"size:64;not null;default:''" json:"nodeKey"`
	NodeName    string     `gorm:"size:128;not null;default:''" json:"nodeName"`
	NodeType    int        `gorm:"not null;default:0" json:"nodeType"`
	TaskType    int8       `gorm:"not null;default:0" json:"taskType"`
	TaskState   int8       `gorm:"not null;default:0" json:"taskState"`
	ExamineMode int8       `gorm:"not null;default:1" json:"examineMode"`
	FromNodeKey string     `gorm:"size:64;not null;default:''" json:"fromNodeKey"`
	GateToken   string     `gorm:"size:64;not null;default:''" json:"gateToken"`
	Payload     string     `gorm:"type:text" json:"payload"`
	CreatedAt   time.Time  `json:"createdAt"`
	FinishTime  *time.Time `json:"finishTime"`
}

func (FlowHisTask) TableName() string { return "flow_his_task" }

// FlowHisTaskActor 历史参与人
type FlowHisTaskActor struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	HisTaskID  uint64     `gorm:"not null;default:0" json:"hisTaskId"`
	TaskID     uint64     `gorm:"not null;default:0" json:"taskId"`
	InstanceID uint64     `gorm:"not null;default:0;index" json:"instanceId"`
	ActorID    uint64     `gorm:"not null;default:0;index" json:"actorId"`
	ActorName  string     `gorm:"size:64;not null;default:''" json:"actorName"`
	ActorType  int8       `gorm:"not null;default:0" json:"actorType"`
	ActorState int8       `gorm:"not null;default:0" json:"actorState"`
	Weight     int        `gorm:"not null;default:0" json:"weight"`
	AgentID    uint64     `gorm:"not null;default:0" json:"agentId"`
	Opinion    string     `gorm:"size:1000;not null;default:''" json:"opinion"`
	FinishTime *time.Time `json:"finishTime"`
	CreatedAt  time.Time  `json:"createdAt"`
	ReadAt     *time.Time `json:"readAt"` // 抄送已读时间（仅 TaskCC 使用）
}

func (FlowHisTaskActor) TableName() string { return "flow_his_task_actor" }

// FlowInstanceComment 流程实例评论
type FlowInstanceComment struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	InstanceID uint64    `gorm:"not null;default:0;index" json:"instanceId"`
	UserID     uint64    `gorm:"not null;default:0" json:"userId"`
	UserName   string    `gorm:"size:64;not null;default:''" json:"userName"`
	Content    string    `gorm:"size:500;not null;default:''" json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (FlowInstanceComment) TableName() string { return "flow_instance_comment" }

// FlowUserDelegate 审批委托（一人仅一条启用委托）
type FlowUserDelegate struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"not null;default:0;uniqueIndex" json:"userId"`
	ToUserID   uint64    `gorm:"not null;default:0;index" json:"toUserId"`
	ToUserName string    `gorm:"size:64;not null;default:''" json:"toUserName"`
	Enabled    int8      `gorm:"not null;default:1" json:"enabled"`
	Remark     string    `gorm:"size:255;not null;default:''" json:"remark"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (FlowUserDelegate) TableName() string { return "flow_user_delegate" }
