package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"arlo-admin/internal/modules/flow/model"
)

// ---------- 实例 Variable 辅助 ----------

func parseVariable(inst *model.FlowInstance) map[string]interface{} {
	v := map[string]interface{}{}
	if inst != nil && inst.Variable != "" {
		_ = json.Unmarshal([]byte(inst.Variable), &v)
	}
	return v
}

func saveVariable(inst *model.FlowInstance, v map[string]interface{}) {
	b, _ := json.Marshal(v)
	inst.Variable = string(b)
}

// SnapshotProcessSetting 发起时把流程设置写入 Variable，供撤销等运行时规则使用
func SnapshotProcessSetting(inst *model.FlowInstance, processSettingJSON string) {
	v := parseVariable(inst)
	if processSettingJSON != "" {
		var setting interface{}
		if err := json.Unmarshal([]byte(processSettingJSON), &setting); err == nil {
			v["processSetting"] = setting
		}
	}
	saveVariable(inst, v)
}

// SnapshotLaunchSelection 暂存发起人自选审批人/抄送，激活时回放
func SnapshotLaunchSelection(inst *model.FlowInstance, sel *LaunchSelection) {
	if inst == nil {
		return
	}
	v := parseVariable(inst)
	if sel == nil {
		delete(v, "launchSelection")
		saveVariable(inst, v)
		return
	}
	payload := map[string]interface{}{
		"assignees": map[string]interface{}{},
		"ccUsers":   map[string]interface{}{},
	}
	asg, _ := payload["assignees"].(map[string]interface{})
	for k, list := range sel.Assignees {
		arr := make([]map[string]interface{}, 0, len(list))
		for _, u := range list {
			arr = append(arr, map[string]interface{}{"id": u.ID, "name": u.Name})
		}
		asg[k] = arr
	}
	cc, _ := payload["ccUsers"].(map[string]interface{})
	for k, list := range sel.CCUsers {
		arr := make([]map[string]interface{}, 0, len(list))
		for _, u := range list {
			arr = append(arr, map[string]interface{}{"id": u.ID, "name": u.Name})
		}
		cc[k] = arr
	}
	v["launchSelection"] = payload
	saveVariable(inst, v)
}

// LoadLaunchSelection 从 Variable 恢复自选人
func LoadLaunchSelection(inst *model.FlowInstance) *LaunchSelection {
	sel := &LaunchSelection{
		Assignees: map[string][]UserRef{},
		CCUsers:   map[string][]UserRef{},
	}
	if inst == nil {
		return sel
	}
	v := parseVariable(inst)
	raw, ok := v["launchSelection"].(map[string]interface{})
	if !ok {
		return sel
	}
	parseBag := func(key string, dest map[string][]UserRef) {
		m, ok := raw[key].(map[string]interface{})
		if !ok {
			return
		}
		for nk, list := range m {
			arr, ok := list.([]interface{})
			if !ok {
				continue
			}
			for _, it := range arr {
				row, ok := it.(map[string]interface{})
				if !ok {
					continue
				}
				var id uint64
				switch t := row["id"].(type) {
				case float64:
					id = uint64(t)
				case int64:
					id = uint64(t)
				case int:
					id = uint64(t)
				case json.Number:
					n, _ := t.Int64()
					id = uint64(n)
				}
				name, _ := row["name"].(string)
				if id == 0 {
					continue
				}
				dest[nk] = append(dest[nk], UserRef{ID: id, Name: name})
			}
		}
	}
	parseBag("assignees", sel.Assignees)
	parseBag("ccUsers", sel.CCUsers)
	return sel
}

func ProcessSettingBool(inst *model.FlowInstance, key string, defaultVal bool) bool {
	return processSettingBool(inst, key, defaultVal)
}

func processSettingBool(inst *model.FlowInstance, key string, defaultVal bool) bool {
	v := parseVariable(inst)
	raw, ok := v["processSetting"]
	if !ok {
		return defaultVal
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return defaultVal
	}
	x, ok := m[key]
	if !ok {
		return defaultVal
	}
	switch t := x.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case string:
		return t == "true" || t == "1"
	default:
		return defaultVal
	}
}

// ---------- 撤销 ----------

// Revoke 发起人撤销审批中的实例（级联撤销活动中的子实例）
func (e *Engine) Revoke(ctx context.Context, instanceID, userID uint64, opinion string) error {
	return e.revokeInstance(ctx, instanceID, userID, opinion, false)
}

// AdminRevoke 流程管理员撤销（跳过「仅发起人」校验）
func (e *Engine) AdminRevoke(ctx context.Context, instanceID, userID uint64, opinion string) error {
	return e.revokeInstance(ctx, instanceID, userID, opinion, true)
}

// Terminate 强制终止审批中的实例（级联终止活动子实例）
func (e *Engine) Terminate(ctx context.Context, instanceID, userID uint64, opinion string) error {
	inst, err := e.Repo.GetInstance(ctx, instanceID)
	if err != nil {
		return err
	}
	if inst.InstanceState != model.InstActive {
		return fmt.Errorf("仅审批中的流程可终止")
	}
	if opinion == "" {
		opinion = "管理员终止"
	}
	children, _ := e.Repo.ListActiveChildren(ctx, inst.ID)
	for i := range children {
		_ = e.Terminate(ctx, children[i].ID, userID, "父流程终止")
	}
	actives, _ := e.Repo.ListActiveTasksByInstance(ctx, inst.ID)
	now := time.Now()
	for _, t := range actives {
		actors, _ := e.Repo.ListTaskActors(ctx, t.ID)
		for i := range actors {
			if actors[i].ActorState == model.ActorPending {
				actors[i].ActorState = model.ActorSkip
				actors[i].Opinion = opinion
				actors[i].FinishTime = &now
			}
		}
		t.TaskState = model.TaskTerminate
		_ = e.Repo.ArchiveTask(ctx, &t, actors)
		_ = e.Repo.DeleteTask(ctx, t.ID)
	}
	return e.finishInstance(ctx, inst, model.InstTerminate, inst.CurrentNodeKey, inst.CurrentNodeName)
}

// AdminTransfer 管理员转办：不校验节点 allowTransfer / 当前办理人
// fromActorID：待办槽位 flow_task_actor.id；多人待办时必填，仅 1 人时可传 0
func (e *Engine) AdminTransfer(ctx context.Context, taskID, adminID, toUserID uint64, toUserName, opinion string, fromActorID uint64) error {
	if toUserID == 0 {
		return fmt.Errorf("请选择转办人")
	}
	task, err := e.Repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	if task.TaskState != model.TaskActive || task.TaskType != model.TaskMajor {
		return fmt.Errorf("当前无可转办任务")
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
	for _, a := range actors {
		if a.ActorID == toUserID && a.ActorState == model.ActorPending {
			return fmt.Errorf("对方已在待办列表中")
		}
	}
	var pending []*model.FlowTaskActor
	for i := range actors {
		if actors[i].ActorState != model.ActorPending {
			continue
		}
		if actors[i].ActorType == model.ActorTypeUser || actors[i].ActorType == model.ActorTypeAppend {
			pending = append(pending, &actors[i])
		}
	}
	if len(pending) == 0 {
		for i := range actors {
			if actors[i].ActorState == model.ActorPending {
				pending = append(pending, &actors[i])
			}
		}
	}
	if len(pending) == 0 {
		return fmt.Errorf("当前节点没有待办人可转办")
	}
	var slot *model.FlowTaskActor
	if fromActorID > 0 {
		for _, a := range pending {
			if a.ID == fromActorID {
				slot = a
				break
			}
		}
		if slot == nil {
			return fmt.Errorf("指定的待办人不存在或已处理")
		}
	} else if len(pending) == 1 {
		slot = pending[0]
	} else {
		return fmt.Errorf("当前节点有多人待办，请指定要转办的办理人")
	}
	if toUserID == slot.ActorID {
		return fmt.Errorf("不能转办给当前办理人")
	}
	now := time.Now()
	if opinion == "" {
		opinion = "管理员转办"
	}
	slot.ActorState = model.ActorTransfer
	slot.Opinion = opinion
	slot.FinishTime = &now
	slot.UpdatedAt = now
	if err := e.Repo.UpdateTaskActor(ctx, slot); err != nil {
		return err
	}
	name := toUserName
	if name == "" {
		name, _, _ = e.Org.GetUser(toUserID)
	}
	if err := e.Repo.CreateTaskActor(ctx, &model.FlowTaskActor{
		TaskID:     task.ID,
		InstanceID: task.InstanceID,
		ActorID:    toUserID,
		ActorName:  name,
		ActorType:  model.ActorTypeUser,
		ActorState: model.ActorPending,
		Weight:     slot.Weight,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		return err
	}
	if refreshed, e2 := e.Repo.ListTaskActors(ctx, task.ID); e2 == nil {
		e.notifyPendingArrival(ctx, inst, task, refreshed)
	}
	return nil
}

// revokeInstance force=true 时跳过发起人校验（父流程级联撤销子实例）
func (e *Engine) revokeInstance(ctx context.Context, instanceID, userID uint64, opinion string, force bool) error {
	inst, err := e.Repo.GetInstance(ctx, instanceID)
	if err != nil {
		return err
	}
	if inst.InstanceState != model.InstActive {
		return fmt.Errorf("仅审批中的流程可撤销")
	}
	if !force {
		if inst.CreateID != userID {
			return fmt.Errorf("仅发起人可撤销")
		}
		if !processSettingBool(inst, "allowRevocation", true) {
			return fmt.Errorf("该流程不允许撤销")
		}
	}
	if opinion == "" {
		opinion = "撤销"
	}

	// 先级联撤销活动子实例，再归档本实例任务
	children, _ := e.Repo.ListActiveChildren(ctx, inst.ID)
	for i := range children {
		_ = e.revokeInstance(ctx, children[i].ID, userID, "父流程撤销", true)
	}

	actives, _ := e.Repo.ListActiveTasksByInstance(ctx, inst.ID)
	now := time.Now()
	for _, t := range actives {
		actors, _ := e.Repo.ListTaskActors(ctx, t.ID)
		for i := range actors {
			if actors[i].ActorState == model.ActorPending {
				actors[i].ActorState = model.ActorSkip
				actors[i].Opinion = opinion
				actors[i].FinishTime = &now
			}
		}
		t.TaskState = model.TaskRevoke
		_ = e.Repo.ArchiveTask(ctx, &t, actors)
		_ = e.Repo.DeleteTask(ctx, t.ID)
	}
	return e.finishInstance(ctx, inst, model.InstRevoke, inst.CurrentNodeKey, inst.CurrentNodeName)
}

// ---------- 转交 ----------

// Transfer 当前处理人将待办转交给他人
func (e *Engine) Transfer(ctx context.Context, taskID, userID uint64, toUserID uint64, toUserName, opinion string) error {
	if toUserID == 0 {
		return fmt.Errorf("请选择转交人")
	}
	if toUserID == userID {
		return fmt.Errorf("不能转交给自己")
	}
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
	mc, err := ParseModel(inst.ModelContent)
	if err != nil {
		return err
	}
	node := FindNode(mc.NodeConfig, task.NodeKey)
	if node == nil {
		return fmt.Errorf("节点不存在")
	}
	if !node.AllowTransfer {
		return fmt.Errorf("当前节点不允许转交")
	}
	actors, err := e.Repo.ListTaskActors(ctx, taskID)
	if err != nil {
		return err
	}
	me, err := e.resolveOperatorActor(ctx, inst, task, actors, userID)
	if err != nil {
		return err
	}
	for _, a := range actors {
		if a.ActorID == toUserID && a.ActorState == model.ActorPending {
			return fmt.Errorf("对方已在待办列表中")
		}
	}
	now := time.Now()
	me.ActorState = model.ActorTransfer
	me.Opinion = opinion
	me.FinishTime = &now
	// AgentID 仅表示「代批人」；本人转交自己的槽位时不要写入，避免流转记录误显示代批
	if me.ActorID != userID {
		me.AgentID = userID
	}
	if err := e.Repo.UpdateTaskActor(ctx, me); err != nil {
		return err
	}
	name := toUserName
	if name == "" {
		name, _, _ = e.Org.GetUser(toUserID)
	}
	if err := e.Repo.CreateTaskActor(ctx, &model.FlowTaskActor{
		TaskID:     task.ID,
		InstanceID: task.InstanceID,
		ActorID:    toUserID,
		ActorName:  name,
		ActorState: model.ActorPending,
		Weight:     me.Weight,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		return err
	}
	if refreshed, e2 := e.Repo.ListTaskActors(ctx, task.ID); e2 == nil {
		e.notifyPendingArrival(ctx, inst, task, refreshed)
	}
	return nil
}

// ---------- 回退 ----------

// Rollback 回退到指定已走过节点（需节点允许回退）
func (e *Engine) Rollback(ctx context.Context, taskID, userID uint64, opinion, targetNodeKey string) error {
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
	mc, err := ParseModel(inst.ModelContent)
	if err != nil {
		return err
	}
	node := FindNode(mc.NodeConfig, task.NodeKey)
	if node == nil {
		return fmt.Errorf("节点不存在")
	}
	if !node.AllowRollback {
		return fmt.Errorf("当前节点不允许回退")
	}
	actors, err := e.Repo.ListTaskActors(ctx, taskID)
	if err != nil {
		return err
	}
	me, err := e.resolveOperatorActor(ctx, inst, task, actors, userID)
	if err != nil {
		return err
	}
	name, _, _ := e.Org.GetUser(userID)
	op := UserRef{ID: userID, Name: name}

	strategy := 2 // 上一节点
	if targetNodeKey != "" {
		strategy = 3
	}
	// 先把当前人标记为拒绝语义，再走驳回路由
	now := time.Now()
	me.ActorState = model.ActorRefuse
	me.Opinion = opinion
	me.FinishTime = &now
	// 仅真实代批时记录 AgentID（本人回退不写，避免「自己代批自己」）
	if me.ActorID != userID {
		me.AgentID = userID
	}
	_ = e.Repo.UpdateTaskActor(ctx, me)
	actors, _ = e.Repo.ListTaskActors(ctx, taskID)
	return e.handleReject(ctx, inst, mc, task, actors, node, op, opinion, strategy, targetNodeKey)
}

// ---------- 加签 / 减签 ----------

// AppendActor 加签：向当前任务追加审批人
// position: 1=前加签(更小 weight，依次审批时先处理) 2=后加签
func (e *Engine) AppendActor(ctx context.Context, taskID, userID uint64, toUserID uint64, toUserName string, position int) error {
	if toUserID == 0 {
		return fmt.Errorf("请选择加签人")
	}
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
	mc, err := ParseModel(inst.ModelContent)
	if err != nil {
		return err
	}
	node := FindNode(mc.NodeConfig, task.NodeKey)
	if node == nil {
		return fmt.Errorf("节点不存在")
	}
	if !node.AllowAppendNode {
		return fmt.Errorf("当前节点不允许加签")
	}
	if node.SetType == 6 {
		return fmt.Errorf("连续多级主管节点不支持加签")
	}
	if node.ExamineMode == 1 {
		// 依次审批：加签人按插入权重加入顺序，与会签不同
	}
	actors, err := e.Repo.ListTaskActors(ctx, taskID)
	if err != nil {
		return err
	}
	me, err := e.resolveOperatorActor(ctx, inst, task, actors, userID)
	if err != nil {
		return err
	}
	for _, a := range actors {
		if a.ActorID == toUserID && (a.ActorState == model.ActorPending || a.ActorState == model.ActorAgree) {
			return fmt.Errorf("该用户已在审批人列表中")
		}
	}
	myWeight := me.Weight
	found := true
	_ = found

	// 插入式重排：前加签插到当前人之前，后加签插到当前人之后（避免简单 ±1 撞权重、展示顺序错乱）
	now := time.Now()
	insertAt := myWeight
	if position != 1 {
		insertAt = myWeight + 1
	}
	for i := range actors {
		a := &actors[i]
		if a.Weight >= insertAt {
			a.Weight++
			a.UpdatedAt = now
			if err := e.Repo.UpdateTaskActor(ctx, a); err != nil {
				return err
			}
		}
	}

	name := toUserName
	if name == "" {
		name, _, _ = e.Org.GetUser(toUserID)
	}
	if err := e.Repo.CreateTaskActor(ctx, &model.FlowTaskActor{
		TaskID:     task.ID,
		InstanceID: task.InstanceID,
		ActorID:    toUserID,
		ActorName:  name,
		ActorType:  model.ActorTypeAppend, // 加签
		ActorState: model.ActorPending,
		Weight:     insertAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		return err
	}
	// 前加签会立刻轮到对方；后加签仅在轮到时由依次推进再通知
	if refreshed, e2 := e.Repo.ListTaskActors(ctx, task.ID); e2 == nil {
		e.notifyPendingArrival(ctx, inst, task, refreshed)
	}
	return nil
}

// RemoveActor 减签：仅可移除尚未处理的「加签人」(ActorType=1)
// 规则（与设计器「允许加签/减签」对齐，避免掏空原审批人导致会签/或签异常）：
// 1. 节点须开启 allowAppendNode，且非连续多级主管
// 2. 操作人须为当前可办人
// 3. 目标必须是 pending 的加签人，不能减自己、不能减原审批人
// 4. 减完后至少保留 1 名 pending
func (e *Engine) RemoveActor(ctx context.Context, taskID, userID, removeRowID uint64) error {
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
	mc, err := ParseModel(inst.ModelContent)
	if err != nil {
		return err
	}
	node := FindNode(mc.NodeConfig, task.NodeKey)
	if node == nil || !node.AllowAppendNode {
		return fmt.Errorf("当前节点不允许减签")
	}
	if node.SetType == 6 {
		return fmt.Errorf("连续多级主管节点不支持减签")
	}
	actors, err := e.Repo.ListTaskActors(ctx, taskID)
	if err != nil {
		return err
	}
	me, err := e.resolveOperatorActor(ctx, inst, task, actors, userID)
	if err != nil {
		return err
	}
	var target *model.FlowTaskActor
	pendingCount := 0
	for i := range actors {
		if actors[i].ActorState == model.ActorPending {
			pendingCount++
		}
		if actors[i].ID == removeRowID {
			target = &actors[i]
		}
	}
	if target == nil {
		return fmt.Errorf("审批人不存在")
	}
	if target.ActorType != model.ActorTypeAppend {
		return fmt.Errorf("只能减签加签人员，不能移除原审批人")
	}
	if target.ActorState != model.ActorPending {
		return fmt.Errorf("只能减签未处理的加签人")
	}
	if target.ActorID == me.ActorID {
		return fmt.Errorf("不能减签自己")
	}
	if pendingCount <= 1 {
		return fmt.Errorf("至少保留一名待审批人")
	}
	return e.Repo.DeleteTaskActor(ctx, target.ID)
}

// AddRuntimeCC 审批中补抄送：归档抄送历史，进入「我收到的」
func (e *Engine) AddRuntimeCC(ctx context.Context, taskID, userID uint64, users []UserRef) error {
	if len(users) == 0 {
		return fmt.Errorf("请选择抄送人")
	}
	task, err := e.Repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	if task.TaskState != model.TaskActive || task.TaskType != model.TaskMajor {
		return fmt.Errorf("当前任务不可抄送")
	}
	inst, err := e.Repo.GetInstance(ctx, task.InstanceID)
	if err != nil {
		return err
	}
	if inst.InstanceState != model.InstActive {
		return fmt.Errorf("流程已结束")
	}
	mc, err := ParseModel(inst.ModelContent)
	if err != nil {
		return err
	}
	node := FindNode(mc.NodeConfig, task.NodeKey)
	if node == nil || !node.AllowCc {
		return fmt.Errorf("当前节点不允许抄送")
	}
	actors, err := e.Repo.ListTaskActors(ctx, taskID)
	if err != nil {
		return err
	}
	if _, err := e.resolveOperatorActor(ctx, inst, task, actors, userID); err != nil {
		return err
	}
	// 去重
	seen := map[uint64]bool{}
	uniq := make([]UserRef, 0, len(users))
	for _, u := range users {
		if u.ID == 0 || seen[u.ID] {
			continue
		}
		seen[u.ID] = true
		if u.Name == "" {
			if n, _, ok := e.Org.GetUser(u.ID); ok {
				u.Name = n
			}
		}
		uniq = append(uniq, u)
	}
	if len(uniq) == 0 {
		return fmt.Errorf("请选择抄送人")
	}
	now := time.Now()
	ccTask := &model.FlowTask{
		InstanceID:  inst.ID,
		NodeKey:     fmt.Sprintf("%s_rtcc_%d", task.NodeKey, now.UnixNano()),
		NodeName:    "抄送人",
		NodeType:    NodeCC,
		TaskType:    model.TaskCC,
		TaskState:   model.TaskComplete,
		ExamineMode: 1,
		FromNodeKey: task.NodeKey,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	ccActors := make([]model.FlowTaskActor, 0, len(uniq))
	for _, u := range uniq {
		ccActors = append(ccActors, model.FlowTaskActor{
			InstanceID: inst.ID,
			ActorID:    u.ID,
			ActorName:  u.Name,
			ActorState: model.ActorAgree,
			Opinion:    "抄送",
			FinishTime: &now,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	return e.Repo.ArchiveTask(ctx, ccTask, ccActors)
}

// Claim 认领：把角色槽转为当前用户待办
func (e *Engine) Claim(ctx context.Context, taskID, userID uint64) error {
	task, err := e.Repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	if task.TaskState != model.TaskActive {
		return fmt.Errorf("任务已结束")
	}
	actors, err := e.Repo.ListTaskActors(ctx, taskID)
	if err != nil {
		return err
	}
	roleIDs, err := e.Org.UserRoleIDs(userID)
	if err != nil {
		return err
	}
	roleSet := map[uint64]struct{}{}
	for _, id := range roleIDs {
		roleSet[id] = struct{}{}
	}
	// 已是本人待办则无需再认领
	for _, a := range actors {
		if a.ActorType != model.ActorTypeRole && a.ActorID == userID && a.ActorState == model.ActorPending {
			return fmt.Errorf("您已在该任务审批人中")
		}
	}
	var target *model.FlowTaskActor
	for i := range actors {
		a := &actors[i]
		if a.ActorState != model.ActorPending || a.ActorType != model.ActorTypeRole {
			continue
		}
		if _, ok := roleSet[a.ActorID]; !ok {
			continue
		}
		target = a
		break
	}
	if target == nil {
		return fmt.Errorf("没有可认领的任务")
	}
	name, _, ok := e.Org.GetUser(userID)
	if !ok {
		return fmt.Errorf("用户不存在")
	}
	now := time.Now()
	target.ActorID = userID
	target.ActorName = name
	target.ActorType = model.ActorTypeUser
	target.UpdatedAt = now
	return e.Repo.UpdateTaskActor(ctx, target)
}

// ListUrgeReceiverIDs 当前活动任务的待处理人（催办对象；角色槽展开为用户）
func (e *Engine) ListUrgeReceiverIDs(ctx context.Context, instanceID uint64) ([]uint64, error) {
	actives, err := e.Repo.ListActiveTasksByInstance(ctx, instanceID)
	if err != nil {
		return nil, err
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
	for _, t := range actives {
		if t.TaskType != model.TaskMajor {
			continue
		}
		actors, _ := e.Repo.ListTaskActors(ctx, t.ID)
		for _, a := range actors {
			if a.ActorState != model.ActorPending {
				continue
			}
			if a.ActorType == model.ActorTypeRole {
				uids, _ := e.Org.ListUserIDsByRole(a.ActorID)
				for _, uid := range uids {
					add(uid)
				}
				continue
			}
			add(a.ActorID)
		}
	}
	return ids, nil
}

// MarkUrged 记录催办时间（防刷）
func (e *Engine) MarkUrged(inst *model.FlowInstance) (ok bool, waitSec int) {
	v := parseVariable(inst)
	const minInterval = 300 // 5 分钟
	now := time.Now().Unix()
	if last, ok := v["lastUrgeAt"].(float64); ok {
		elapsed := now - int64(last)
		if elapsed < minInterval {
			return false, int(minInterval - elapsed)
		}
	}
	v["lastUrgeAt"] = now
	saveVariable(inst, v)
	return true, 0
}
