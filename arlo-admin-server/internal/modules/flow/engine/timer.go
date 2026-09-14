package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"arlo-admin/internal/modules/flow/model"
)

// applyApprovalTimers 按节点 term*/remind* 写入任务到期与提醒信息（合并写 payload，保留 gateStack 等）
func applyApprovalTimers(task *model.FlowTask, node *Node, now time.Time) {
	if task == nil || node == nil {
		return
	}
	p := map[string]interface{}{}
	if task.Payload != "" {
		_ = json.Unmarshal([]byte(task.Payload), &p)
	}
	changed := false
	if node.TermAuto && node.Term > 0 {
		exp := now.Add(time.Duration(node.Term) * time.Hour)
		task.ExpireTime = &exp
		p["termAuto"] = true
		p["termMode"] = node.TermMode // 0=自动通过 1=自动拒绝
		changed = true
	}
	if node.Remind {
		if at, ok := calcRemindAt(node, now); ok {
			p["remindAt"] = at.Format(time.RFC3339)
			p["reminded"] = false
			if node.DelayType == 2 {
				p["remindDaily"] = true
				if node.ExtendConfig != nil {
					if s, ok := node.ExtendConfig["remindTime"].(string); ok {
						p["remindClock"] = strings.TrimSpace(s)
					}
				}
			} else {
				p["remindDaily"] = false
				delete(p, "remindClock")
			}
			changed = true
		}
	}
	if !changed {
		return
	}
	b, _ := json.Marshal(p)
	task.Payload = string(b)
}

func calcRemindAt(node *Node, now time.Time) (time.Time, bool) {
	if node == nil {
		return time.Time{}, false
	}
	raw := ""
	if node.ExtendConfig != nil {
		if s, ok := node.ExtendConfig["remindTime"].(string); ok {
			raw = strings.TrimSpace(s)
		}
	}
	switch node.DelayType {
	case 2: // 自动计算：当日/次日的 HH:mm:ss
		if raw == "" {
			return time.Time{}, false
		}
		parts := strings.Split(raw, ":")
		if len(parts) < 2 {
			return time.Time{}, false
		}
		var h, m, s int
		fmt.Sscanf(parts[0], "%d", &h)
		fmt.Sscanf(parts[1], "%d", &m)
		if len(parts) >= 3 {
			fmt.Sscanf(parts[2], "%d", &s)
		}
		at := time.Date(now.Year(), now.Month(), now.Day(), h, m, s, 0, now.Location())
		if !at.After(now) {
			at = at.Add(24 * time.Hour)
		}
		return at, true
	default: // 固定时长：N:d|h|m（与延时节点一致）
		if raw == "" {
			return now.Add(time.Hour), true
		}
		return now.Add(parseRemindDuration(raw)), true
	}
}

func parseRemindDuration(raw string) time.Duration {
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

type taskPayload struct {
	TermAuto    bool   `json:"termAuto"`
	TermMode    int    `json:"termMode"`
	RemindAt    string `json:"remindAt"`
	Reminded    bool   `json:"reminded"`
	RemindDaily bool   `json:"remindDaily"` // delayType=2：每日到点提醒
	RemindClock string `json:"remindClock"` // HH:mm:ss
}

func parseTaskPayload(raw string) taskPayload {
	var p taskPayload
	if raw == "" {
		return p
	}
	_ = json.Unmarshal([]byte(raw), &p)
	return p
}

func writeTaskPayload(task *model.FlowTask, p taskPayload) {
	if task == nil {
		return
	}
	m := map[string]interface{}{}
	if task.Payload != "" {
		_ = json.Unmarshal([]byte(task.Payload), &m)
	}
	if p.TermAuto {
		m["termAuto"] = p.TermAuto
		m["termMode"] = p.TermMode
	}
	if p.RemindAt != "" {
		m["remindAt"] = p.RemindAt
	}
	m["reminded"] = p.Reminded
	m["remindDaily"] = p.RemindDaily
	if p.RemindDaily && p.RemindClock != "" {
		m["remindClock"] = p.RemindClock
	} else {
		delete(m, "remindClock")
	}
	b, _ := json.Marshal(m)
	task.Payload = string(b)
}

// ExecuteTimeout 审批节点超时：自动通过或自动拒绝
func (e *Engine) ExecuteTimeout(ctx context.Context, taskID uint64) error {
	task, err := e.Repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	if task.TaskType != model.TaskMajor || task.TaskState != model.TaskActive {
		return nil
	}
	if task.ExpireTime == nil || task.ExpireTime.After(time.Now()) {
		return nil
	}
	p := parseTaskPayload(task.Payload)
	if !p.TermAuto {
		// 兼容：有 expire_time 的主任务也按超时处理
		p.TermAuto = true
	}

	inst, err := e.Repo.GetInstance(ctx, task.InstanceID)
	if err != nil {
		return err
	}
	if inst.InstanceState != model.InstActive {
		return nil
	}
	actors, err := e.Repo.ListTaskActors(ctx, taskID)
	if err != nil {
		return err
	}
	mc, err := ParseModel(inst.ModelContent)
	if err != nil {
		return err
	}
	node := FindNode(mc.NodeConfig, task.NodeKey)
	now := time.Now()
	op := UserRef{ID: 0, Name: "系统"}

	if p.TermMode == 1 {
		return e.timeoutReject(ctx, inst, mc, task, actors, node, op, now)
	}
	return e.timeoutPass(ctx, inst, mc, task, actors, node, op, now)
}

func (e *Engine) timeoutPass(
	ctx context.Context,
	inst *model.FlowInstance,
	mc *ModelContent,
	task *model.FlowTask,
	actors []model.FlowTaskActor,
	node *Node,
	op UserRef,
	now time.Time,
) error {
	opinion := "超时自动通过"
	for i := range actors {
		if actors[i].ActorState != model.ActorPending {
			continue
		}
		actors[i].ActorState = model.ActorAgree
		actors[i].Opinion = opinion
		actors[i].FinishTime = &now
		actors[i].UpdatedAt = now
		_ = e.Repo.UpdateTaskActor(ctx, &actors[i])
	}
	actors, _ = e.Repo.ListTaskActors(ctx, task.ID)
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
	var child *Node
	if node != nil {
		child = node.ChildNode
	}
	return e.continueAfterLeaf(ctx, st, task.NodeKey, task.NodeName, child)
}

func (e *Engine) timeoutReject(
	ctx context.Context,
	inst *model.FlowInstance,
	mc *ModelContent,
	task *model.FlowTask,
	actors []model.FlowTaskActor,
	node *Node,
	op UserRef,
	now time.Time,
) error {
	opinion := "超时自动拒绝"
	for i := range actors {
		if actors[i].ActorState != model.ActorPending {
			continue
		}
		actors[i].ActorState = model.ActorRefuse
		actors[i].Opinion = opinion
		actors[i].FinishTime = &now
		actors[i].UpdatedAt = now
		_ = e.Repo.UpdateTaskActor(ctx, &actors[i])
	}
	actors, _ = e.Repo.ListTaskActors(ctx, task.ID)

	// 按节点驳回策略处理；终止时实例态为超时
	if node == nil {
		task.TaskState = model.TaskReject
		task.UpdatedAt = now
		_ = e.Repo.ArchiveTask(ctx, task, actors)
		_ = e.Repo.DeleteTask(ctx, task.ID)
		actives, _ := e.Repo.ListActiveTasksByInstance(ctx, inst.ID)
		for _, t := range actives {
			as, _ := e.Repo.ListTaskActors(ctx, t.ID)
			t.TaskState = model.TaskTerminate
			_ = e.Repo.ArchiveTask(ctx, &t, as)
			_ = e.Repo.DeleteTask(ctx, t.ID)
		}
		return e.finishInstance(ctx, inst, model.InstTimeout, task.NodeKey, task.NodeName)
	}
	return e.handleRejectWithTerminateState(ctx, inst, mc, task, actors, node, op, opinion, 0, "", model.InstTimeout)
}

// RemindDue 若到提醒时间且（一次性未提醒 / 每日到点），返回接收人用户 id；否则返回 nil
func (e *Engine) RemindDue(ctx context.Context, task *model.FlowTask) ([]uint64, error) {
	if task == nil || task.TaskType != model.TaskMajor || task.TaskState != model.TaskActive {
		return nil, nil
	}
	p := parseTaskPayload(task.Payload)
	if p.RemindAt == "" {
		return nil, nil
	}
	if !p.RemindDaily && p.Reminded {
		return nil, nil
	}
	at, err := time.Parse(time.RFC3339, p.RemindAt)
	if err != nil {
		return nil, nil
	}
	if at.After(time.Now()) {
		return nil, nil
	}
	return e.listTaskPendingUserIDs(ctx, task.ID)
}

// MarkReminded 提醒发送后：一次性标记已提醒；每日模式推进到下一提醒时刻
func (e *Engine) MarkReminded(ctx context.Context, task *model.FlowTask) error {
	if task == nil {
		return nil
	}
	p := parseTaskPayload(task.Payload)
	now := time.Now()
	if p.RemindDaily {
		p.RemindAt = nextDailyRemindAt(p, now).Format(time.RFC3339)
		p.Reminded = false
	} else {
		p.Reminded = true
	}
	writeTaskPayload(task, p)
	task.UpdatedAt = now
	return e.Repo.UpdateTask(ctx, task)
}

// nextDailyRemindAt 在上次提醒时刻基础上 +24h，并追赶到「严格晚于 now」
func nextDailyRemindAt(p taskPayload, now time.Time) time.Time {
	var base time.Time
	if at, err := time.Parse(time.RFC3339, p.RemindAt); err == nil {
		base = at
	} else if p.RemindClock != "" {
		parts := strings.Split(p.RemindClock, ":")
		var h, m, s int
		if len(parts) >= 1 {
			fmt.Sscanf(parts[0], "%d", &h)
		}
		if len(parts) >= 2 {
			fmt.Sscanf(parts[1], "%d", &m)
		}
		if len(parts) >= 3 {
			fmt.Sscanf(parts[2], "%d", &s)
		}
		base = time.Date(now.Year(), now.Month(), now.Day(), h, m, s, 0, now.Location())
	} else {
		base = now
	}
	next := base.Add(24 * time.Hour)
	for !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func (e *Engine) listTaskPendingUserIDs(ctx context.Context, taskID uint64) ([]uint64, error) {
	actors, err := e.Repo.ListTaskActors(ctx, taskID)
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
	return ids, nil
}
