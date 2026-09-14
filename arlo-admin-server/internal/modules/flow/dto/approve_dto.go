package dto

// LaunchProcessItem 可发起流程
type LaunchProcessItem struct {
	ProcessID      uint64 `json:"processId"`
	CategoryID     uint64 `json:"categoryId"`
	CategoryName   string `json:"categoryName"`
	ProcessKey     string `json:"processKey"`
	ProcessName    string `json:"processName"`
	ProcessIcon    string `json:"processIcon"`
	ProcessBgcolor string `json:"processBgcolor"`
	ProcessType    string `json:"processType"`
	Remark         string `json:"remark"`
}

// LaunchFormDetail 发起表单详情
type LaunchFormDetail struct {
	ProcessID      uint64                 `json:"processId"`
	ProcessKey     string                 `json:"processKey"`
	ProcessName    string                 `json:"processName"`
	ProcessType    string                 `json:"processType"`
	ProcessForm    map[string]interface{} `json:"processForm"`
	ModelContent   map[string]interface{} `json:"modelContent"`
	SelectionNodes []LaunchSelectionNode  `json:"selectionNodes"`
	// FormRenderType designer=epic设计表单 vue=系统自定义页面
	FormRenderType string `json:"formRenderType"`
	FormComponent  string `json:"formComponent,omitempty"` // vue 时：相对 views 的组件路径
}

// LaunchSelectionNode 发起时需自选的审批/抄送节点
type LaunchSelectionNode struct {
	NodeKey    string `json:"nodeKey"`
	NodeName   string `json:"nodeName"`
	Kind       string `json:"kind"` // assignee | cc
	SelectMode int    `json:"selectMode"`
	Required   bool   `json:"required"`
}

// LaunchSelectionPayload 暂存回填的自选人
type LaunchSelectionPayload struct {
	Assignees map[string][]IDName `json:"assignees"`
	CCUsers   map[string][]IDName `json:"ccUsers"`
}

// RejectTarget 可驳回目标节点
type RejectTarget struct {
	NodeKey  string `json:"nodeKey"`
	NodeName string `json:"nodeName"`
}

// LaunchRequest 发起审批
type LaunchRequest struct {
	ProcessID uint64                 `json:"processId" binding:"required"`
	FormData  map[string]interface{} `json:"formData"`
	// 节点自选审批人 nodeKey -> [{id,name}]
	NodeAssignees map[string][]IDName `json:"nodeAssignees"`
	// 节点自选抄送
	NodeCCUsers map[string][]IDName `json:"nodeCcUsers"`
	// 暂存待审（不进入审批，instanceState=-1）
	SaveAsDraft bool `json:"saveAsDraft"`
}

// ActivateDraftRequest 暂存单正式发起
type ActivateDraftRequest struct {
	InstanceID    uint64                 `json:"instanceId" binding:"required"`
	FormData      map[string]interface{} `json:"formData"`
	NodeAssignees map[string][]IDName    `json:"nodeAssignees"`
	NodeCCUsers   map[string][]IDName    `json:"nodeCcUsers"`
}

// UpdateDraftRequest 更新暂存单（仍保持暂存，不进入审批）
type UpdateDraftRequest struct {
	InstanceID    uint64                 `json:"instanceId" binding:"required"`
	FormData      map[string]interface{} `json:"formData"`
	NodeAssignees map[string][]IDName    `json:"nodeAssignees"`
	NodeCCUsers   map[string][]IDName    `json:"nodeCcUsers"`
}

// DeleteDraftRequest 删除暂存单
type DeleteDraftRequest struct {
	InstanceID uint64 `json:"instanceId" binding:"required"`
}

type IDName struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// ClaimRequest 认领
type ClaimRequest struct {
	TaskID uint64 `json:"taskId" binding:"required"`
}

// ConsentRequest 同意
type ConsentRequest struct {
	TaskID   uint64                 `json:"taskId" binding:"required"`
	Opinion  string                 `json:"opinion"`
	FormData map[string]interface{} `json:"formData"`
}

// ResubmitRequest 发起人改单后重新提交
type ResubmitRequest struct {
	TaskID   uint64                 `json:"taskId" binding:"required"`
	FormData map[string]interface{} `json:"formData"`
}

// RejectRequest 拒绝/驳回
type RejectRequest struct {
	TaskID         uint64                 `json:"taskId" binding:"required"`
	Opinion        string                 `json:"opinion"`
	FormData       map[string]interface{} `json:"formData"`
	RejectStrategy int                    `json:"rejectStrategy"` // 0=用节点配置
	RejectNodeKey  string                 `json:"rejectNodeKey"`
}

// BatchConsentRequest 批量同意
type BatchConsentRequest struct {
	TaskIDs []uint64 `json:"taskIds" binding:"required"`
	Opinion string   `json:"opinion"`
}

// BatchRejectRequest 批量驳回（默认按各节点策略；也可强制终止）
type BatchRejectRequest struct {
	TaskIDs        []uint64 `json:"taskIds" binding:"required"`
	Opinion        string   `json:"opinion"`
	RejectStrategy int      `json:"rejectStrategy"` // 0=节点配置（策略3且未预置目标会失败）；4=统一终止
}

// BatchOperateResult 批量结果
type BatchOperateResult struct {
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors"`
}

// DelegateSetting 我的委托设置
type DelegateSetting struct {
	ToUserID   uint64 `json:"toUserId"`
	ToUserName string `json:"toUserName"`
	Enabled    bool   `json:"enabled"`
	Remark     string `json:"remark"`
}

// SetDelegateRequest 设置委托
type SetDelegateRequest struct {
	ToUserID   uint64 `json:"toUserId"`
	ToUserName string `json:"toUserName"`
	Enabled    bool   `json:"enabled"`
	Remark     string `json:"remark"`
	Clear      bool   `json:"clear"` // true=清除委托
}

// RevokeRequest 撤销
type RevokeRequest struct {
	InstanceID uint64 `json:"instanceId" binding:"required"`
	Opinion    string `json:"opinion"`
}

// TransferRequest 转交 / 管理员转办
type TransferRequest struct {
	TaskID      uint64 `json:"taskId" binding:"required"`
	ToUserID    uint64 `json:"toUserId" binding:"required"`
	ToName      string `json:"toName"`
	Opinion     string `json:"opinion"`
	FromActorID uint64 `json:"fromActorId"` // 管理员转办：指定待办槽位（flow_task_actor.id）；仅 1 人时可空
}

// RollbackRequest 回退
type RollbackRequest struct {
	TaskID        uint64 `json:"taskId" binding:"required"`
	Opinion       string `json:"opinion"`
	TargetNodeKey string `json:"targetNodeKey"` // 空=上一审批节点
}

// UrgeRequest 催办
type UrgeRequest struct {
	InstanceID uint64 `json:"instanceId" binding:"required"`
}

// TerminateRequest 管理员终止
type TerminateRequest struct {
	InstanceID uint64 `json:"instanceId" binding:"required"`
	Opinion    string `json:"opinion"`
}

// AppendRequest 加签
type AppendRequest struct {
	TaskID   uint64 `json:"taskId" binding:"required"`
	ToUserID uint64 `json:"toUserId" binding:"required"`
	ToName   string `json:"toName"`
	Position int    `json:"position"` // 1前加签 2后加签，默认2
}

// RemoveActorRequest 减签
type RemoveActorRequest struct {
	TaskID  uint64 `json:"taskId" binding:"required"`
	ActorID uint64 `json:"actorId" binding:"required"` // flow_task_actor.id
}

// RuntimeCCRequest 运行时补抄送
type RuntimeCCRequest struct {
	TaskID uint64   `json:"taskId" binding:"required"`
	Users  []IDName `json:"users" binding:"required,min=1"`
}

// CommentRequest 发表评论
type CommentRequest struct {
	InstanceID     uint64   `json:"instanceId" binding:"required"`
	Content        string   `json:"content" binding:"required,max=500"`
	MentionUserIDs []uint64 `json:"mentionUserIds"`
}

// CommentItem 评论
type CommentItem struct {
	ID        uint64 `json:"id"`
	UserID    uint64 `json:"userId"`
	UserName  string `json:"userName"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

// PendingTaskItem 待办
type PendingTaskItem struct {
	ActorID          uint64 `json:"actorId"`
	TaskID           uint64 `json:"taskId"`
	InstanceID       uint64 `json:"instanceId"` // 办理实例（子流程待办时为子实例）
	ParentInstanceID uint64 `json:"parentInstanceId,omitempty"`
	NodeName         string `json:"nodeName"` // 展示：子流程时「子流程 · 审核人」
	WorkNodeName     string `json:"workNodeName,omitempty"`
	NodeKey          string `json:"nodeKey"`
	ExamineMode      int8   `json:"examineMode"`
	ProcessName      string `json:"processName"` // 展示名：子实例时用主流程名
	SubProcessName   string `json:"subProcessName,omitempty"`
	IsSubProcess     bool   `json:"isSubProcess,omitempty"`
	CreateBy         string `json:"createBy"`
	CreateID         uint64 `json:"createId"`
	CreatedAt        string `json:"createdAt"`
	CanHandle        bool   `json:"canHandle"`
	AllowBatchOperate bool  `json:"allowBatchOperate"`
	IsDelegate       bool   `json:"isDelegate"` // 代他人审批
	ActorUserID      uint64 `json:"actorUserId,omitempty"`
}

// InstanceDetail 详情
type InstanceDetail struct {
	InstanceID       uint64                   `json:"instanceId"` // 主流程视角时为根实例
	WorkInstanceID   uint64                   `json:"workInstanceId,omitempty"`
	ParentInstanceID uint64                   `json:"parentInstanceId,omitempty"`
	ProcessID        uint64                   `json:"processId"`
	ProcessName      string                   `json:"processName"` // 展示名：子实例时用主流程名
	SubProcessName   string                   `json:"subProcessName,omitempty"`
	IsSubProcess     bool                     `json:"isSubProcess,omitempty"`
	ProcessKey       string                   `json:"processKey"`
	ProcessVersion   int                      `json:"processVersion"`
	InstanceState    int8                     `json:"instanceState"`
	CurrentNodeKey   string                   `json:"currentNodeKey"`
	CurrentNodeName  string                   `json:"currentNodeName"`
	WorkNodeKey      string                   `json:"workNodeKey,omitempty"`
	WorkNodeName     string                   `json:"workNodeName,omitempty"`
	CreateID         uint64                   `json:"createId"`
	CreateBy         string                   `json:"createBy"`
	CreatedAt        string                   `json:"createdAt"`
	FinishTime       string                   `json:"finishTime"`
	FormSchema       map[string]interface{}   `json:"formSchema"`
	FormData         map[string]interface{}   `json:"formData"`
	FormConfig       []map[string]interface{} `json:"formConfig"`
	FormRenderType   string                   `json:"formRenderType"`             // designer | vue
	FormComponent    string                   `json:"formComponent,omitempty"`    // vue 组件路径
	ModelContent     map[string]interface{}   `json:"modelContent"`
	TaskID           uint64                   `json:"taskId"`
	CanConsent       bool                     `json:"canConsent"`
	CanReject        bool                     `json:"canReject"`
	CanResubmit      bool                     `json:"canResubmit"` // 发起人改单重提
	CanTransfer      bool                     `json:"canTransfer"`
	CanAppend        bool                     `json:"canAppend"`
	CanRemove        bool                     `json:"canRemove"` // 减签（仅加签人）
	CanRollback      bool                     `json:"canRollback"`
	CanRevoke        bool                     `json:"canRevoke"`
	CanUrge          bool                     `json:"canUrge"`
	CanTerminate     bool                     `json:"canTerminate"`     // 流程监控：管理员终止
	CanAdminTransfer bool                     `json:"canAdminTransfer"` // 流程监控：管理员转办
	CanCc            bool                     `json:"canCc"`            // 运行时补抄送
	CanComment       bool                     `json:"canComment"`       // 可评论
	CanActivateDraft bool                     `json:"canActivateDraft"` // 暂存：正式发起
	CanUpdateDraft   bool                     `json:"canUpdateDraft"`   // 暂存：再次保存
	CanDeleteDraft   bool                     `json:"canDeleteDraft"`   // 暂存：删除
	SelectionNodes   []LaunchSelectionNode    `json:"selectionNodes,omitempty"`
	LaunchSelection  *LaunchSelectionPayload  `json:"launchSelection,omitempty"`
	SecondOperatePrompt bool                  `json:"secondOperatePrompt"` // 关键操作二次确认
	AllowDelegate    bool                     `json:"allowDelegate"` // 流程是否允许委托代批
	TaskNodeType     int                      `json:"taskNodeType"`
	RejectStrategy   int                      `json:"rejectStrategy"`
	RejectStart      int                      `json:"rejectStart"`
	RejectNodeKey    string                   `json:"rejectNodeKey"` // 节点预配置的指定驳回目标
	RejectTargets    []RejectTarget           `json:"rejectTargets"`
	RollbackTargets  []RejectTarget           `json:"rollbackTargets"`
	AllowTransfer    bool                     `json:"allowTransfer"`
	AllowAppendNode  bool                     `json:"allowAppendNode"`
	AllowRollback    bool                     `json:"allowRollback"`
	AllowCc          bool                     `json:"allowCc"`
	TaskActors       []TaskActorItem          `json:"taskActors"`
	Timeline         []TimelineItem           `json:"timeline"`
	Comments         []CommentItem            `json:"comments"`
}

// TaskActorItem 当前任务参与人（减签用）
type TaskActorItem struct {
	ID         uint64 `json:"id"`
	ActorID    uint64 `json:"actorId"`
	ActorName  string `json:"actorName"`
	ActorType  int8   `json:"actorType"`
	ActorState int8   `json:"actorState"`
	Weight     int    `json:"weight"`
}

type TimelineItem struct {
	NodeName   string `json:"nodeName"`
	NodeKey    string `json:"nodeKey"`
	NodeType   int    `json:"nodeType"`
	TaskState  int8   `json:"taskState"`
	ActorName  string `json:"actorName"`
	AgentName  string `json:"agentName,omitempty"` // 代批人姓名（有值时前端展示「委托人（受托人代批）」）
	ActorState int8   `json:"actorState"`
	Opinion    string `json:"opinion"`
	FinishTime string `json:"finishTime"`
	CreatedAt  string `json:"createdAt"`
	// 子流程节点：关联子实例，供主流程流转记录点开查看
	ChildInstanceID    uint64 `json:"childInstanceId,omitempty"`
	ChildProcessName   string `json:"childProcessName,omitempty"`
	ChildInstanceState int8   `json:"childInstanceState,omitempty"`
}

// InstanceListItem 我的申请 / 已审批 / 抄送等列表项
type InstanceListItem struct {
	// ID 历史参与人行 id（已办/抄送列表唯一键；我的申请等可为空）
	ID              uint64 `json:"id,omitempty"`
	InstanceID      uint64 `json:"instanceId"`
	ProcessName     string `json:"processName"` // 展示名：子实例时用主流程名
	SubProcessName  string `json:"subProcessName,omitempty"`
	IsSubProcess    bool   `json:"isSubProcess,omitempty"`
	ProcessKey      string `json:"processKey"`
	CurrentNodeName string `json:"currentNodeName"`
	InstanceState   int8   `json:"instanceState"`
	CreateBy        string `json:"createBy"`
	CreatedAt       string `json:"createdAt"`
	FinishTime      string `json:"finishTime"`
	NodeName        string `json:"nodeName"`
	TaskID          uint64 `json:"taskId"`
	HisTaskID       uint64 `json:"hisTaskId,omitempty"`
	ActorState      int8   `json:"actorState"`
	IsDelegate      bool   `json:"isDelegate,omitempty"` // 当前用户作为受托人代批完成
	Read            bool   `json:"read"`                 // 抄送是否已读（仅我收到的；其它列表恒 false）
}
