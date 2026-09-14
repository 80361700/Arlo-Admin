package engine

import (
	"encoding/json"
	"fmt"
)

// 设计器节点类型
const (
	NodeStart      = 0
	NodeApproval   = 1
	NodeCC         = 2
	NodeCondition  = 3 // 分支条
	NodeExclusive  = 4
	NodeSubProcess = 5
	NodeDelay      = 6
	NodeTrigger    = 7
	NodeParallel   = 8
	NodeInclusive  = 9
	NodeRoute      = 23
	NodeAutoPass   = 30
	NodeAutoReject = 31
	NodeEnd        = -1
)

// Node 流程节点（与设计器 modelContent.nodeConfig 对齐）
type Node struct {
	NodeName         string                   `json:"nodeName"`
	NodeKey          string                   `json:"nodeKey"`
	Type             int                      `json:"type"`
	ChildNode        *Node                    `json:"childNode"`
	ConditionNodes   []*Node                  `json:"conditionNodes"`
	ParallelNodes    []*Node                  `json:"parallelNodes"`
	InclusiveNodes   []*Node                  `json:"inclusiveNodes"`
	RouteNodes       []*RouteNode             `json:"routeNodes"`
	PriorityLevel    int                      `json:"priorityLevel"`
	ConditionList    [][]ConditionItem        `json:"conditionList"`
	SetType          int                      `json:"setType"`
	NodeAssigneeList []Assignee               `json:"nodeAssigneeList"`
	ExamineLevel     int                      `json:"examineLevel"`
	GroupStrategy    int                      `json:"groupStrategy"`
	SelectMode       int                      `json:"selectMode"`
	NodeCandidate    *NodeCandidate           `json:"nodeCandidate"`
	DirectorMode     int                      `json:"directorMode"`
	DirectorLevel    int                      `json:"directorLevel"`
	ExamineMode      int                      `json:"examineMode"`
	ApproveSelf      int                      `json:"approveSelf"`
	AllowTransfer    bool                     `json:"allowTransfer"`
	AllowAppendNode  bool                     `json:"allowAppendNode"`
	AllowRollback    bool                     `json:"allowRollback"`
	AllowCc          bool                     `json:"allowCc"`
	AllowSelection   bool                     `json:"allowSelection"`
	RejectStrategy   int                      `json:"rejectStrategy"`
	RejectStart      int                      `json:"rejectStart"`
	ActionURL        string                   `json:"actionUrl"`
	ExtendConfig     map[string]interface{}   `json:"extendConfig"`
	SubProcessValue  string                   `json:"subProcessValue"`
	CallProcess      string                   `json:"callProcess"`
	DelayType        int                      `json:"delayType"`
	TriggerType      int                      `json:"triggerType"`
	TermAuto         bool                     `json:"termAuto"`
	Term             int                      `json:"term"`
	TermMode         int                      `json:"termMode"`
	Remind           bool                     `json:"remind"`
}

// RouteNode 路由分支项
type RouteNode struct {
	NodeName      string            `json:"nodeName"`
	NodeKey       string            `json:"nodeKey"` // 目标节点 key
	PriorityLevel int               `json:"priorityLevel"`
	ConditionList [][]ConditionItem `json:"conditionList"`
}

// ConditionItem 条件项
type ConditionItem struct {
	Type      string        `json:"type"` // form | initiator | custom
	Label     string        `json:"label"`
	Field     string        `json:"field"`
	Operator  string        `json:"operator"`
	Value     interface{}   `json:"value"`
	ValueList []OrgMember   `json:"valueList"`
}

// OrgMember 发起人条件选中项
type OrgMember struct {
	ID   interface{} `json:"id"`
	Name string      `json:"name"`
	Kind string      `json:"kind"` // dept | role | user
}

// Assignee 人员/角色
type Assignee struct {
	ID   interface{} `json:"id"`
	Name string      `json:"name"`
}

// NodeCandidate 发起人自选候选
type NodeCandidate struct {
	Type      int        `json:"type"`
	Assignees []Assignee `json:"assignees"`
}

// ModelContent 流程模型根
type ModelContent struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	NodeConfig *Node  `json:"nodeConfig"`
}

// ParseModel 解析模型快照
func ParseModel(raw string) (*ModelContent, error) {
	if raw == "" {
		return nil, fmt.Errorf("流程模型为空")
	}
	var m ModelContent
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, fmt.Errorf("流程模型解析失败: %w", err)
	}
	if m.NodeConfig == nil {
		return nil, fmt.Errorf("流程模型缺少 nodeConfig")
	}
	return &m, nil
}

// FindNode 按 key 查找节点（含各分支）
func FindNode(root *Node, key string) *Node {
	if root == nil || key == "" {
		return nil
	}
	if root.NodeKey == key {
		return root
	}
	if n := FindNode(root.ChildNode, key); n != nil {
		return n
	}
	for _, c := range root.ConditionNodes {
		if n := FindNode(c, key); n != nil {
			return n
		}
		if n := FindNode(c.ChildNode, key); n != nil {
			return n
		}
	}
	for _, c := range root.ParallelNodes {
		if n := FindNode(c, key); n != nil {
			return n
		}
		if n := FindNode(c.ChildNode, key); n != nil {
			return n
		}
	}
	for _, c := range root.InclusiveNodes {
		if n := FindNode(c, key); n != nil {
			return n
		}
		if n := FindNode(c.ChildNode, key); n != nil {
			return n
		}
	}
	return nil
}

// FindParentOf 查找直接挂载 child 的父节点（含 gateway）
func FindParentOf(root *Node, childKey string) *Node {
	if root == nil {
		return nil
	}
	if root.ChildNode != nil && root.ChildNode.NodeKey == childKey {
		return root
	}
	checkBranch := func(branches []*Node) *Node {
		for _, b := range branches {
			if b.NodeKey == childKey {
				return root
			}
			if b.ChildNode != nil && b.ChildNode.NodeKey == childKey {
				return b
			}
			if p := FindParentOf(b.ChildNode, childKey); p != nil {
				return p
			}
			if p := FindParentOf(b, childKey); p != nil && p != b {
				return p
			}
		}
		return nil
	}
	if p := checkBranch(root.ConditionNodes); p != nil {
		return p
	}
	if p := checkBranch(root.ParallelNodes); p != nil {
		return p
	}
	if p := checkBranch(root.InclusiveNodes); p != nil {
		return p
	}
	return FindParentOf(root.ChildNode, childKey)
}

// CollectApprovalPath 从某节点沿 child 链收集可驳回的审批节点（含自身若为审批）
func CollectApprovalNodes(root *Node) []*Node {
	var out []*Node
	var walk func(n *Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		if n.Type == NodeApproval {
			out = append(out, n)
		}
		walk(n.ChildNode)
		for _, c := range n.ConditionNodes {
			walk(c)
			walk(c.ChildNode)
		}
		for _, c := range n.ParallelNodes {
			walk(c)
			walk(c.ChildNode)
		}
		for _, c := range n.InclusiveNodes {
			walk(c)
			walk(c.ChildNode)
		}
	}
	walk(root)
	return out
}

func AssigneeIDUint(a Assignee) uint64 {
	switch v := a.ID.(type) {
	case float64:
		return uint64(v)
	case int:
		return uint64(v)
	case int64:
		return uint64(v)
	case uint64:
		return v
	case json.Number:
		i, _ := v.Int64()
		return uint64(i)
	case string:
		var n uint64
		fmt.Sscanf(v, "%d", &n)
		return n
	default:
		return 0
	}
}

func OrgMemberIDUint(m OrgMember) uint64 {
	return AssigneeIDUint(Assignee{ID: m.ID})
}
