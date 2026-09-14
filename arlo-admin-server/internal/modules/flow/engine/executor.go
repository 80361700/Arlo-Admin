package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"arlo-admin/internal/modules/flow/model"
)

// RuntimeRepo 引擎持久化
type RuntimeRepo interface {
	CreateInstance(ctx context.Context, inst *model.FlowInstance) error
	UpdateInstance(ctx context.Context, inst *model.FlowInstance) error
	GetInstance(ctx context.Context, id uint64) (*model.FlowInstance, error)

	CreateTask(ctx context.Context, task *model.FlowTask, actors []model.FlowTaskActor) error
	GetTask(ctx context.Context, id uint64) (*model.FlowTask, error)
	ListActiveTasksByInstance(ctx context.Context, instanceID uint64) ([]model.FlowTask, error)
	ListActiveTasksByGate(ctx context.Context, instanceID uint64, gateToken string) ([]model.FlowTask, error)
	UpdateTask(ctx context.Context, task *model.FlowTask) error
	DeleteTask(ctx context.Context, taskID uint64) error
	ListTaskActors(ctx context.Context, taskID uint64) ([]model.FlowTaskActor, error)
	UpdateTaskActor(ctx context.Context, actor *model.FlowTaskActor) error
	CreateTaskActor(ctx context.Context, actor *model.FlowTaskActor) error
	DeleteTaskActor(ctx context.Context, actorID uint64) error

	ArchiveTask(ctx context.Context, task *model.FlowTask, actors []model.FlowTaskActor) error

	// 同一人重复审批跳过：已在本实例中同意过的用户
	ListInstanceAgreedUserIDs(ctx context.Context, instanceID uint64) ([]uint64, error)

	// 子流程
	CountActiveChildren(ctx context.Context, parentID uint64, parentNodeKey string) (int64, error)
	ListActiveChildren(ctx context.Context, parentID uint64) ([]model.FlowInstance, error)

	// 历史任务（驳回「上一节点」按实际路径）
	ListHisTasks(ctx context.Context, instanceID uint64) ([]model.FlowHisTask, error)
}

// LaunchSelection 发起时自选审批人/抄送人：key=nodeKey
type LaunchSelection struct {
	Assignees map[string][]UserRef
	CCUsers   map[string][]UserRef
}

// Engine 流程引擎
type Engine struct {
	Repo      RuntimeRepo
	Org       OrgStore
	Notify    Notifier      // 可选：抄送提醒等站内信
	Delegates DelegateStore // 可选：审批委托
}

// Notifier 运行时通知（由服务层注入站内信实现）
type Notifier interface {
	Notify(ctx context.Context, title, content string, receiverIDs []uint64) error
}

// DelegateStore 审批委托查询
type DelegateStore interface {
	PrincipalsOf(ctx context.Context, delegateUserID uint64) []uint64
}

type walkState struct {
	inst      *model.FlowInstance
	model     *ModelContent
	formData  map[string]interface{}
	eval      *EvalContext
	selection *LaunchSelection
	fromKey   string
	gateToken string
	// gateStack 外层网关 token，避免并行内再套排他时覆盖外层汇合键
	gateStack []string
	operator  UserRef
	opinion   string
}

func (st *walkState) pushGate(token string) {
	if st.gateToken != "" {
		st.gateStack = append(st.gateStack, st.gateToken)
	}
	st.gateToken = token
}

func (st *walkState) popGate() {
	if len(st.gateStack) == 0 {
		st.gateToken = ""
		return
	}
	st.gateToken = st.gateStack[len(st.gateStack)-1]
	st.gateStack = st.gateStack[:len(st.gateStack)-1]
}

// clearGatesOutsideTarget 若目标不在当前/外层网关子树内，逐层弹出 gateToken
func clearGatesOutsideTarget(st *walkState, targetKey string) {
	if st == nil || st.model == nil || targetKey == "" {
		return
	}
	for st.gateToken != "" {
		gw := FindNode(st.model.NodeConfig, st.gateToken)
		if gw != nil && FindNode(gw, targetKey) != nil {
			return
		}
		st.popGate()
	}
}

// findRejectableAncestor 向上找可驳回落地的发起/审批节点（跳过网关、条件条等）
func findRejectableAncestor(root *Node, childKey string) *Node {
	curKey := childKey
	seen := map[string]bool{}
	for i := 0; i < 64; i++ {
		if curKey == "" || seen[curKey] {
			return nil
		}
		seen[curKey] = true
		p := FindParentOf(root, curKey)
		if p == nil {
			return nil
		}
		switch p.Type {
		case NodeStart, NodeApproval:
			return p
		default:
			curKey = p.NodeKey
		}
	}
	return nil
}

func (st *walkState) clone() *walkState {
	cp := *st
	if st.gateStack != nil {
		cp.gateStack = append([]string(nil), st.gateStack...)
	}
	return &cp
}

// mergeGateStackPayload 把外层网关栈写入任务 payload，便于异步办理后恢复
func mergeGateStackPayload(existing string, stack []string) string {
	m := map[string]interface{}{}
	if strings.TrimSpace(existing) != "" {
		_ = json.Unmarshal([]byte(existing), &m)
	}
	if len(stack) == 0 {
		delete(m, "gateStack")
	} else {
		m["gateStack"] = append([]string(nil), stack...)
	}
	if len(m) == 0 {
		return existing
	}
	b, err := json.Marshal(m)
	if err != nil {
		return existing
	}
	return string(b)
}

func restoreGateStack(st *walkState, payload string) {
	if st == nil || strings.TrimSpace(payload) == "" {
		return
	}
	var meta struct {
		GateStack []string `json:"gateStack"`
	}
	if err := json.Unmarshal([]byte(payload), &meta); err != nil {
		return
	}
	if len(meta.GateStack) > 0 {
		st.gateStack = append([]string(nil), meta.GateStack...)
	}
}

func applyTaskGate(st *walkState, task *model.FlowTask) {
	if st == nil || task == nil {
		return
	}
	task.GateToken = st.gateToken
	task.Payload = mergeGateStackPayload(task.Payload, st.gateStack)
}

// Start 发起实例并推进到首个待办
func (e *Engine) Start(ctx context.Context, inst *model.FlowInstance, selection *LaunchSelection) error {
	mc, err := ParseModel(inst.ModelContent)
	if err != nil {
		return err
	}
	formData := map[string]interface{}{}
	if inst.FormData != "" {
		_ = json.Unmarshal([]byte(inst.FormData), &formData)
	}
	roles, _ := e.Org.UserRoleIDs(inst.CreateID)
	eval := &EvalContext{
		FormData:       formData,
		InitiatorID:    inst.CreateID,
		InitiatorDept:  inst.CreateDeptID,
		InitiatorRoles: roles,
		BelongChecker:  BuildBelongChecker(e.Org, inst.CreateID, inst.CreateDeptID, roles),
	}
	if err := e.Repo.CreateInstance(ctx, inst); err != nil {
		return err
	}
	st := &walkState{
		inst:      inst,
		model:     mc,
		formData:  formData,
		eval:      eval,
		selection: selection,
		fromKey:   "",
	}
	start := mc.NodeConfig
	// 发起节点本身不产生审批待办，进入 child
	next := start.ChildNode
	if next == nil {
		return e.finishInstance(ctx, inst, model.InstComplete, start.NodeKey, start.NodeName)
	}
	return e.enterNode(ctx, st, next)
}

// ContinueFromStart 已有实例从发起节点继续推进（暂存激活）
func (e *Engine) ContinueFromStart(ctx context.Context, inst *model.FlowInstance, selection *LaunchSelection) error {
	if inst == nil || inst.ID == 0 {
		return fmt.Errorf("实例不存在")
	}
	mc, err := ParseModel(inst.ModelContent)
	if err != nil {
		return err
	}
	formData := map[string]interface{}{}
	if inst.FormData != "" {
		_ = json.Unmarshal([]byte(inst.FormData), &formData)
	}
	roles, _ := e.Org.UserRoleIDs(inst.CreateID)
	eval := &EvalContext{
		FormData:       formData,
		InitiatorID:    inst.CreateID,
		InitiatorDept:  inst.CreateDeptID,
		InitiatorRoles: roles,
		BelongChecker:  BuildBelongChecker(e.Org, inst.CreateID, inst.CreateDeptID, roles),
	}
	inst.InstanceState = model.InstActive
	inst.FinishTime = nil
	inst.UpdatedAt = time.Now()
	if err := e.Repo.UpdateInstance(ctx, inst); err != nil {
		return err
	}
	revertDraft := func() {
		inst.InstanceState = model.InstDraft
		inst.UpdatedAt = time.Now()
		_ = e.Repo.UpdateInstance(ctx, inst)
	}
	st := &walkState{
		inst:      inst,
		model:     mc,
		formData:  formData,
		eval:      eval,
		selection: selection,
		fromKey:   "",
	}
	start := mc.NodeConfig
	if start == nil {
		revertDraft()
		return fmt.Errorf("流程模型缺少发起节点")
	}
	next := start.ChildNode
	if next == nil {
		return e.finishInstance(ctx, inst, model.InstComplete, start.NodeKey, start.NodeName)
	}
	if err := e.enterNode(ctx, st, next); err != nil {
		revertDraft()
		return err
	}
	return nil
}

// Consent 同意
func (e *Engine) Consent(ctx context.Context, taskID, userID uint64, opinion string, formData map[string]interface{}) error {
	return e.completeActor(ctx, taskID, userID, opinion, formData, true, 0, "")
}

// Reject 拒绝/驳回
// rejectStrategyOverride: 0 用节点配置；否则强制策略
// rejectNodeKey: 策略=指定节点时的目标
func (e *Engine) Reject(ctx context.Context, taskID, userID uint64, opinion string, formData map[string]interface{}, strategy int, rejectNodeKey string) error {
	return e.completeActor(ctx, taskID, userID, opinion, formData, false, strategy, rejectNodeKey)
}

func (e *Engine) completeActor(ctx context.Context, taskID, userID uint64, opinion string, formData map[string]interface{}, agree bool, strategy int, rejectNodeKey string) error {
	task, err := e.Repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	if task.TaskState != model.TaskActive {
		return fmt.Errorf("任务已结束")
	}
	inst, err := e.Repo.GetInstance(ctx, task.InstanceID)
	if err != nil {
		return err
	}
	if inst.InstanceState != model.InstActive {
		return fmt.Errorf("流程已结束")
	}
	actors, err := e.Repo.ListTaskActors(ctx, taskID)
	if err != nil {
		return err
	}

	principals := []uint64(nil)
	if e.Delegates != nil && processSettingBool(inst, "allowDelegate", true) {
		principals = e.Delegates.PrincipalsOf(ctx, userID)
	}
	me := findActableActor(actors, userID, principals)
	if me == nil {
		return fmt.Errorf("无权处理该任务")
	}
	now := time.Now()
	me.Opinion = opinion
	if me.ActorID != userID {
		me.AgentID = userID // 代批人
	}
	me.FinishTime = &now
	if agree {
		me.ActorState = model.ActorAgree
	} else {
		me.ActorState = model.ActorRefuse
	}
	if err := e.Repo.UpdateTaskActor(ctx, me); err != nil {
		return err
	}

	if formData != nil {
		b, _ := json.Marshal(formData)
		inst.FormData = string(b)
		_ = e.Repo.UpdateInstance(ctx, inst)
	}

	mc, err := ParseModel(inst.ModelContent)
	if err != nil {
		return err
	}
	node := FindNode(mc.NodeConfig, task.NodeKey)
	if node == nil {
		return fmt.Errorf("节点不存在: %s", task.NodeKey)
	}

	name, _, _ := e.Org.GetUser(userID)
	op := UserRef{ID: userID, Name: name}

	if !agree {
		return e.handleReject(ctx, inst, mc, task, actors, node, op, opinion, strategy, rejectNodeKey)
	}

	// 多人审批
	mode := int8(node.ExamineMode)
	if mode == 0 {
		mode = 1
	}
	taskDone := false
	switch mode {
	case 2: // 会签：全部同意
		all := true
		for _, a := range actors {
			if a.ID == me.ID {
				continue
			}
			if a.ActorState == model.ActorPending {
				all = false
				break
			}
			if a.ActorState == model.ActorRefuse {
				all = false
				break
			}
		}
		taskDone = all
	case 3: // 或签：一人同意即可
		taskDone = true
		for i := range actors {
			if actors[i].ID != me.ID && actors[i].ActorState == model.ActorPending {
				actors[i].ActorState = model.ActorSkip
				actors[i].FinishTime = &now
				_ = e.Repo.UpdateTaskActor(ctx, &actors[i])
			}
		}
	default: // 依次
		nextWeight := me.Weight + 1
		var next *model.FlowTaskActor
		for i := range actors {
			if actors[i].Weight == nextWeight && actors[i].ActorState == model.ActorPending {
				next = &actors[i]
				break
			}
		}
		if next != nil {
			// 依次：下一人仍保持 pending，当前已同意；任务未结束
			taskDone = false
		} else {
			taskDone = true
		}
		// 依次审批时，未轮到的保持 pending；已同意的留下
		// 若还有下一人，不激活特殊状态（都是 pending，前端按 weight 过滤）
		// 约定：依次时只有 weight==当前最小 pending 的可操作 —— 在 API 层校验
	}

	if !taskDone {
		// 依次审批：通知下一位当前可办人
		if mode == 1 {
			if refreshed, e2 := e.Repo.ListTaskActors(ctx, taskID); e2 == nil {
				e.notifyPendingArrival(ctx, inst, task, refreshed)
			}
		}
		return nil
	}

	// 归档任务并前进
	actors, _ = e.Repo.ListTaskActors(ctx, taskID)
	task.TaskState = model.TaskComplete
	task.UpdatedAt = now
	if err := e.Repo.ArchiveTask(ctx, task, actors); err != nil {
		return err
	}
	if err := e.Repo.DeleteTask(ctx, task.ID); err != nil {
		return err
	}

	fd := map[string]interface{}{}
	_ = json.Unmarshal([]byte(inst.FormData), &fd)
	roles, _ := e.Org.UserRoleIDs(inst.CreateID)
	st := &walkState{
		inst:     inst,
		model:    mc,
		formData: fd,
		eval: &EvalContext{
			FormData:       fd,
			InitiatorID:    inst.CreateID,
			InitiatorDept:  inst.CreateDeptID,
			InitiatorRoles: roles,
			BelongChecker:  BuildBelongChecker(e.Org, inst.CreateID, inst.CreateDeptID, roles),
		},
		fromKey:   task.NodeKey,
		gateToken: task.GateToken,
		operator:  op,
		opinion:   opinion,
	}
	restoreGateStack(st, task.Payload)

	// 驳回再审：rejectStart=2 时回到原驳回节点（不在网关等待时）
	if task.GateToken == "" {
		if fromKey, rs, ok := e.takeRejectReturn(inst); ok && rs == 2 {
			_ = e.Repo.UpdateInstance(ctx, inst)
			target := FindNode(mc.NodeConfig, fromKey)
			if target != nil {
				return e.enterNode(ctx, st, target)
			}
		} else if ok {
			_ = e.Repo.UpdateInstance(ctx, inst)
		}
	}

	var child *Node
	if node != nil {
		child = node.ChildNode
	}
	return e.continueAfterLeaf(ctx, st, task.NodeKey, task.NodeName, child)
}

func (e *Engine) handleReject(ctx context.Context, inst *model.FlowInstance, mc *ModelContent, task *model.FlowTask, actors []model.FlowTaskActor, node *Node, op UserRef, opinion string, strategyOverride int, rejectNodeKey string) error {
	return e.handleRejectWithTerminateState(ctx, inst, mc, task, actors, node, op, opinion, strategyOverride, rejectNodeKey, model.InstReject)
}

// handleRejectWithTerminateState 驳回；terminateState 用于策略=终止时的实例终态（普通驳回 InstReject，超时 InstTimeout）
func (e *Engine) handleRejectWithTerminateState(ctx context.Context, inst *model.FlowInstance, mc *ModelContent, task *model.FlowTask, actors []model.FlowTaskActor, node *Node, op UserRef, opinion string, strategyOverride int, rejectNodeKey string, terminateState int8) error {
	now := time.Now()
	strategy := 3 // 与设计器默认「指定节点」一致
	if node != nil {
		strategy = node.RejectStrategy
	}
	if strategyOverride > 0 {
		strategy = strategyOverride
	}
	if strategy == 0 {
		// 与设计器默认一致：指定节点（3）；未选目标时由前端/批量校验拦截
		strategy = 3
	}
	if terminateState == 0 {
		terminateState = model.InstReject
	}

	// 标记其余人跳过
	for i := range actors {
		if actors[i].ActorState == model.ActorPending {
			actors[i].ActorState = model.ActorSkip
			actors[i].FinishTime = &now
			_ = e.Repo.UpdateTaskActor(ctx, &actors[i])
		}
	}
	actors, _ = e.Repo.ListTaskActors(ctx, task.ID)
	task.TaskState = model.TaskReject
	_ = e.Repo.ArchiveTask(ctx, task, actors)
	_ = e.Repo.DeleteTask(ctx, task.ID)

	// 清理同实例其他活动任务：
	// - 终止策略：清全部
	// - 有网关令牌：仅清同令牌兄弟，避免误杀其它并行网关/延时任务
	// - 无令牌（线性）：清全部
	actives, _ := e.Repo.ListActiveTasksByInstance(ctx, inst.ID)
	for _, t := range actives {
		if strategy != 4 && task.GateToken != "" && t.GateToken != task.GateToken {
			continue
		}
		as, _ := e.Repo.ListTaskActors(ctx, t.ID)
		t.TaskState = model.TaskTerminate
		_ = e.Repo.ArchiveTask(ctx, &t, as)
		_ = e.Repo.DeleteTask(ctx, t.ID)
	}

	fd := map[string]interface{}{}
	_ = json.Unmarshal([]byte(inst.FormData), &fd)
	roles, _ := e.Org.UserRoleIDs(inst.CreateID)
	st := &walkState{
		inst:     inst,
		model:    mc,
		formData: fd,
		eval: &EvalContext{
			FormData:       fd,
			InitiatorID:    inst.CreateID,
			InitiatorDept:  inst.CreateDeptID,
			InitiatorRoles: roles,
			BelongChecker:  BuildBelongChecker(e.Org, inst.CreateID, inst.CreateDeptID, roles),
		},
		fromKey:  task.NodeKey,
		operator: op,
		opinion:  opinion,
	}

	switch strategy {
	case 1: // 驳回到发起人：发起人改单后重新提交
		start := mc.NodeConfig
		if start == nil {
			return e.finishInstance(ctx, inst, terminateState, task.NodeKey, task.NodeName)
		}
		e.setRejectReturn(inst, task.NodeKey, nodeRejectStart(node))
		_ = e.Repo.UpdateInstance(ctx, inst)
		inst.InstanceState = model.InstActive
		return e.createInitiatorReviseTask(ctx, st, start)
	case 2: // 上一节点：优先按本实例历史路径，回退到模型 DFS
		prev := e.findPreviousApprovalRuntime(ctx, inst.ID, mc.NodeConfig, task.NodeKey)
		if prev == nil {
			return e.finishInstance(ctx, inst, terminateState, task.NodeKey, task.NodeName)
		}
		e.setRejectReturn(inst, task.NodeKey, nodeRejectStart(node))
		_ = e.Repo.UpdateInstance(ctx, inst)
		return e.enterNode(ctx, st, prev)
	case 3: // 指定节点
		key := rejectNodeKey
		if key == "" && node != nil && node.ExtendConfig != nil {
			if v, ok := node.ExtendConfig["rejectNodeKey"].(string); ok {
				key = v
			}
		}
		target := FindNode(mc.NodeConfig, key)
		if target == nil {
			return fmt.Errorf("驳回目标节点不存在")
		}
		e.setRejectReturn(inst, task.NodeKey, nodeRejectStart(node))
		_ = e.Repo.UpdateInstance(ctx, inst)
		if target.Type == NodeStart {
			inst.InstanceState = model.InstActive
			return e.createInitiatorReviseTask(ctx, st, target)
		}
		return e.enterNode(ctx, st, target)
	case 5: // 模型父节点：落到可办理的发起/审批祖先，跳过网关/条件条
		target := findRejectableAncestor(mc.NodeConfig, task.NodeKey)
		if target == nil {
			// 无合适祖先时退回上一审批，再不行才终止
			prev := e.findPreviousApprovalRuntime(ctx, inst.ID, mc.NodeConfig, task.NodeKey)
			if prev == nil {
				return e.finishInstance(ctx, inst, terminateState, task.NodeKey, task.NodeName)
			}
			target = prev
		}
		e.setRejectReturn(inst, task.NodeKey, nodeRejectStart(node))
		_ = e.Repo.UpdateInstance(ctx, inst)
		if target.Type == NodeStart {
			inst.InstanceState = model.InstActive
			return e.createInitiatorReviseTask(ctx, st, target)
		}
		return e.enterNode(ctx, st, target)
	default: // 4 终止
		return e.finishInstance(ctx, inst, terminateState, task.NodeKey, task.NodeName)
	}
}

func nodeRejectStart(node *Node) int {
	if node == nil {
		return 1
	}
	return node.RejectStart
}

// createInitiatorReviseTask 驳回/回退到发起人后，给发起人一张改单重提待办
func (e *Engine) createInitiatorReviseTask(ctx context.Context, st *walkState, start *Node) error {
	if start == nil {
		return fmt.Errorf("发起人节点不存在")
	}
	now := time.Now()
	name := start.NodeName
	if name == "" {
		name = "发起人"
	}
	if !strings.Contains(name, "修改") {
		name = name + "（修改）"
	}
	task := &model.FlowTask{
		InstanceID:  st.inst.ID,
		NodeKey:     start.NodeKey,
		NodeName:    name,
		NodeType:    NodeStart,
		TaskType:    model.TaskMajor,
		TaskState:   model.TaskActive,
		ExamineMode: 1,
		FromNodeKey: st.fromKey,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	actors := []model.FlowTaskActor{{
		InstanceID: st.inst.ID,
		ActorID:    st.inst.CreateID,
		ActorName:  st.inst.CreateBy,
		ActorState: model.ActorPending,
		Weight:     1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}}
	st.inst.CurrentNodeKey = start.NodeKey
	st.inst.CurrentNodeName = name
	st.inst.InstanceState = model.InstActive
	st.inst.FinishTime = nil
	st.inst.UpdatedAt = now
	if err := e.Repo.UpdateInstance(ctx, st.inst); err != nil {
		return err
	}
	return e.Repo.CreateTask(ctx, task, actors)
}

// Resubmit 发起人改单后重新提交（仅 NodeType=发起人 的待办）
func (e *Engine) Resubmit(ctx context.Context, taskID, userID uint64, formData map[string]interface{}) error {
	task, err := e.Repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	if task.TaskState != model.TaskActive {
		return fmt.Errorf("任务已结束")
	}
	if task.NodeType != NodeStart {
		return fmt.Errorf("当前不是发起人改单任务")
	}
	inst, err := e.Repo.GetInstance(ctx, task.InstanceID)
	if err != nil {
		return err
	}
	if inst.CreateID != userID {
		return fmt.Errorf("仅发起人可重新提交")
	}
	actors, err := e.Repo.ListTaskActors(ctx, taskID)
	if err != nil {
		return err
	}
	if !CanUserAct(task, actors, userID) {
		return fmt.Errorf("无权处理该任务")
	}
	mc, err := ParseModel(inst.ModelContent)
	if err != nil {
		return err
	}
	now := time.Now()
	if formData != nil {
		b, _ := json.Marshal(formData)
		inst.FormData = string(b)
	}
	for i := range actors {
		if actors[i].ActorID == userID && actors[i].ActorState == model.ActorPending {
			actors[i].ActorState = model.ActorAgree
			actors[i].Opinion = "重新提交"
			actors[i].FinishTime = &now
			_ = e.Repo.UpdateTaskActor(ctx, &actors[i])
			break
		}
	}
	actors, _ = e.Repo.ListTaskActors(ctx, taskID)
	task.TaskState = model.TaskComplete
	task.UpdatedAt = now
	if err := e.Repo.ArchiveTask(ctx, task, actors); err != nil {
		return err
	}
	if err := e.Repo.DeleteTask(ctx, task.ID); err != nil {
		return err
	}

	fd := map[string]interface{}{}
	_ = json.Unmarshal([]byte(inst.FormData), &fd)
	roles, _ := e.Org.UserRoleIDs(inst.CreateID)
	st := &walkState{
		inst:     inst,
		model:    mc,
		formData: fd,
		eval: &EvalContext{
			FormData:       fd,
			InitiatorID:    inst.CreateID,
			InitiatorDept:  inst.CreateDeptID,
			InitiatorRoles: roles,
			BelongChecker:  BuildBelongChecker(e.Org, inst.CreateID, inst.CreateDeptID, roles),
		},
		fromKey:  task.NodeKey,
		operator: UserRef{ID: userID, Name: inst.CreateBy},
		opinion:  "重新提交",
	}

	fromKey, rs, ok := e.takeRejectReturn(inst)
	_ = e.Repo.UpdateInstance(ctx, inst)

	// rejectStart=2：回到原驳回节点；否则从发起人后继续往下
	if ok && rs == 2 && fromKey != "" {
		target := FindNode(mc.NodeConfig, fromKey)
		if target != nil {
			return e.enterNode(ctx, st, target)
		}
	}
	start := mc.NodeConfig
	if start != nil && start.ChildNode != nil {
		return e.enterNode(ctx, st, start.ChildNode)
	}
	return e.finishInstance(ctx, inst, model.InstComplete, task.NodeKey, task.NodeName)
}

func (e *Engine) setRejectReturn(inst *model.FlowInstance, fromKey string, rejectStart int) {
	v := map[string]interface{}{}
	if inst.Variable != "" {
		_ = json.Unmarshal([]byte(inst.Variable), &v)
	}
	v["rejectFromKey"] = fromKey
	if rejectStart == 0 {
		rejectStart = 1
	}
	v["rejectStart"] = rejectStart
	b, _ := json.Marshal(v)
	inst.Variable = string(b)
}

func (e *Engine) takeRejectReturn(inst *model.FlowInstance) (fromKey string, rejectStart int, ok bool) {
	if inst.Variable == "" {
		return "", 0, false
	}
	v := map[string]interface{}{}
	if err := json.Unmarshal([]byte(inst.Variable), &v); err != nil {
		return "", 0, false
	}
	fromKey, _ = v["rejectFromKey"].(string)
	switch n := v["rejectStart"].(type) {
	case float64:
		rejectStart = int(n)
	case int:
		rejectStart = n
	}
	if fromKey == "" {
		return "", 0, false
	}
	delete(v, "rejectFromKey")
	delete(v, "rejectStart")
	b, _ := json.Marshal(v)
	inst.Variable = string(b)
	return fromKey, rejectStart, true
}

func findPreviousApproval(root *Node, currentKey string) *Node {
	// 优先按模型父链找上一可办理节点，避免 DFS 全树顺序在并行/分支下指错
	if p := findRejectableAncestor(root, currentKey); p != nil {
		return p
	}
	nodes := CollectApprovalNodes(root)
	var prev *Node
	for _, n := range nodes {
		if n.NodeKey == currentKey {
			return prev
		}
		prev = n
	}
	return nil
}

// findPreviousApprovalRuntime 取本实例最近一次已完成审批节点（不含当前），再映射到模型节点
func (e *Engine) findPreviousApprovalRuntime(ctx context.Context, instanceID uint64, root *Node, currentKey string) *Node {
	his, err := e.Repo.ListHisTasks(ctx, instanceID)
	if err == nil && len(his) > 0 {
		for i := len(his) - 1; i >= 0; i-- {
			h := his[i]
			if h.NodeKey == "" || h.NodeKey == currentKey {
				continue
			}
			if h.TaskType != model.TaskMajor {
				continue
			}
			if h.TaskState != model.TaskComplete {
				continue
			}
			// 发起人改单重提不算「上一审批」
			if h.NodeType == NodeStart {
				continue
			}
			if n := FindNode(root, h.NodeKey); n != nil && (n.Type == NodeApproval || n.Type == NodeStart) {
				return n
			}
		}
	}
	return findPreviousApproval(root, currentKey)
}

func (e *Engine) finishInstance(ctx context.Context, inst *model.FlowInstance, state int8, nodeKey, nodeName string) error {
	now := time.Now()
	inst.InstanceState = state
	inst.CurrentNodeKey = nodeKey
	inst.CurrentNodeName = nodeName
	inst.FinishTime = &now
	inst.UpdatedAt = now
	return e.Repo.UpdateInstance(ctx, inst)
}

func (e *Engine) enterNode(ctx context.Context, st *walkState, node *Node) error {
	if node == nil {
		return e.finishInstance(ctx, st.inst, model.InstComplete, st.fromKey, "")
	}
	prevKey, prevName := st.inst.CurrentNodeKey, st.inst.CurrentNodeName
	st.inst.CurrentNodeKey = node.NodeKey
	st.inst.CurrentNodeName = node.NodeName
	_ = e.Repo.UpdateInstance(ctx, st.inst)

	err := e.dispatchEnterNode(ctx, st, node)
	if err != nil {
		// 进入失败时回滚当前节点，避免留下「流程图执行中、却无待办」的空壳
		st.inst.CurrentNodeKey = prevKey
		st.inst.CurrentNodeName = prevName
		_ = e.Repo.UpdateInstance(ctx, st.inst)
		return err
	}
	return nil
}

func (e *Engine) dispatchEnterNode(ctx context.Context, st *walkState, node *Node) error {
	switch node.Type {
	case NodeEnd:
		return e.finishInstance(ctx, st.inst, model.InstComplete, node.NodeKey, node.NodeName)
	case NodeAutoPass:
		return e.autoPass(ctx, st, node)
	case NodeAutoReject:
		return e.autoReject(ctx, st, node)
	case NodeApproval:
		return e.createApprovalTask(ctx, st, node)
	case NodeCC:
		return e.handleCC(ctx, st, node)
	case NodeExclusive:
		br := PickExclusiveBranch(node.ConditionNodes, st.eval)
		if br == nil {
			return fmt.Errorf("条件分支无可用路径")
		}
		// 压栈外层网关，避免覆盖并行/包容的汇合 token
		st.pushGate(node.NodeKey)
		if br.ChildNode != nil {
			return e.enterNode(ctx, st, br.ChildNode)
		}
		st.popGate()
		if node.ChildNode != nil {
			return e.enterNode(ctx, st, node.ChildNode)
		}
		return e.continueAfterLeaf(ctx, st, node.NodeKey, node.NodeName, nil)
	case NodeInclusive:
		branches := PickInclusiveBranches(node.InclusiveNodes, st.eval)
		if len(branches) == 0 {
			return fmt.Errorf("包容分支无可用路径")
		}
		st.pushGate(node.NodeKey)
		gateToken := st.gateToken
		for _, br := range branches {
			target := br.ChildNode
			if target == nil {
				continue
			}
			st2 := st.clone()
			st2.gateToken = gateToken
			if err := e.enterNode(ctx, st2, target); err != nil {
				return err
			}
		}
		left, _ := e.Repo.ListActiveTasksByGate(ctx, st.inst.ID, gateToken)
		if len(left) == 0 {
			st.popGate()
			if node.ChildNode != nil {
				return e.enterNode(ctx, st, node.ChildNode)
			}
			return e.continueAfterLeaf(ctx, st, node.NodeKey, node.NodeName, nil)
		}
		return nil
	case NodeParallel:
		st.pushGate(node.NodeKey)
		gateToken := st.gateToken
		for _, br := range node.ParallelNodes {
			target := br.ChildNode
			if target == nil {
				continue
			}
			st2 := st.clone()
			st2.gateToken = gateToken
			if err := e.enterNode(ctx, st2, target); err != nil {
				return err
			}
		}
		left, _ := e.Repo.ListActiveTasksByGate(ctx, st.inst.ID, gateToken)
		if len(left) == 0 {
			st.popGate()
			if node.ChildNode != nil {
				return e.enterNode(ctx, st, node.ChildNode)
			}
			return e.continueAfterLeaf(ctx, st, node.NodeKey, node.NodeName, nil)
		}
		return nil
	case NodeRoute:
		targetKey := PickRouteTarget(node.RouteNodes, st.eval)
		if targetKey == "" {
			if node.ChildNode != nil {
				return e.enterNode(ctx, st, node.ChildNode)
			}
			return e.finishInstance(ctx, st.inst, model.InstComplete, node.NodeKey, node.NodeName)
		}
		target := FindNode(st.model.NodeConfig, targetKey)
		if target == nil {
			return fmt.Errorf("路由目标不存在: %s", targetKey)
		}
		// 跨出当前并行/包容网关时弹出 gate，避免令牌残留导致汇合错乱
		clearGatesOutsideTarget(st, target.NodeKey)
		return e.enterNode(ctx, st, target)
	case NodeDelay:
		return e.handleDelay(ctx, st, node)
	case NodeTrigger:
		return e.handleTrigger(ctx, st, node)
	case NodeSubProcess:
		return e.handleSubProcess(ctx, st, node)
	case NodeStart, NodeCondition:
		if node.ChildNode != nil {
			return e.enterNode(ctx, st, node.ChildNode)
		}
		if st.fromKey != "" {
			// 条件条无 child，回到网关后续 —— 由网关逻辑处理
			return e.finishInstance(ctx, st.inst, model.InstComplete, node.NodeKey, node.NodeName)
		}
		return nil
	default:
		return fmt.Errorf("不支持的节点类型: %d (%s)", node.Type, node.NodeName)
	}
}

func (e *Engine) autoPass(ctx context.Context, st *walkState, node *Node) error {
	return e.autoPassWithOpinion(ctx, st, node, "自动通过")
}

func (e *Engine) autoPassWithOpinion(ctx context.Context, st *walkState, node *Node, opinion string) error {
	if opinion == "" {
		opinion = "自动通过"
	}
	now := time.Now()
	task := &model.FlowTask{
		InstanceID:  st.inst.ID,
		NodeKey:     node.NodeKey,
		NodeName:    node.NodeName,
		NodeType:    node.Type,
		TaskType:    model.TaskMajor,
		TaskState:   model.TaskComplete,
		ExamineMode: 1,
		FromNodeKey: st.fromKey,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	applyTaskGate(st, task)
	actors := []model.FlowTaskActor{{
		InstanceID: st.inst.ID,
		ActorID:    0,
		ActorName:  "系统",
		ActorState: model.ActorAgree,
		Opinion:    opinion,
		FinishTime: &now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}}
	if err := e.Repo.ArchiveTask(ctx, task, actors); err != nil {
		return err
	}
	// 并行/包容分支内自动通过且无后续子节点时，须等待兄弟分支，不可直接完结整单
	return e.continueAfterLeaf(ctx, st, node.NodeKey, node.NodeName, node.ChildNode)
}

func (e *Engine) autoReject(ctx context.Context, st *walkState, node *Node) error {
	return e.autoRejectWithOpinion(ctx, st, node, "自动拒绝")
}

// autoRejectWithOpinion 归档「自动拒绝」历史后整单拒绝，并清理其它活动任务
func (e *Engine) autoRejectWithOpinion(ctx context.Context, st *walkState, node *Node, opinion string) error {
	if opinion == "" {
		opinion = "自动拒绝"
	}
	now := time.Now()
	task := &model.FlowTask{
		InstanceID:  st.inst.ID,
		NodeKey:     node.NodeKey,
		NodeName:    node.NodeName,
		NodeType:    node.Type,
		TaskType:    model.TaskMajor,
		TaskState:   model.TaskReject,
		ExamineMode: 1,
		FromNodeKey: st.fromKey,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	applyTaskGate(st, task)
	actors := []model.FlowTaskActor{{
		InstanceID: st.inst.ID,
		ActorID:    0,
		ActorName:  "系统",
		ActorState: model.ActorRefuse,
		Opinion:    opinion,
		FinishTime: &now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}}
	if err := e.Repo.ArchiveTask(ctx, task, actors); err != nil {
		return err
	}
	// 整单拒绝：清掉并行/延时等残留活动任务
	actives, _ := e.Repo.ListActiveTasksByInstance(ctx, st.inst.ID)
	for _, t := range actives {
		as, _ := e.Repo.ListTaskActors(ctx, t.ID)
		t.TaskState = model.TaskTerminate
		_ = e.Repo.ArchiveTask(ctx, &t, as)
		_ = e.Repo.DeleteTask(ctx, t.ID)
	}
	return e.finishInstance(ctx, st.inst, model.InstReject, node.NodeKey, node.NodeName)
}

func (e *Engine) createApprovalTask(ctx context.Context, st *walkState, node *Node) error {
	var selected []UserRef
	if st.selection != nil && st.selection.Assignees != nil {
		selected = st.selection.Assignees[node.NodeKey]
	}

	now := time.Now()
	mode := int8(node.ExamineMode)
	if mode == 0 {
		mode = 1
	}
	task := &model.FlowTask{
		InstanceID:  st.inst.ID,
		NodeKey:     node.NodeKey,
		NodeName:    node.NodeName,
		NodeType:    node.Type,
		TaskType:    model.TaskMajor,
		TaskState:   model.TaskActive,
		ExamineMode: mode,
		FromNodeKey: st.fromKey,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	applyTaskGate(st, task)

	// 角色认领池：挂角色槽，不展开用户
	if UseClaimPool(node) {
		roles, err := ResolveClaimRoles(node, selected)
		if err != nil {
			return err
		}
		actors := make([]model.FlowTaskActor, 0, len(roles))
		for i, r := range roles {
			actors = append(actors, model.FlowTaskActor{
				InstanceID: st.inst.ID,
				ActorID:    r.ID, // 角色 id
				ActorName:  r.Name,
				ActorType:  model.ActorTypeRole,
				ActorState: model.ActorPending,
				Weight:     i + 1,
				CreatedAt:  now,
				UpdatedAt:  now,
			})
		}
		applyApprovalTimers(task, node, now)
		if err := e.Repo.CreateTask(ctx, task, actors); err != nil {
			return err
		}
		e.notifyPendingArrival(ctx, st.inst, task, actors)
		return nil
	}

	users, err := ResolveAssignees(node, st.inst.CreateID, selected, e.Org)
	if err != nil {
		return err
	}
	beforeSkip := len(users)
	skippedByRepeat := false
	// 同一人重复节点自动跳过：仅当前审批段内已同意过的人不进入本节点（驳回/重提后需重新审）
	if processSettingBool(st.inst, "repeatOperateSkip", true) && len(users) > 0 {
		if agreed, e2 := e.Repo.ListInstanceAgreedUserIDs(ctx, st.inst.ID); e2 == nil && len(agreed) > 0 {
			filtered := filterOutAgreedUsers(users, agreed)
			if len(filtered) < len(users) {
				skippedByRepeat = true
			}
			users = filtered
		}
	}
	// approveSelf / 重复跳过后无人：自动通过
	if len(users) == 0 {
		if beforeSkip > 0 && skippedByRepeat {
			return e.autoPassWithOpinion(ctx, st, node, "重复审批人自动跳过")
		}
		return e.autoPass(ctx, st, node)
	}
	actors := make([]model.FlowTaskActor, 0, len(users))
	for i, u := range users {
		actors = append(actors, model.FlowTaskActor{
			InstanceID: st.inst.ID,
			ActorID:    u.ID,
			ActorName:  u.Name,
			ActorType:  model.ActorTypeUser,
			ActorState: model.ActorPending,
			Weight:     i + 1,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	applyApprovalTimers(task, node, now)
	if err := e.Repo.CreateTask(ctx, task, actors); err != nil {
		return err
	}
	e.notifyPendingArrival(ctx, st.inst, task, actors)
	return nil
}

func (e *Engine) handleCC(ctx context.Context, st *walkState, node *Node) error {
	var selected []UserRef
	if st.selection != nil && st.selection.CCUsers != nil {
		selected = st.selection.CCUsers[node.NodeKey]
	}
	users := ResolveCC(node, selected, e.Org)
	now := time.Now()
	task := &model.FlowTask{
		InstanceID:  st.inst.ID,
		NodeKey:     node.NodeKey,
		NodeName:    node.NodeName,
		NodeType:    node.Type,
		TaskType:    model.TaskCC,
		TaskState:   model.TaskComplete,
		ExamineMode: 1,
		FromNodeKey: st.fromKey,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	applyTaskGate(st, task)
	actors := make([]model.FlowTaskActor, 0, len(users))
	for _, u := range users {
		actors = append(actors, model.FlowTaskActor{
			InstanceID: st.inst.ID,
			ActorID:    u.ID,
			ActorName:  u.Name,
			ActorState: model.ActorAgree,
			Opinion:    "抄送",
			FinishTime: &now,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	if err := e.Repo.ArchiveTask(ctx, task, actors); err != nil {
		return err
	}
	// 抄送节点开启 remind：到达即通知抄送人
	if node.Remind && e.Notify != nil && len(users) > 0 {
		ids := make([]uint64, 0, len(users))
		for _, u := range users {
			if u.ID > 0 {
				ids = append(ids, u.ID)
			}
		}
		if len(ids) > 0 {
			title := fmt.Sprintf("抄送：%s", st.inst.ProcessName)
			content := fmt.Sprintf("流程「%s」已抄送给您（节点「%s」），请知悉。", st.inst.ProcessName, node.NodeName)
			_ = e.Notify.Notify(ctx, title, content, ids)
		}
	}
	return e.continueAfterLeaf(ctx, st, node.NodeKey, node.NodeName, node.ChildNode)
}

func (e *Engine) handleDelay(ctx context.Context, st *walkState, node *Node) error {
	// 完整：创建延时任务，到期由 TickDelays 推进；固定时长从 extendConfig.time 解析
	dur := parseDelayDuration(node)
	exp := time.Now().Add(dur)
	now := time.Now()
	payload, _ := json.Marshal(map[string]interface{}{"delay": dur.String()})
	task := &model.FlowTask{
		InstanceID:  st.inst.ID,
		NodeKey:     node.NodeKey,
		NodeName:    node.NodeName,
		NodeType:    node.Type,
		TaskType:    model.TaskDelay,
		TaskState:   model.TaskActive,
		FromNodeKey: st.fromKey,
		ExpireTime:  &exp,
		Payload:     string(payload),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	applyTaskGate(st, task)
	return e.Repo.CreateTask(ctx, task, nil)
}

func parseDelayDuration(node *Node) time.Duration {
	if node == nil {
		return time.Minute
	}
	// delayType=2：自动计算到指定时刻（extendConfig.time = HH:mm:ss），取下一次到达该时刻的时长
	if node.DelayType == 2 {
		return durationUntilClock(node)
	}
	// delayType=1 / 默认：固定时长，extendConfig.time 如 "1:m" / "1:h" / "1:d"
	if node.ExtendConfig == nil {
		return time.Minute
	}
	raw, _ := node.ExtendConfig["time"].(string)
	if raw == "" {
		return time.Minute
	}
	var n int
	var unit string
	fmt.Sscanf(raw, "%d:%s", &n, &unit)
	if n <= 0 {
		n = 1
	}
	switch unit {
	case "h":
		return time.Duration(n) * time.Hour
	case "d":
		return time.Duration(n) * 24 * time.Hour
	default:
		return time.Duration(n) * time.Minute
	}
}

// durationUntilClock 解析 HH:mm:ss，返回距下一次该时刻的时长（已过则取次日）
func durationUntilClock(node *Node) time.Duration {
	if node.ExtendConfig == nil {
		return time.Minute
	}
	raw, _ := node.ExtendConfig["time"].(string)
	if raw == "" {
		return time.Minute
	}
	var h, m, s int
	n, _ := fmt.Sscanf(raw, "%d:%d:%d", &h, &m, &s)
	if n < 2 {
		return time.Minute
	}
	if h < 0 || h > 23 || m < 0 || m > 59 || s < 0 || s > 59 {
		return time.Minute
	}
	now := time.Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), h, m, s, 0, now.Location())
	if !target.After(now) {
		target = target.Add(24 * time.Hour)
	}
	d := target.Sub(now)
	if d < time.Second {
		return time.Second
	}
	return d
}

// continueAfterLeaf 节点无后续子节点时：若在并行/包容/排他网关内则汇聚，否则结束实例
func (e *Engine) continueAfterLeaf(ctx context.Context, st *walkState, nodeKey, nodeName string, child *Node) error {
	st.fromKey = nodeKey
	if child != nil {
		return e.enterNode(ctx, st, child)
	}
	if st.gateToken != "" {
		left, _ := e.Repo.ListActiveTasksByGate(ctx, st.inst.ID, st.gateToken)
		if len(left) == 0 {
			gate := FindNode(st.model.NodeConfig, st.gateToken)
			st.popGate()
			if gate != nil && gate.ChildNode != nil {
				return e.enterNode(ctx, st, gate.ChildNode)
			}
			// 本层网关无 child：继续尝试外层汇合或结束
			return e.continueAfterLeaf(ctx, st, nodeKey, nodeName, nil)
		}
		return nil
	}
	return e.finishInstance(ctx, st.inst, model.InstComplete, nodeKey, nodeName)
}

// CompleteDelayTask 延时到期推进
func (e *Engine) CompleteDelayTask(ctx context.Context, taskID uint64) error {
	task, err := e.Repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	if task.TaskType != model.TaskDelay || task.TaskState != model.TaskActive {
		return nil
	}
	inst, err := e.Repo.GetInstance(ctx, task.InstanceID)
	if err != nil {
		return err
	}
	mc, err := ParseModel(inst.ModelContent)
	if err != nil {
		return err
	}
	node := FindNode(mc.NodeConfig, task.NodeKey)
	actors, _ := e.Repo.ListTaskActors(ctx, task.ID)
	task.TaskState = model.TaskComplete
	_ = e.Repo.ArchiveTask(ctx, task, actors)
	_ = e.Repo.DeleteTask(ctx, task.ID)

	fd := map[string]interface{}{}
	_ = json.Unmarshal([]byte(inst.FormData), &fd)
	roles, _ := e.Org.UserRoleIDs(inst.CreateID)
	st := &walkState{
		inst:     inst,
		model:    mc,
		formData: fd,
		eval: &EvalContext{
			FormData:       fd,
			InitiatorID:    inst.CreateID,
			InitiatorDept:  inst.CreateDeptID,
			InitiatorRoles: roles,
			BelongChecker:  BuildBelongChecker(e.Org, inst.CreateID, inst.CreateDeptID, roles),
		},
		fromKey:   task.NodeKey,
		gateToken: task.GateToken,
	}
	restoreGateStack(st, task.Payload)

	// 延迟触发器到期：先执行 HTTP，再按触发节点子树推进
	if triggerPayloadKind(task.Payload) == "trigger" {
		tn := nodeFromTriggerPayload(task.Payload)
		if tn == nil {
			tn = node
		}
		if tn != nil && tn.NodeKey == "" && node != nil {
			tn.NodeKey = node.NodeKey
			tn.NodeName = node.NodeName
		}
		opinion, terr := invokeHTTPTrigger(ctx, inst, tn, fd)
		if terr != nil {
			return e.finishInstance(ctx, inst, model.InstReject, task.NodeKey, task.NodeName)
		}
		_ = opinion
		var child *Node
		if node != nil {
			child = node.ChildNode
		}
		return e.continueAfterLeaf(ctx, st, task.NodeKey, task.NodeName, child)
	}

	var child *Node
	if node != nil {
		child = node.ChildNode
	}
	return e.continueAfterLeaf(ctx, st, task.NodeKey, task.NodeName, child)
}

func (e *Engine) handleTrigger(ctx context.Context, st *walkState, node *Node) error {
	// triggerType=2：延迟后执行 HTTP（复用延时任务到期推进）
	if node.TriggerType == 2 {
		dur := parseDelayDuration(node)
		exp := time.Now().Add(dur)
		now := time.Now()
		payload, _ := json.Marshal(map[string]interface{}{
			"kind":         "trigger",
			"nodeKey":      node.NodeKey,
			"nodeName":     node.NodeName,
			"extendConfig": node.ExtendConfig,
			"delay":        dur.String(),
		})
		task := &model.FlowTask{
			InstanceID:  st.inst.ID,
			NodeKey:     node.NodeKey,
			NodeName:    node.NodeName,
			NodeType:    node.Type,
			TaskType:    model.TaskDelay,
			TaskState:   model.TaskActive,
			FromNodeKey: st.fromKey,
			ExpireTime:  &exp,
			Payload:     string(payload),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		applyTaskGate(st, task)
		return e.Repo.CreateTask(ctx, task, nil)
	}
	return e.runTriggerNow(ctx, st, node)
}

func (e *Engine) runTriggerNow(ctx context.Context, st *walkState, node *Node) error {
	now := time.Now()
	opinion, err := invokeHTTPTrigger(ctx, st.inst, node, st.formData)
	payload, _ := json.Marshal(node.ExtendConfig)
	task := &model.FlowTask{
		InstanceID:  st.inst.ID,
		NodeKey:     node.NodeKey,
		NodeName:    node.NodeName,
		NodeType:    node.Type,
		TaskType:    model.TaskTrig,
		TaskState:   model.TaskComplete,
		FromNodeKey: st.fromKey,
		Payload:     string(payload),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	applyTaskGate(st, task)
	actorState := model.ActorAgree
	if err != nil {
		task.TaskState = model.TaskReject
		actorState = model.ActorRefuse
		if opinion == "" {
			opinion = err.Error()
		} else {
			opinion = err.Error()
		}
	} else if opinion == "" {
		opinion = "触发器执行"
	}
	actors := []model.FlowTaskActor{{
		InstanceID: st.inst.ID,
		ActorName:  "系统",
		ActorState: actorState,
		Opinion:    opinion,
		FinishTime: &now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}}
	if aerr := e.Repo.ArchiveTask(ctx, task, actors); aerr != nil {
		return aerr
	}
	if err != nil {
		// 同步触发失败：终止本实例（与子流程失败语义一致，避免静默跳过）
		return e.finishInstance(ctx, st.inst, model.InstReject, node.NodeKey, node.NodeName)
	}
	return e.continueAfterLeaf(ctx, st, node.NodeKey, node.NodeName, node.ChildNode)
}

func (e *Engine) handleSubProcess(ctx context.Context, st *walkState, node *Node) error {
	// 子流程：创建占位活动任务，业务层根据 callProcess/subProcessValue 启动子实例后回写（仅同步等待）
	ref := node.CallProcess
	if ref == "" {
		ref = node.SubProcessValue
	}
	if ref == "" {
		return fmt.Errorf("子流程节点[%s]未配置子流程", node.NodeName)
	}
	now := time.Now()
	payload, _ := json.Marshal(map[string]string{"callProcess": ref})
	task := &model.FlowTask{
		InstanceID:  st.inst.ID,
		NodeKey:     node.NodeKey,
		NodeName:    node.NodeName,
		NodeType:    node.Type,
		TaskType:    model.TaskSub,
		TaskState:   model.TaskActive,
		FromNodeKey: st.fromKey,
		Payload:     string(payload),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	applyTaskGate(st, task)
	return e.Repo.CreateTask(ctx, task, nil)
}

// ResumeAfterSubProcess 子流程结束后继续父节点
func (e *Engine) ResumeAfterSubProcess(ctx context.Context, parentInstanceID uint64, parentNodeKey string) error {
	inst, err := e.Repo.GetInstance(ctx, parentInstanceID)
	if err != nil {
		return err
	}
	tasks, err := e.Repo.ListActiveTasksByInstance(ctx, parentInstanceID)
	if err != nil {
		return err
	}
	var task *model.FlowTask
	for i := range tasks {
		if tasks[i].NodeKey == parentNodeKey && tasks[i].TaskType == model.TaskSub {
			task = &tasks[i]
			break
		}
	}
	if task == nil {
		return nil
	}
	mc, err := ParseModel(inst.ModelContent)
	if err != nil {
		return err
	}
	node := FindNode(mc.NodeConfig, parentNodeKey)
	actors, _ := e.Repo.ListTaskActors(ctx, task.ID)
	now := time.Now()
	task.TaskState = model.TaskComplete
	if len(actors) == 0 {
		actors = []model.FlowTaskActor{{
			InstanceID: task.InstanceID,
			ActorName:  "系统",
			ActorState: model.ActorAgree,
			Opinion:    "子流程完成",
			FinishTime: &now,
			CreatedAt:  task.CreatedAt,
			UpdatedAt:  now,
		}}
	}
	_ = e.Repo.ArchiveTask(ctx, task, actors)
	_ = e.Repo.DeleteTask(ctx, task.ID)

	fd := map[string]interface{}{}
	_ = json.Unmarshal([]byte(inst.FormData), &fd)
	roles, _ := e.Org.UserRoleIDs(inst.CreateID)
	st := &walkState{
		inst:     inst,
		model:    mc,
		formData: fd,
		eval: &EvalContext{
			FormData:       fd,
			InitiatorID:    inst.CreateID,
			InitiatorDept:  inst.CreateDeptID,
			InitiatorRoles: roles,
			BelongChecker:  BuildBelongChecker(e.Org, inst.CreateID, inst.CreateDeptID, roles),
		},
		fromKey:   task.NodeKey,
		gateToken: task.GateToken,
	}
	restoreGateStack(st, task.Payload)
	var child *Node
	if node != nil {
		child = node.ChildNode
	}
	return e.continueAfterLeaf(ctx, st, task.NodeKey, task.NodeName, child)
}

// FailAfterSubProcess 子流程失败：归档父占位任务并终止父实例（同步子流程语义）
func (e *Engine) FailAfterSubProcess(ctx context.Context, parentInstanceID uint64, parentNodeKey string, childState int8, opinion string) error {
	inst, err := e.Repo.GetInstance(ctx, parentInstanceID)
	if err != nil {
		return err
	}
	if inst.InstanceState != model.InstActive {
		return nil
	}
	tasks, err := e.Repo.ListActiveTasksByInstance(ctx, parentInstanceID)
	if err != nil {
		return err
	}
	var task *model.FlowTask
	for i := range tasks {
		if tasks[i].NodeKey == parentNodeKey && tasks[i].TaskType == model.TaskSub {
			task = &tasks[i]
			break
		}
	}
	if task == nil {
		// 无占位任务时仍终止父实例，避免卡死
		return e.finishInstance(ctx, inst, model.InstReject, parentNodeKey, inst.CurrentNodeName)
	}
	now := time.Now()
	if opinion == "" {
		opinion = "子流程未通过"
	}
	actors := []model.FlowTaskActor{{
		InstanceID: task.InstanceID,
		ActorName:  "系统",
		ActorState: model.ActorRefuse,
		Opinion:    opinion,
		FinishTime: &now,
		CreatedAt:  task.CreatedAt,
		UpdatedAt:  now,
	}}
	task.TaskState = model.TaskReject
	if childState == model.InstRevoke {
		task.TaskState = model.TaskRevoke
	} else if childState == model.InstTimeout {
		task.TaskState = model.TaskTerminate
	}
	_ = e.Repo.ArchiveTask(ctx, task, actors)
	_ = e.Repo.DeleteTask(ctx, task.ID)

	// 清理父实例上其它活动任务（并行兄弟等），整单按拒绝结束
	left, _ := e.Repo.ListActiveTasksByInstance(ctx, parentInstanceID)
	for i := range left {
		t := left[i]
		as, _ := e.Repo.ListTaskActors(ctx, t.ID)
		for j := range as {
			if as[j].ActorState == model.ActorPending {
				as[j].ActorState = model.ActorSkip
				as[j].Opinion = opinion
				as[j].FinishTime = &now
			}
		}
		t.TaskState = model.TaskTerminate
		_ = e.Repo.ArchiveTask(ctx, &t, as)
		_ = e.Repo.DeleteTask(ctx, t.ID)
	}

	parentState := model.InstReject
	if childState == model.InstTimeout {
		parentState = model.InstTimeout
	} else if childState == model.InstRevoke {
		parentState = model.InstReject
	}
	return e.finishInstance(ctx, inst, parentState, task.NodeKey, task.NodeName)
}

// CanUserAct 依次审批时是否轮到该用户（认领池角色槽 ActorTypeRole 不可直接办）
// principalIDs: 委托给当前用户的委托人 id，允许代批
func CanUserAct(task *model.FlowTask, actors []model.FlowTaskActor, userID uint64, principalIDs ...uint64) bool {
	allow := map[uint64]struct{}{userID: {}}
	for _, id := range principalIDs {
		if id > 0 {
			allow[id] = struct{}{}
		}
	}
	isUserPending := func(a model.FlowTaskActor) bool {
		if a.ActorState != model.ActorPending || a.ActorType == model.ActorTypeRole {
			return false
		}
		_, ok := allow[a.ActorID]
		return ok
	}
	if task.ExamineMode != 1 {
		for _, a := range actors {
			if isUserPending(a) {
				return true
			}
		}
		return false
	}
	// 依次：最小 weight 的 pending（含未认领角色槽，挡住后续）
	minW := 0
	for _, a := range actors {
		if a.ActorState == model.ActorPending {
			if minW == 0 || a.Weight < minW {
				minW = a.Weight
			}
		}
	}
	for _, a := range actors {
		if isUserPending(a) && a.Weight == minW {
			return true
		}
	}
	return false
}

// collectActablePendingUserIDs 当前真正可办的待办接收人（依次只取最小 weight；角色槽展开）
func (e *Engine) collectActablePendingUserIDs(task *model.FlowTask, actors []model.FlowTaskActor) []uint64 {
	if task == nil {
		return nil
	}
	seen := map[uint64]bool{}
	var ids []uint64
	add := func(id uint64) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true
		ids = append(ids, id)
	}
	addActor := func(a model.FlowTaskActor) {
		if a.ActorState != model.ActorPending {
			return
		}
		if a.ActorType == model.ActorTypeRole {
			if e.Org != nil {
				uids, _ := e.Org.ListUserIDsByRole(a.ActorID)
				for _, uid := range uids {
					add(uid)
				}
			}
			return
		}
		add(a.ActorID)
	}
	if task.ExamineMode != 1 {
		for _, a := range actors {
			addActor(a)
		}
		return ids
	}
	minW := 0
	for _, a := range actors {
		if a.ActorState == model.ActorPending {
			if minW == 0 || a.Weight < minW {
				minW = a.Weight
			}
		}
	}
	for _, a := range actors {
		if a.ActorState == model.ActorPending && a.Weight == minW {
			addActor(a)
		}
	}
	return ids
}

// notifyPendingArrival 待办到达站内信（仅当前可办人）
func (e *Engine) notifyPendingArrival(ctx context.Context, inst *model.FlowInstance, task *model.FlowTask, actors []model.FlowTaskActor) {
	if e.Notify == nil || inst == nil || task == nil {
		return
	}
	ids := e.collectActablePendingUserIDs(task, actors)
	if len(ids) == 0 {
		return
	}
	title := fmt.Sprintf("待办到达：%s", inst.ProcessName)
	content := fmt.Sprintf("流程「%s」节点「%s」待您处理（发起人：%s）。", inst.ProcessName, task.NodeName, inst.CreateBy)
	_ = e.Notify.Notify(ctx, title, content, ids)
}

// resolveOperatorActor 解析当前操作人可办理的 pending 槽（含委托代批）
func (e *Engine) resolveOperatorActor(ctx context.Context, inst *model.FlowInstance, task *model.FlowTask, actors []model.FlowTaskActor, userID uint64) (*model.FlowTaskActor, error) {
	principals := []uint64(nil)
	if e.Delegates != nil && processSettingBool(inst, "allowDelegate", true) {
		principals = e.Delegates.PrincipalsOf(ctx, userID)
	}
	if !CanUserAct(task, actors, userID, principals...) {
		return nil, fmt.Errorf("无权处理该任务")
	}
	me := findActableActor(actors, userID, principals)
	if me == nil {
		return nil, fmt.Errorf("无权处理该任务")
	}
	return me, nil
}

func findActableActor(actors []model.FlowTaskActor, userID uint64, principalIDs []uint64) *model.FlowTaskActor {
	for i := range actors {
		a := &actors[i]
		if a.ActorState != model.ActorPending || a.ActorType == model.ActorTypeRole {
			continue
		}
		if a.ActorID == userID {
			return a
		}
	}
	if len(principalIDs) == 0 {
		return nil
	}
	set := map[uint64]struct{}{}
	for _, id := range principalIDs {
		set[id] = struct{}{}
	}
	for i := range actors {
		a := &actors[i]
		if a.ActorState != model.ActorPending || a.ActorType == model.ActorTypeRole {
			continue
		}
		if _, ok := set[a.ActorID]; ok {
			return a
		}
	}
	return nil
}

func filterOutAgreedUsers(users []UserRef, agreedIDs []uint64) []UserRef {
	if len(users) == 0 || len(agreedIDs) == 0 {
		return users
	}
	set := map[uint64]struct{}{}
	for _, id := range agreedIDs {
		if id > 0 {
			set[id] = struct{}{}
		}
	}
	out := make([]UserRef, 0, len(users))
	for _, u := range users {
		if _, ok := set[u.ID]; ok {
			continue
		}
		out = append(out, u)
	}
	return out
}
