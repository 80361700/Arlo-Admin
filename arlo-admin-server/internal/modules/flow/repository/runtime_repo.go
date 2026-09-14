package repository

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"arlo-admin/internal/database"
	domainmodel "arlo-admin/internal/domain/model"
	"arlo-admin/internal/modules/flow/model"

	"gorm.io/gorm"
)

func (r *FlowRepository) CreateInstance(ctx context.Context, inst *model.FlowInstance) error {
	return r.db().WithContext(ctx).Create(inst).Error
}

func (r *FlowRepository) UpdateInstance(ctx context.Context, inst *model.FlowInstance) error {
	return r.db().WithContext(ctx).Save(inst).Error
}

func (r *FlowRepository) GetInstance(ctx context.Context, id uint64) (*model.FlowInstance, error) {
	var inst model.FlowInstance
	if err := r.db().WithContext(ctx).First(&inst, id).Error; err != nil {
		return nil, err
	}
	return &inst, nil
}

// DeleteInstanceHard 物理删除实例及其活动/历史任务（仅用于暂存草稿）
func (r *FlowRepository) DeleteInstanceHard(ctx context.Context, instanceID uint64) error {
	return r.db().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("instance_id = ?", instanceID).Delete(&model.FlowTaskActor{}).Error; err != nil {
			return err
		}
		if err := tx.Where("instance_id = ?", instanceID).Delete(&model.FlowTask{}).Error; err != nil {
			return err
		}
		if err := tx.Where("instance_id = ?", instanceID).Delete(&model.FlowHisTaskActor{}).Error; err != nil {
			return err
		}
		if err := tx.Where("instance_id = ?", instanceID).Delete(&model.FlowHisTask{}).Error; err != nil {
			return err
		}
		if err := tx.Where("instance_id = ?", instanceID).Delete(&model.FlowInstanceComment{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", instanceID).Delete(&model.FlowInstance{}).Error
	})
}

func (r *FlowRepository) CreateTask(ctx context.Context, task *model.FlowTask, actors []model.FlowTaskActor) error {
	return r.db().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		for i := range actors {
			actors[i].TaskID = task.ID
			actors[i].InstanceID = task.InstanceID
			if err := tx.Create(&actors[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *FlowRepository) GetTask(ctx context.Context, id uint64) (*model.FlowTask, error) {
	var t model.FlowTask
	if err := r.db().WithContext(ctx).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *FlowRepository) ListActiveTasksByInstance(ctx context.Context, instanceID uint64) ([]model.FlowTask, error) {
	var list []model.FlowTask
	err := r.db().WithContext(ctx).Where("instance_id = ? AND task_state = ?", instanceID, model.TaskActive).Find(&list).Error
	return list, err
}

func (r *FlowRepository) ListActiveTasksByGate(ctx context.Context, instanceID uint64, gateToken string) ([]model.FlowTask, error) {
	var list []model.FlowTask
	err := r.db().WithContext(ctx).
		Where("instance_id = ? AND gate_token = ? AND task_state = ?", instanceID, gateToken, model.TaskActive).
		Find(&list).Error
	return list, err
}

func (r *FlowRepository) UpdateTask(ctx context.Context, task *model.FlowTask) error {
	return r.db().WithContext(ctx).Save(task).Error
}

func (r *FlowRepository) DeleteTask(ctx context.Context, taskID uint64) error {
	return r.db().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("task_id = ?", taskID).Delete(&model.FlowTaskActor{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.FlowTask{}, taskID).Error
	})
}

func (r *FlowRepository) ListTaskActors(ctx context.Context, taskID uint64) ([]model.FlowTaskActor, error) {
	var list []model.FlowTaskActor
	err := r.db().WithContext(ctx).Where("task_id = ?", taskID).Order("weight ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *FlowRepository) UpdateTaskActor(ctx context.Context, actor *model.FlowTaskActor) error {
	return r.db().WithContext(ctx).Save(actor).Error
}

func (r *FlowRepository) CreateTaskActor(ctx context.Context, actor *model.FlowTaskActor) error {
	return r.db().WithContext(ctx).Create(actor).Error
}

func (r *FlowRepository) DeleteTaskActor(ctx context.Context, actorID uint64) error {
	return r.db().WithContext(ctx).Delete(&model.FlowTaskActor{}, actorID).Error
}

func (r *FlowRepository) ArchiveTask(ctx context.Context, task *model.FlowTask, actors []model.FlowTaskActor) error {
	now := time.Now()
	return r.db().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		his := model.FlowHisTask{
			TaskID:      task.ID,
			InstanceID:  task.InstanceID,
			NodeKey:     task.NodeKey,
			NodeName:    task.NodeName,
			NodeType:    task.NodeType,
			TaskType:    task.TaskType,
			TaskState:   task.TaskState,
			ExamineMode: task.ExamineMode,
			FromNodeKey: task.FromNodeKey,
			GateToken:   task.GateToken,
			Payload:     task.Payload,
			CreatedAt:   task.CreatedAt,
			FinishTime:  &now,
		}
		if err := tx.Create(&his).Error; err != nil {
			return err
		}
		for _, a := range actors {
			ha := model.FlowHisTaskActor{
				HisTaskID:  his.ID,
				TaskID:     task.ID,
				InstanceID: task.InstanceID,
				ActorID:    a.ActorID,
				ActorName:  a.ActorName,
				ActorType:  a.ActorType,
				ActorState: a.ActorState,
				Weight:     a.Weight,
				AgentID:    a.AgentID,
				Opinion:    a.Opinion,
				FinishTime: a.FinishTime,
				CreatedAt:  a.CreatedAt,
			}
			if ha.FinishTime == nil && a.ActorState != model.ActorPending {
				ha.FinishTime = &now
			}
			if err := tx.Create(&ha).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *FlowRepository) CountActiveChildren(ctx context.Context, parentID uint64, parentNodeKey string) (int64, error) {
	var n int64
	err := r.db().WithContext(ctx).Model(&model.FlowInstance{}).
		Where("parent_instance_id = ? AND parent_node_key = ? AND instance_state = ?", parentID, parentNodeKey, model.InstActive).
		Count(&n).Error
	return n, err
}

// ListActiveChildren 父实例下全部活动中的子实例（不限节点）
func (r *FlowRepository) ListActiveChildren(ctx context.Context, parentID uint64) ([]model.FlowInstance, error) {
	var list []model.FlowInstance
	err := r.db().WithContext(ctx).
		Where("parent_instance_id = ? AND instance_state = ?", parentID, model.InstActive).
		Order("id ASC").Find(&list).Error
	return list, err
}

// GetLatestChildByParentNode 父实例某子流程节点下最近一次子实例（含已结束）
func (r *FlowRepository) GetLatestChildByParentNode(ctx context.Context, parentID uint64, parentNodeKey string) (*model.FlowInstance, error) {
	var inst model.FlowInstance
	err := r.db().WithContext(ctx).
		Where("parent_instance_id = ? AND parent_node_key = ?", parentID, parentNodeKey).
		Order("id DESC").
		First(&inst).Error
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

// ApproveListFilter 审批列表通用筛选（流程名 / 创建人 / 实例状态 / 时间）
type ApproveListFilter struct {
	Keyword       string
	CreateBy      string
	InstanceState *int8
	Begin         *time.Time
	End           *time.Time
}

func applyJoinedInstanceFilters(q *gorm.DB, f ApproveListFilter, timeExpr string) *gorm.DB {
	if kw := strings.TrimSpace(f.Keyword); kw != "" {
		q = q.Where(sqlRootCol("i", "process_name")+" LIKE ?", "%"+kw+"%")
	}
	if by := strings.TrimSpace(f.CreateBy); by != "" {
		q = q.Where(sqlRootCol("i", "create_by")+" LIKE ?", "%"+by+"%")
	}
	if f.InstanceState != nil {
		q = q.Where(sqlRootCol("i", "instance_state")+" = ?", *f.InstanceState)
	}
	if timeExpr != "" {
		if f.Begin != nil {
			q = q.Where(timeExpr+" >= ?", *f.Begin)
		}
		if f.End != nil {
			q = q.Where(timeExpr+" <= ?", *f.End)
		}
	}
	return q
}

func applyInstanceModelFilters(tx *gorm.DB, f ApproveListFilter) *gorm.DB {
	if kw := strings.TrimSpace(f.Keyword); kw != "" {
		tx = tx.Where("process_name LIKE ?", "%"+kw+"%")
	}
	if by := strings.TrimSpace(f.CreateBy); by != "" {
		tx = tx.Where("create_by LIKE ?", "%"+by+"%")
	}
	if f.InstanceState != nil {
		tx = tx.Where("instance_state = ?", *f.InstanceState)
	}
	if f.Begin != nil {
		tx = tx.Where("created_at >= ?", *f.Begin)
	}
	if f.End != nil {
		tx = tx.Where("created_at <= ?", *f.End)
	}
	return tx
}

// ListPendingTasks 用户待办（主办任务 + pending actor）
// selfUserID：当前登录用户；principalIDs：委托给我的委托人（仅 allowDelegate≠false 的流程计入）
func (r *FlowRepository) ListPendingTasks(ctx context.Context, selfUserID uint64, principalIDs []uint64, f ApproveListFilter, page, pageSize int) ([]PendingTaskRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	userIDs := make([]uint64, 0, 1+len(principalIDs))
	userIDs = append(userIDs, selfUserID)
	for _, id := range principalIDs {
		if id != 0 && id != selfUserID {
			userIDs = append(userIDs, id)
		}
	}
	if len(userIDs) == 0 || selfUserID == 0 {
		return nil, 0, nil
	}
	db := database.DB.WithContext(ctx)
	type row struct {
		RowID            uint64    `gorm:"column:row_id"`
		ActorUserID      uint64    `gorm:"column:actor_user_id"`
		TaskID           uint64    `gorm:"column:task_id"`
		InstanceID       uint64    `gorm:"column:instance_id"`
		ParentInstanceID uint64    `gorm:"column:parent_instance_id"`
		NodeName         string    `gorm:"column:node_name"`
		WorkNodeName     string    `gorm:"column:work_node_name"`
		NodeKey          string    `gorm:"column:node_key"`
		ExamineMode      int8      `gorm:"column:examine_mode"`
		ProcessName      string    `gorm:"column:process_name"`
		SubProcessName   string    `gorm:"column:sub_process_name"`
		IsSubProcess     bool      `gorm:"column:is_sub_process"`
		CreateBy         string    `gorm:"column:create_by"`
		CreateID         uint64    `gorm:"column:create_id"`
		CreatedAt        time.Time `gorm:"column:created_at"`
		Weight           int       `gorm:"column:weight"`
		Variable         string    `gorm:"column:variable"`
	}
	// 本人槽位始终可见；代批仅当流程未显式关闭 allowDelegate（缺省视为允许）
	delegateOK := `(
		a.actor_id = ?
		OR CASE
			WHEN JSON_EXTRACT(i.variable, '$.processSetting.allowDelegate') IS NULL THEN 1
			WHEN JSON_TYPE(JSON_EXTRACT(i.variable, '$.processSetting.allowDelegate')) = 'NULL' THEN 1
			WHEN JSON_EXTRACT(i.variable, '$.processSetting.allowDelegate') = CAST('false' AS JSON) THEN 0
			WHEN JSON_UNQUOTE(JSON_EXTRACT(i.variable, '$.processSetting.allowDelegate')) IN ('false', '0') THEN 0
			ELSE 1
		END = 1
	)`
	q := db.Table("flow_task_actor AS a").
		Select(`a.id AS row_id, a.actor_id AS actor_user_id, a.task_id, a.instance_id, a.weight,
			i.parent_instance_id,
			t.node_name AS work_node_name, t.node_key, t.examine_mode, t.created_at,
			` + sqlDisplayNodeName("i", "parent", "t") + ` AS node_name,
			` + sqlDisplayProcessName("i", "parent") + ` AS process_name,
			` + sqlSubProcessName("i") + ` AS sub_process_name,
			` + sqlIsSubProcess("i") + ` AS is_sub_process,
			` + sqlRootCol("i", "create_by") + ` AS create_by,
			` + sqlRootCol("i", "create_id") + ` AS create_id,
			i.variable`).
		Joins("JOIN flow_task t ON t.id = a.task_id AND t.task_state = ? AND t.task_type = ?", model.TaskActive, model.TaskMajor).
		Joins("JOIN flow_instance i ON i.id = a.instance_id AND i.instance_state = ?", model.InstActive).
		Joins(sqlAncestorJoins).
		Where("a.actor_id IN ? AND a.actor_state = ? AND a.actor_type IN ?", userIDs, model.ActorPending, []int8{model.ActorTypeUser, model.ActorTypeAppend}).
		Where(delegateOK, selfUserID).
		// 本人已是该任务候选人时，不再展示同任务的代批行（避免一条待办出现两条）
		Where(`NOT (
			a.actor_id <> ?
			AND EXISTS (
				SELECT 1 FROM flow_task_actor a2
				WHERE a2.task_id = a.task_id
					AND a2.actor_id = ?
					AND a2.actor_state = ?
					AND a2.actor_type IN ?
			)
		)`, selfUserID, selfUserID, model.ActorPending, []int8{model.ActorTypeUser, model.ActorTypeAppend})
	q = applyJoinedInstanceFilters(q, f, "t.created_at")

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []row
	err := q.Order("t.created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]PendingTaskRow, 0, len(rows))
	for _, r0 := range rows {
		out = append(out, PendingTaskRow{
			ActorID:          r0.RowID,
			ActorUserID:      r0.ActorUserID,
			TaskID:           r0.TaskID,
			InstanceID:       r0.InstanceID,
			ParentInstanceID: r0.ParentInstanceID,
			NodeName:         r0.NodeName,
			WorkNodeName:     r0.WorkNodeName,
			NodeKey:          r0.NodeKey,
			ExamineMode:      r0.ExamineMode,
			ProcessName:      r0.ProcessName,
			SubProcessName:   r0.SubProcessName,
			IsSubProcess:     r0.IsSubProcess,
			CreateBy:         r0.CreateBy,
			CreateID:         r0.CreateID,
			CreatedAt:        r0.CreatedAt.Format("2006-01-02 15:04:05"),
			Weight:           r0.Weight,
			Variable:         r0.Variable,
		})
	}
	return out, total, nil
}

type PendingTaskRow struct {
	ActorID          uint64 `json:"actorId"`
	ActorUserID      uint64 `json:"actorUserId"`
	TaskID           uint64 `json:"taskId"`
	InstanceID       uint64 `json:"instanceId"`
	ParentInstanceID uint64 `json:"parentInstanceId"`
	NodeName         string `json:"nodeName"`
	WorkNodeName     string `json:"workNodeName"`
	NodeKey          string `json:"nodeKey"`
	ExamineMode      int8   `json:"examineMode"`
	ProcessName      string `json:"processName"`
	SubProcessName   string `json:"subProcessName"`
	IsSubProcess     bool   `json:"isSubProcess"`
	CreateBy         string `json:"createBy"`
	CreateID         uint64 `json:"createId"`
	CreatedAt        string `json:"createdAt"`
	Weight           int    `json:"weight"`
	Variable         string `json:"-"`
}

// 子流程待办节点展示：根流程当前节点 · 子内节点（如「子流程 · 审核人」）
func sqlDisplayNodeName(instAlias, parentAlias, taskAlias string) string {
	_ = parentAlias
	return `CASE WHEN ` + instAlias + `.parent_instance_id > 0 THEN CONCAT(IFNULL(NULLIF(` + sqlRootCol(instAlias, "current_node_name") + `,''),'子流程'), ' · ', ` + taskAlias + `.node_name) ELSE ` + taskAlias + `.node_name END`
}

// sqlRootCol 沿 parent→p2…p8 向上取根实例字段（嵌套子流程标题/并单用）
func sqlRootCol(instAlias, col string) string {
	return `CASE
		WHEN IFNULL(` + instAlias + `.parent_instance_id,0)=0 THEN ` + instAlias + `.` + col + `
		WHEN IFNULL(parent.parent_instance_id,0)=0 THEN parent.` + col + `
		WHEN IFNULL(p2.parent_instance_id,0)=0 THEN p2.` + col + `
		WHEN IFNULL(p3.parent_instance_id,0)=0 THEN p3.` + col + `
		WHEN IFNULL(p4.parent_instance_id,0)=0 THEN p4.` + col + `
		WHEN IFNULL(p5.parent_instance_id,0)=0 THEN p5.` + col + `
		WHEN IFNULL(p6.parent_instance_id,0)=0 THEN p6.` + col + `
		WHEN IFNULL(p7.parent_instance_id,0)=0 THEN p7.` + col + `
		WHEN IFNULL(p8.parent_instance_id,0)=0 THEN p8.` + col + `
		ELSE COALESCE(p8.` + col + `, p7.` + col + `, p6.` + col + `, p5.` + col + `, p4.` + col + `, p3.` + col + `, p2.` + col + `, parent.` + col + `, ` + instAlias + `.` + col + `)
	END`
}

// 子实例列表标题用根主流程名；自身名放到 sub_process_name
func sqlDisplayProcessName(instAlias, parentAlias string) string {
	_ = parentAlias
	return sqlRootCol(instAlias, "process_name")
}

func sqlSubProcessName(instAlias string) string {
	return `CASE WHEN ` + instAlias + `.parent_instance_id > 0 THEN ` + instAlias + `.process_name ELSE '' END`
}

func sqlIsSubProcess(instAlias string) string {
	return `CASE WHEN ` + instAlias + `.parent_instance_id > 0 THEN 1 ELSE 0 END`
}

// sqlAncestorJoins 自实例 i 向上挂 8 层祖先，供 sqlRootCol 取根
const sqlAncestorJoins = `
LEFT JOIN flow_instance parent ON parent.id = i.parent_instance_id
LEFT JOIN flow_instance p2 ON p2.id = parent.parent_instance_id
LEFT JOIN flow_instance p3 ON p3.id = p2.parent_instance_id
LEFT JOIN flow_instance p4 ON p4.id = p3.parent_instance_id
LEFT JOIN flow_instance p5 ON p5.id = p4.parent_instance_id
LEFT JOIN flow_instance p6 ON p6.id = p5.parent_instance_id
LEFT JOIN flow_instance p7 ON p7.id = p6.parent_instance_id
LEFT JOIN flow_instance p8 ON p8.id = p7.parent_instance_id`

// ListLaunchProcesses 可发起的启用流程
func (r *FlowRepository) ListLaunchProcesses(ctx context.Context, name string) ([]model.FlowProcess, error) {
	tx := r.db().WithContext(ctx).Where("process_state = ? AND process_type IN ?", 1, []string{"main", "business"})
	if name != "" {
		tx = tx.Where("process_name LIKE ?", "%"+name+"%")
	}
	var list []model.FlowProcess
	err := tx.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *FlowRepository) ListHisTasks(ctx context.Context, instanceID uint64) ([]model.FlowHisTask, error) {
	var list []model.FlowHisTask
	err := r.db().WithContext(ctx).Where("instance_id = ?", instanceID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *FlowRepository) ListHisActorsByInstance(ctx context.Context, instanceID uint64) ([]model.FlowHisTaskActor, error) {
	var list []model.FlowHisTaskActor
	err := r.db().WithContext(ctx).Where("instance_id = ?", instanceID).Order("id ASC").Find(&list).Error
	return list, err
}

// ListInstanceAgreedUserIDs 当前「审批段」内已同意的用户（用于重复节点跳过）。
// 不含发起人/改单重提的同意；驳回或回到发起人之后的历史同意也不再计入，
// 避免「驳回→修改重提」后同一审批人被误跳过直接通过。
func (r *FlowRepository) ListInstanceAgreedUserIDs(ctx context.Context, instanceID uint64) ([]uint64, error) {
	db := r.db().WithContext(ctx)
	// 水位：最近一次发起/改单重提，或最近一次驳回归档任务
	var sinceID uint64
	_ = db.Table("flow_his_task").
		Select("COALESCE(MAX(id), 0)").
		Where("instance_id = ? AND (node_type = ? OR task_state = ?)",
			instanceID, 0, model.TaskReject). // 0=发起人节点
		Scan(&sinceID)

	var ids []uint64
	err := db.Table("flow_his_task_actor AS a").
		Select("DISTINCT a.actor_id").
		Joins("JOIN flow_his_task h ON h.id = a.his_task_id").
		Where("a.instance_id = ? AND a.actor_type = ? AND a.actor_state = ? AND h.task_type = ? AND h.node_type = ? AND h.id > ?",
			instanceID, model.ActorTypeUser, model.ActorAgree, model.TaskMajor, 1, sinceID). // 1=审批节点
		Pluck("a.actor_id", &ids).Error
	return ids, err
}

func (r *FlowRepository) ListMyInstances(ctx context.Context, userID uint64, f ApproveListFilter, page, pageSize int) ([]model.FlowInstance, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	// 子流程实例不进「我的申请」，避免与主流程并列为两条
	tx := r.db().WithContext(ctx).Model(&model.FlowInstance{}).
		Where("create_id = ? AND parent_instance_id = 0", userID)
	tx = applyInstanceModelFilters(tx, f)
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FlowInstance
	err := tx.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ListMonitorInstances 流程监控：主实例，可按流程 id / 名称 / 创建人 / 状态 / 时间筛选
// onlyActive：未指定 instanceState 时，true 仅审批中，false 排除暂存草稿
func (r *FlowRepository) ListMonitorInstances(ctx context.Context, processIDs []uint64, allProcesses bool, f ApproveListFilter, onlyActive bool, page, pageSize int) ([]model.FlowInstance, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if !allProcesses && len(processIDs) == 0 {
		return []model.FlowInstance{}, 0, nil
	}
	tx := r.db().WithContext(ctx).Model(&model.FlowInstance{}).Where("parent_instance_id = 0")
	if !allProcesses {
		tx = tx.Where("process_id IN ?", processIDs)
	}
	if f.InstanceState != nil {
		tx = tx.Where("instance_state = ?", *f.InstanceState)
	} else if onlyActive {
		tx = tx.Where("instance_state = ?", model.InstActive)
	} else {
		// 监控不展示暂存草稿
		tx = tx.Where("instance_state >= ?", model.InstActive)
	}
	// 状态已在上方处理，避免 applyInstanceModelFilters 重复
	tx = applyInstanceModelFilters(tx, ApproveListFilter{
		Keyword:  f.Keyword,
		CreateBy: f.CreateBy,
		Begin:    f.Begin,
		End:      f.End,
	})
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FlowInstance
	err := tx.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *FlowRepository) ListApprovedByUser(ctx context.Context, userID uint64, f ApproveListFilter, page, pageSize int) ([]dtoApprovedRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	db := r.db().WithContext(ctx)
	type row struct {
		ActorRowID       uint64
		HisTaskID        uint64
		ActorID          uint64
		AgentID          uint64
		ActorState       int8
		FinishTime       *time.Time
		TaskID           uint64
		NodeName         string
		InstanceID       uint64
		ParentInstanceID uint64
		ProcessName      string
		ProcessKey       string
		CreateBy         string
		CreatedAt        time.Time
		InstanceState    int8
		SortID           uint64
	}

	// 本人槽位办理，或作为受托人代批完成（agent_id）
	actorFilter := "(a.actor_id = ? OR a.agent_id = ?)"

	// 子实例办理并到根主单：列表 instance/标题取根流程，不再单独占一条「子流程」
	displayCols := `
		` + sqlRootCol("i", "id") + ` AS instance_id,
		i.parent_instance_id AS parent_instance_id,
		` + sqlDisplayProcessName("i", "parent") + ` AS process_name,
		` + sqlRootCol("i", "process_key") + ` AS process_key,
		` + sqlRootCol("i", "create_by") + ` AS create_by,
		` + sqlRootCol("i", "created_at") + ` AS created_at,
		` + sqlRootCol("i", "instance_state") + ` AS instance_state`

	nodeColsHis := `CASE WHEN i.parent_instance_id > 0 THEN CONCAT('子流程 · ', h.node_name) ELSE h.node_name END AS node_name`
	nodeColsActive := `CASE WHEN i.parent_instance_id > 0 THEN CONCAT('子流程 · ', t.node_name) ELSE t.node_name END AS node_name`

	// 1) 节点已归档的历史办理
	hisQ := db.Table("flow_his_task_actor AS a").
		Select(`a.id AS actor_row_id, a.his_task_id, a.actor_id, a.agent_id, a.actor_state, a.finish_time, a.task_id, `+nodeColsHis+`, `+displayCols+`, a.id AS sort_id`).
		Joins("JOIN flow_his_task h ON h.id = a.his_task_id").
		Joins("JOIN flow_instance i ON i.id = a.instance_id").
		Joins(sqlAncestorJoins).
		Where(actorFilter+" AND a.actor_state IN ? AND h.task_type = ?", userID, userID, []int8{model.ActorAgree, model.ActorRefuse}, model.TaskMajor)
	hisQ = applyJoinedInstanceFilters(hisQ, ApproveListFilter{
		Keyword:       f.Keyword,
		CreateBy:      f.CreateBy,
		InstanceState: f.InstanceState,
	}, "")
	var hisRows []row
	err := hisQ.Scan(&hisRows).Error
	if err != nil {
		return nil, 0, err
	}

	// 2) 节点未结束但本人已同意/拒绝（会签/依次/加签未走完时也应进「已审批」）
	activeQ := db.Table("flow_task_actor AS a").
		Select(`a.id AS actor_row_id, 0 AS his_task_id, a.actor_id, a.agent_id, a.actor_state, a.finish_time, a.task_id, `+nodeColsActive+`, `+displayCols+`, a.id AS sort_id`).
		Joins("JOIN flow_task t ON t.id = a.task_id").
		Joins("JOIN flow_instance i ON i.id = a.instance_id").
		Joins(sqlAncestorJoins).
		Where(actorFilter+" AND a.actor_state IN ? AND t.task_type = ? AND t.task_state = ?",
			userID, userID, []int8{model.ActorAgree, model.ActorRefuse}, model.TaskMajor, model.TaskActive)
	activeQ = applyJoinedInstanceFilters(activeQ, ApproveListFilter{
		Keyword:       f.Keyword,
		CreateBy:      f.CreateBy,
		InstanceState: f.InstanceState,
	}, "")
	var activeRows []row
	err = activeQ.Scan(&activeRows).Error
	if err != nil {
		return nil, 0, err
	}

	all := append(hisRows, activeRows...)
	sort.Slice(all, func(i, j int) bool {
		ti, tj := all[i].FinishTime, all[j].FinishTime
		if ti != nil && tj != nil && !ti.Equal(*tj) {
			return ti.After(*tj)
		}
		if ti != nil && tj == nil {
			return true
		}
		if ti == nil && tj != nil {
			return false
		}
		return all[i].SortID > all[j].SortID
	})

	// 同一主单多条办理（含子流程内）合并为一条：优先保留主实例上的办理
	merged := make([]row, 0, len(all))
	idxByRoot := map[uint64]int{}
	for _, r0 := range all {
		rootID := r0.InstanceID
		if rootID == 0 {
			continue
		}
		if idx, ok := idxByRoot[rootID]; ok {
			prev := merged[idx]
			// 已有的是子流程并上来的，当前是主实例办理 → 替换
			if prev.ParentInstanceID > 0 && r0.ParentInstanceID == 0 {
				merged[idx] = r0
			}
			continue
		}
		idxByRoot[rootID] = len(merged)
		merged = append(merged, r0)
	}

	if kw := strings.TrimSpace(f.Keyword); kw != "" {
		kwLower := strings.ToLower(kw)
		filtered := merged[:0]
		for _, r0 := range merged {
			if strings.Contains(strings.ToLower(r0.ProcessName), kwLower) {
				filtered = append(filtered, r0)
			}
		}
		merged = filtered
	}

	if by := strings.TrimSpace(f.CreateBy); by != "" {
		byLower := strings.ToLower(by)
		filtered := merged[:0]
		for _, r0 := range merged {
			if strings.Contains(strings.ToLower(r0.CreateBy), byLower) {
				filtered = append(filtered, r0)
			}
		}
		merged = filtered
	}

	if f.InstanceState != nil {
		st := *f.InstanceState
		filtered := merged[:0]
		for _, r0 := range merged {
			if r0.InstanceState == st {
				filtered = append(filtered, r0)
			}
		}
		merged = filtered
	}

	if f.Begin != nil || f.End != nil {
		filtered := merged[:0]
		for _, r0 := range merged {
			t := r0.CreatedAt
			if r0.FinishTime != nil {
				t = *r0.FinishTime
			}
			if f.Begin != nil && t.Before(*f.Begin) {
				continue
			}
			if f.End != nil && t.After(*f.End) {
				continue
			}
			filtered = append(filtered, r0)
		}
		merged = filtered
	}

	total := int64(len(merged))
	start := (page - 1) * pageSize
	if start > len(merged) {
		start = len(merged)
	}
	endIdx := start + pageSize
	if endIdx > len(merged) {
		endIdx = len(merged)
	}
	pageRows := merged[start:endIdx]

	out := make([]dtoApprovedRow, 0, len(pageRows))
	for _, r0 := range pageRows {
		ft := ""
		if r0.FinishTime != nil {
			ft = r0.FinishTime.Format("2006-01-02 15:04:05")
		}
		out = append(out, dtoApprovedRow{
			ActorRowID:   r0.ActorRowID,
			HisTaskID:    r0.HisTaskID,
			InstanceID:   r0.InstanceID,
			ProcessName:  r0.ProcessName,
			ProcessKey:   r0.ProcessKey,
			NodeName:     r0.NodeName,
			TaskID:       r0.TaskID,
			ActorState:   r0.ActorState,
			InstanceState: r0.InstanceState,
			CreateBy:     r0.CreateBy,
			CreatedAt:    r0.CreatedAt.Format("2006-01-02 15:04:05"),
			FinishTime:   ft,
			// 当前用户是代批人（非槽位本人）
			IsDelegate: r0.AgentID == userID && r0.ActorID != userID,
			// 已并到主单视角，不再打子流程角标
			IsSubProcess:   false,
			SubProcessName: "",
		})
	}
	return out, total, nil
}

type dtoApprovedRow struct {
	ActorRowID     uint64
	HisTaskID      uint64
	InstanceID     uint64
	ProcessName    string
	SubProcessName string
	IsSubProcess   bool
	ProcessKey     string
	NodeName       string
	TaskID         uint64
	ActorState     int8
	InstanceState  int8
	CreateBy       string
	CreatedAt      string
	FinishTime     string
	IsDelegate     bool
	Read           bool
}

func (r *FlowRepository) ListReceivedCC(ctx context.Context, userID uint64, f ApproveListFilter, page, pageSize int) ([]dtoApprovedRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	db := r.db().WithContext(ctx)
	nameCols := sqlDisplayProcessName("i", "parent") + ` AS process_name, ` +
		sqlSubProcessName("i") + ` AS sub_process_name, ` +
		sqlIsSubProcess("i") + ` AS is_sub_process`
	q := db.Table("flow_his_task_actor AS a").
		Select(`a.id AS actor_row_id, a.his_task_id, a.actor_state, a.finish_time, a.task_id, a.read_at, h.node_name, h.instance_id, `+nameCols+`, `+sqlRootCol("i", "process_key")+` AS process_key, `+sqlRootCol("i", "create_by")+` AS create_by, i.created_at, i.instance_state`).
		Joins("JOIN flow_his_task h ON h.id = a.his_task_id").
		Joins("JOIN flow_instance i ON i.id = a.instance_id").
		Joins(sqlAncestorJoins).
		Where("a.actor_id = ? AND h.task_type = ?", userID, model.TaskCC)
	q = applyJoinedInstanceFilters(q, f, "COALESCE(a.finish_time, i.created_at)")
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	type row struct {
		ActorRowID     uint64
		HisTaskID      uint64
		ActorState     int8
		FinishTime     *time.Time
		TaskID         uint64
		ReadAt         *time.Time
		NodeName       string
		InstanceID     uint64
		ProcessName    string
		SubProcessName string
		IsSubProcess   bool
		ProcessKey     string
		CreateBy       string
		CreatedAt      time.Time
		InstanceState  int8
	}
	var rows []row
	err := q.Order("CASE WHEN a.read_at IS NULL THEN 0 ELSE 1 END ASC, a.id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]dtoApprovedRow, 0, len(rows))
	for _, r0 := range rows {
		ft := ""
		if r0.FinishTime != nil {
			ft = r0.FinishTime.Format("2006-01-02 15:04:05")
		}
		out = append(out, dtoApprovedRow{
			ActorRowID: r0.ActorRowID, HisTaskID: r0.HisTaskID,
			InstanceID: r0.InstanceID, ProcessName: r0.ProcessName, SubProcessName: r0.SubProcessName, IsSubProcess: r0.IsSubProcess, ProcessKey: r0.ProcessKey,
			NodeName: r0.NodeName, TaskID: r0.TaskID, ActorState: r0.ActorState,
			InstanceState: r0.InstanceState, CreateBy: r0.CreateBy,
			CreatedAt: r0.CreatedAt.Format("2006-01-02 15:04:05"), FinishTime: ft,
			Read: r0.ReadAt != nil,
		})
	}
	return out, total, nil
}

// ListClaimableByRoles 角色认领池任务：actor_type=角色槽，且角色在用户角色列表中
func (r *FlowRepository) ListClaimableByRoles(ctx context.Context, roleIDs []uint64, f ApproveListFilter, page, pageSize int) ([]dtoApprovedRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if len(roleIDs) == 0 {
		return []dtoApprovedRow{}, 0, nil
	}
	db := r.db().WithContext(ctx)
	nameCols := sqlDisplayProcessName("i", "parent") + ` AS process_name, ` +
		sqlSubProcessName("i") + ` AS sub_process_name, ` +
		sqlIsSubProcess("i") + ` AS is_sub_process`
	q := db.Table("flow_task_actor AS a").
		Select(`a.id AS actor_row_id, a.actor_state, a.finish_time, a.task_id, t.node_name, t.instance_id, `+nameCols+`, `+sqlRootCol("i", "process_key")+` AS process_key, `+sqlRootCol("i", "create_by")+` AS create_by, i.created_at, i.instance_state`).
		Joins("JOIN flow_task t ON t.id = a.task_id AND t.task_state = ? AND t.task_type = ?", model.TaskActive, model.TaskMajor).
		Joins("JOIN flow_instance i ON i.id = a.instance_id AND i.instance_state = ?", model.InstActive).
		Joins(sqlAncestorJoins).
		Where("a.actor_type = ? AND a.actor_state = ? AND a.actor_id IN ?", model.ActorTypeRole, model.ActorPending, roleIDs)
	q = applyJoinedInstanceFilters(q, f, "t.created_at")
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	type row struct {
		ActorRowID     uint64
		ActorState     int8
		FinishTime     *time.Time
		TaskID         uint64
		NodeName       string
		InstanceID     uint64
		ProcessName    string
		SubProcessName string
		IsSubProcess   bool
		ProcessKey     string
		CreateBy       string
		CreatedAt      time.Time
		InstanceState  int8
	}
	var rows []row
	err := q.Order("t.created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]dtoApprovedRow, 0, len(rows))
	for _, r0 := range rows {
		out = append(out, dtoApprovedRow{
			ActorRowID: r0.ActorRowID, InstanceID: r0.InstanceID, ProcessName: r0.ProcessName, SubProcessName: r0.SubProcessName, IsSubProcess: r0.IsSubProcess, ProcessKey: r0.ProcessKey,
			NodeName: r0.NodeName, TaskID: r0.TaskID, ActorState: r0.ActorState,
			InstanceState: r0.InstanceState, CreateBy: r0.CreateBy,
			CreatedAt: r0.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return out, total, nil
}

func (r *FlowRepository) ListExpiredDelayTasks(ctx context.Context, now time.Time, limit int) ([]model.FlowTask, error) {
	var list []model.FlowTask
	err := r.db().WithContext(ctx).
		Where("task_type = ? AND task_state = ? AND expire_time IS NOT NULL AND expire_time <= ?", model.TaskDelay, model.TaskActive, now).
		Limit(limit).Find(&list).Error
	return list, err
}

// ListExpiredMajorTimeoutTasks 审批任务超时（termAuto 写入的 expire_time）
func (r *FlowRepository) ListExpiredMajorTimeoutTasks(ctx context.Context, now time.Time, limit int) ([]model.FlowTask, error) {
	var list []model.FlowTask
	err := r.db().WithContext(ctx).
		Where("task_type = ? AND task_state = ? AND expire_time IS NOT NULL AND expire_time <= ?", model.TaskMajor, model.TaskActive, now).
		Order("expire_time ASC").Limit(limit).Find(&list).Error
	return list, err
}

// ListActiveMajorTasks 活动审批任务（用于扫描提醒）
func (r *FlowRepository) ListActiveMajorTasks(ctx context.Context, limit int) ([]model.FlowTask, error) {
	var list []model.FlowTask
	err := r.db().WithContext(ctx).
		Where("task_type = ? AND task_state = ?", model.TaskMajor, model.TaskActive).
		Order("id ASC").Limit(limit).Find(&list).Error
	return list, err
}

// ListActiveSubTasks 活动子流程占位任务（用于补启未创建的子实例）
func (r *FlowRepository) ListActiveSubTasks(ctx context.Context, limit int) ([]model.FlowTask, error) {
	var list []model.FlowTask
	err := r.db().WithContext(ctx).
		Where("task_type = ? AND task_state = ?", model.TaskSub, model.TaskActive).
		Order("id ASC").Limit(limit).Find(&list).Error
	return list, err
}

func (r *FlowRepository) CreateInstanceComment(ctx context.Context, c *model.FlowInstanceComment) error {
	return r.db().WithContext(ctx).Create(c).Error
}

func (r *FlowRepository) ListInstanceComments(ctx context.Context, instanceID uint64) ([]model.FlowInstanceComment, error) {
	var list []model.FlowInstanceComment
	err := r.db().WithContext(ctx).Where("instance_id = ?", instanceID).Order("id ASC").Find(&list).Error
	return list, err
}

// MarkReceivedCCRead 标记抄送参与人为已读（仅本人、仅抄送任务）
func (r *FlowRepository) MarkReceivedCCRead(ctx context.Context, actorRowID, userID uint64) error {
	if actorRowID == 0 || userID == 0 {
		return nil
	}
	now := time.Now()
	return r.db().WithContext(ctx).Exec(`
UPDATE flow_his_task_actor AS a
INNER JOIN flow_his_task AS h ON h.id = a.his_task_id
SET a.read_at = ?
WHERE a.id = ? AND a.actor_id = ? AND h.task_type = ? AND a.read_at IS NULL
`, now, actorRowID, userID, model.TaskCC).Error
}

// --- OrgStore 实现 ---

type OrgStoreImpl struct{}

func NewOrgStore() *OrgStoreImpl { return &OrgStoreImpl{} }

func (o *OrgStoreImpl) db() *gorm.DB { return database.DB }

func (o *OrgStoreImpl) GetUser(id uint64) (name string, deptID uint64, ok bool) {
	var u domainmodel.User
	if err := o.db().Select("id", "name", "dept_id").First(&u, id).Error; err != nil {
		return "", 0, false
	}
	return u.Name, u.DeptID, true
}

func (o *OrgStoreImpl) ListUserIDsByRole(roleID uint64) ([]uint64, error) {
	var ids []uint64
	err := o.db().Model(&domainmodel.UserRole{}).Where("role_id = ?", roleID).Pluck("user_id", &ids).Error
	return ids, err
}

func (o *OrgStoreImpl) ListUserIDsByDept(deptID uint64) ([]uint64, error) {
	var ids []uint64
	err := o.db().Model(&domainmodel.User{}).Where("dept_id = ? AND status = 1", deptID).Pluck("id", &ids).Error
	return ids, err
}

func (o *OrgStoreImpl) GetDeptLeader(deptID uint64) (leaderID uint64, ok bool) {
	var d domainmodel.Dept
	if err := o.db().Select("id", "leader_id").First(&d, deptID).Error; err != nil {
		return 0, false
	}
	return d.LeaderID, d.LeaderID > 0
}

func (o *OrgStoreImpl) GetDeptParent(deptID uint64) (parentID uint64, ok bool) {
	var d domainmodel.Dept
	if err := o.db().Select("id", "parent_id").First(&d, deptID).Error; err != nil {
		return 0, false
	}
	return d.ParentID, true
}

func (o *OrgStoreImpl) UserRoleIDs(userID uint64) ([]uint64, error) {
	var ids []uint64
	err := o.db().Model(&domainmodel.UserRole{}).Where("user_id = ?", userID).Pluck("role_id", &ids).Error
	return ids, err
}

func (o *OrgStoreImpl) UserInDept(userID, deptID uint64) bool {
	_, d, ok := o.GetUser(userID)
	if !ok {
		return false
	}
	if d == deptID {
		return true
	}
	// 向上匹配祖先部门
	cur := d
	for i := 0; i < 32 && cur > 0; i++ {
		if cur == deptID {
			return true
		}
		p, ok := o.GetDeptParent(cur)
		if !ok || p == 0 || p == cur {
			break
		}
		cur = p
	}
	return false
}

// ---------- 审批委托 ----------

func (r *FlowRepository) GetUserDelegate(ctx context.Context, userID uint64) (*model.FlowUserDelegate, error) {
	var d model.FlowUserDelegate
	err := r.db().WithContext(ctx).Where("user_id = ?", userID).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *FlowRepository) UpsertUserDelegate(ctx context.Context, d *model.FlowUserDelegate) error {
	var old model.FlowUserDelegate
	err := r.db().WithContext(ctx).Where("user_id = ?", d.UserID).First(&old).Error
	now := time.Now()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		d.CreatedAt = now
		d.UpdatedAt = now
		return r.db().WithContext(ctx).Create(d).Error
	}
	if err != nil {
		return err
	}
	old.ToUserID = d.ToUserID
	old.ToUserName = d.ToUserName
	old.Enabled = d.Enabled
	old.Remark = d.Remark
	old.UpdatedAt = now
	return r.db().WithContext(ctx).Save(&old).Error
}

func (r *FlowRepository) ClearUserDelegate(ctx context.Context, userID uint64) error {
	return r.db().WithContext(ctx).Where("user_id = ?", userID).Delete(&model.FlowUserDelegate{}).Error
}

// ListPrincipalIDsByDelegate 谁把审批委托给了 delegateUserID
func (r *FlowRepository) ListPrincipalIDsByDelegate(ctx context.Context, delegateUserID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db().WithContext(ctx).Model(&model.FlowUserDelegate{}).
		Where("to_user_id = ? AND enabled = 1 AND user_id <> ?", delegateUserID, delegateUserID).
		Pluck("user_id", &ids).Error
	return ids, err
}
