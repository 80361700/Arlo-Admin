package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"arlo-admin/internal/modules/flow/dto"
	"arlo-admin/internal/modules/flow/engine"
	"arlo-admin/internal/modules/flow/model"
	"arlo-admin/internal/modules/flow/repository"
	msgdto "arlo-admin/internal/modules/message/dto"
	msgrepo "arlo-admin/internal/modules/message/repository"
	msgsvc "arlo-admin/internal/modules/message/service"
	"arlo-admin/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrProcessDisabled = errors.New("流程未启用或不存在")
	ErrTaskNotFound    = errors.New("任务不存在")
	ErrCannotHandle    = errors.New("当前不可处理该任务")
)

func (s *FlowService) org() engine.OrgStore {
	return repository.NewOrgStore()
}

type flowMessageNotifier struct{}

func (flowMessageNotifier) Notify(ctx context.Context, title, content string, receiverIDs []uint64) error {
	if len(receiverIDs) == 0 {
		return nil
	}
	msgSvc := msgsvc.NewMessageService(msgrepo.NewMessageRepository())
	return msgSvc.Send(ctx, &msgdto.SendMessageRequest{
		Title:       title,
		Content:     content,
		Type:        1,
		ReceiverIDs: receiverIDs,
	}, 0, "系统")
}

type flowDelegateStore struct {
	repo *repository.FlowRepository
}

func (d flowDelegateStore) PrincipalsOf(ctx context.Context, delegateUserID uint64) []uint64 {
	ids, err := d.repo.ListPrincipalIDsByDelegate(ctx, delegateUserID)
	if err != nil {
		return nil
	}
	return ids
}

func (s *FlowService) engine() *engine.Engine {
	return &engine.Engine{
		Repo:      s.repo,
		Org:       s.org(),
		Notify:    flowMessageNotifier{},
		Delegates: flowDelegateStore{repo: s.repo},
	}
}

func (s *FlowService) principalsOf(ctx context.Context, userID uint64) []uint64 {
	ids, _ := s.repo.ListPrincipalIDsByDelegate(ctx, userID)
	return ids
}

func (s *FlowService) actPrincipals(ctx context.Context, inst *model.FlowInstance, userID uint64) []uint64 {
	if inst == nil || !engine.ProcessSettingBool(inst, "allowDelegate", true) {
		return nil
	}
	return s.principalsOf(ctx, userID)
}

func (s *FlowService) ListLaunchProcesses(ctx context.Context, name string, userID uint64) ([]dto.LaunchProcessItem, error) {
	list, err := s.repo.ListLaunchProcesses(ctx, strings.TrimSpace(name))
	if err != nil {
		return nil, err
	}
	cats, _ := s.repo.ListCategories(ctx)
	catName := map[uint64]string{}
	for _, c := range cats {
		catName[c.ID] = c.Name
	}
	roleIDs, _ := s.org().UserRoleIDs(userID)
	roleSet := map[uint64]struct{}{}
	for _, id := range roleIDs {
		roleSet[id] = struct{}{}
	}
	out := make([]dto.LaunchProcessItem, 0, len(list))
	for _, p := range list {
		if !canUserLaunch(p.ModelContent, roleSet) {
			continue
		}
		icon, color := decodeIcon(p.ProcessIcon)
		cname := catName[p.CategoryID]
		if cname == "" {
			cname = "未分类"
		}
		out = append(out, dto.LaunchProcessItem{
			ProcessID:      p.ID,
			CategoryID:     p.CategoryID,
			CategoryName:   cname,
			ProcessKey:     p.ProcessKey,
			ProcessName:    p.ProcessName,
			ProcessIcon:    icon,
			ProcessBgcolor: color,
			ProcessType:    p.ProcessType,
			Remark:         p.Remark,
		})
	}
	catSort := map[uint64]int{}
	for _, c := range cats {
		catSort[c.ID] = c.Sort
	}
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := catSort[out[i].CategoryID], catSort[out[j].CategoryID]
		if si != sj {
			return si < sj
		}
		if out[i].CategoryID != out[j].CategoryID {
			return out[i].CategoryID < out[j].CategoryID
		}
		return out[i].ProcessID > out[j].ProcessID
	})
	return out, nil
}

// canUserLaunch 发起人节点未指定角色则可全员发起；指定了则需命中角色
func canUserLaunch(modelJSON string, roleSet map[uint64]struct{}) bool {
	mc, err := engine.ParseModel(modelJSON)
	if err != nil || mc.NodeConfig == nil {
		return true
	}
	start := mc.NodeConfig
	if len(start.NodeAssigneeList) == 0 {
		return true
	}
	for _, a := range start.NodeAssigneeList {
		id := engine.AssigneeIDUint(a)
		if _, ok := roleSet[id]; ok {
			return true
		}
	}
	return false
}

func (s *FlowService) GetLaunchForm(ctx context.Context, processID, userID uint64) (*dto.LaunchFormDetail, error) {
	p, err := s.repo.GetProcess(ctx, processID)
	if err != nil {
		return nil, err
	}
	if p.ProcessState != 1 {
		return nil, ErrProcessDisabled
	}
	roleIDs, _ := s.org().UserRoleIDs(userID)
	roleSet := map[uint64]struct{}{}
	for _, id := range roleIDs {
		roleSet[id] = struct{}{}
	}
	if !canUserLaunch(p.ModelContent, roleSet) {
		return nil, fmt.Errorf("无权发起该流程")
	}
	formMap := unmarshalMap(p.ProcessForm)
	renderType, component := "designer", ""
	// 业务流程：使用绑定表单 schema / 系统表单组件
	if p.ProcessType == "business" {
		if fid, ok := parseBindFormID(p.ProcessSetting); ok && fid > 0 {
			if f, e := s.repo.GetForm(ctx, fid); e == nil && f != nil {
				if f.FormType == 2 && strings.TrimSpace(f.PcURL) != "" {
					renderType = "vue"
					component = normalizeFormPcURL(f.PcURL)
				} else if f.FormSchema != "" {
					formMap = unmarshalMap(f.FormSchema)
				}
			}
		}
	}
	if renderType == "designer" {
		if mc, e := engine.ParseModel(p.ModelContent); e == nil && mc != nil && mc.NodeConfig != nil {
			formMap = s.mergeActionURLForm(ctx, formMap, mc.NodeConfig.ActionURL)
		}
	} else if renderType == "vue" {
		// 系统表单主表 + 发起节点设计子表单：子表 schema 单独下发
		if mc, e := engine.ParseModel(p.ModelContent); e == nil && mc != nil && mc.NodeConfig != nil {
			formMap = s.mergeActionURLForm(ctx, map[string]interface{}{}, mc.NodeConfig.ActionURL)
		}
	}
	detail := &dto.LaunchFormDetail{
		ProcessID:      p.ID,
		ProcessKey:     p.ProcessKey,
		ProcessName:    p.ProcessName,
		ProcessType:    p.ProcessType,
		ProcessForm:    formMap,
		ModelContent:   unmarshalMap(p.ModelContent),
		FormRenderType: renderType,
		FormComponent:  component,
	}
	if mc, e := engine.ParseModel(p.ModelContent); e == nil {
		detail.SelectionNodes = collectSelectionNodes(mc.NodeConfig)
	}
	return detail, nil
}

func collectSelectionNodes(root *engine.Node) []dto.LaunchSelectionNode {
	var out []dto.LaunchSelectionNode
	var walk func(n *engine.Node)
	walk = func(n *engine.Node) {
		if n == nil {
			return
		}
		if n.Type == engine.NodeApproval && n.SetType == 4 {
			mode := n.SelectMode
			if mode == 0 {
				mode = 1
			}
			out = append(out, dto.LaunchSelectionNode{
				NodeKey: n.NodeKey, NodeName: n.NodeName, Kind: "assignee", SelectMode: mode, Required: true,
			})
		}
		if n.Type == engine.NodeCC && n.AllowSelection {
			out = append(out, dto.LaunchSelectionNode{
				NodeKey: n.NodeKey, NodeName: n.NodeName, Kind: "cc", SelectMode: 2, Required: false,
			})
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

// validateLaunchCandidates 发起人自选：人数/角色模式 + 候选名单
func validateLaunchCandidates(root *engine.Node, req *dto.LaunchRequest) error {
	var walk func(n *engine.Node) error
	walk = func(n *engine.Node) error {
		if n == nil {
			return nil
		}
		if n.Type == engine.NodeApproval && n.SetType == 4 {
			list := req.NodeAssignees[n.NodeKey]
			mode := n.SelectMode
			if mode == 0 {
				mode = 1
			}
			if mode == 1 && len(list) > 1 {
				return fmt.Errorf("节点「%s」仅可选择一人", n.NodeName)
			}
			selected := make([]engine.UserRef, 0, len(list))
			for _, it := range list {
				selected = append(selected, engine.UserRef{ID: it.ID, Name: it.Name})
			}
			if err := engine.ValidateNodeCandidates(n, selected); err != nil {
				return err
			}
		}
		if err := walk(n.ChildNode); err != nil {
			return err
		}
		for _, c := range n.ConditionNodes {
			if err := walk(c); err != nil {
				return err
			}
			if err := walk(c.ChildNode); err != nil {
				return err
			}
		}
		for _, c := range n.ParallelNodes {
			if err := walk(c); err != nil {
				return err
			}
			if err := walk(c.ChildNode); err != nil {
				return err
			}
		}
		for _, c := range n.InclusiveNodes {
			if err := walk(c); err != nil {
				return err
			}
			if err := walk(c.ChildNode); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(root)
}

func (s *FlowService) Launch(ctx context.Context, req *dto.LaunchRequest, userID uint64, userName string) (uint64, error) {
	p, err := s.repo.GetProcess(ctx, req.ProcessID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrProcessDisabled
		}
		return 0, err
	}
	if p.ProcessState != 1 {
		return 0, ErrProcessDisabled
	}
	if p.ModelContent == "" {
		return 0, fmt.Errorf("流程模型未配置")
	}

	roleIDs, _ := s.org().UserRoleIDs(userID)
	roleSet := map[uint64]struct{}{}
	for _, id := range roleIDs {
		roleSet[id] = struct{}{}
	}
	if !canUserLaunch(p.ModelContent, roleSet) {
		return 0, fmt.Errorf("无权发起该流程")
	}

	// 校验发起人自选节点（暂存跳过）
	if !req.SaveAsDraft {
		if mc, e := engine.ParseModel(p.ModelContent); e == nil {
			for _, sel := range collectSelectionNodes(mc.NodeConfig) {
				if !sel.Required {
					continue
				}
				list := req.NodeAssignees[sel.NodeKey]
				if sel.Kind == "cc" {
					list = req.NodeCCUsers[sel.NodeKey]
				}
				if len(list) == 0 {
					label := "审批人"
					if sel.Kind == "cc" {
						label = "抄送人"
					}
					return 0, fmt.Errorf("请为节点「%s」选择%s", sel.NodeName, label)
				}
			}
			if err := validateLaunchCandidates(mc.NodeConfig, req); err != nil {
				return 0, err
			}
		}
	}

	// 按发起人节点 formConfig 剥离只读/隐藏字段写入；正式发起再校验必填
	processForm := p.ProcessForm
	if p.ProcessType == "business" {
		if fid, ok := parseBindFormID(p.ProcessSetting); ok && fid > 0 {
			if f, e := s.repo.GetForm(ctx, fid); e == nil && f != nil {
				if f.FormType == 2 {
					processForm = `{}`
				} else if f.FormSchema != "" {
					processForm = f.FormSchema
				}
			}
		}
	}
	schema := unmarshalMap(processForm)
	if mc, e := engine.ParseModel(p.ModelContent); e == nil && mc != nil && mc.NodeConfig != nil {
		schema = s.mergeActionURLForm(ctx, schema, mc.NodeConfig.ActionURL)
	}
	cfg := startFormConfig(p.ModelContent)
	cfg = appendMissingFormConfig(cfg, schema, 1)
	req.FormData = applyFormConfigWrite(nil, req.FormData, cfg)
	if !req.SaveAsDraft {
		if err := validateRequiredFields(schema, req.FormData, cfg, true); err != nil {
			return 0, err
		}
	}

	_, deptID, _ := s.org().GetUser(userID)
	formBytes, _ := json.Marshal(req.FormData)
	if req.FormData == nil {
		formBytes = []byte("{}")
	}

	state := model.InstActive
	curKey, curName := "", ""
	if req.SaveAsDraft {
		state = model.InstDraft
		if mc, e := engine.ParseModel(p.ModelContent); e == nil && mc != nil && mc.NodeConfig != nil {
			curKey = mc.NodeConfig.NodeKey
			curName = mc.NodeConfig.NodeName
			if curName == "" {
				curName = "发起人"
			}
			curName = curName + "（暂存）"
		} else {
			curName = "暂存"
		}
	}

	inst := &model.FlowInstance{
		ProcessID:       p.ID,
		ProcessKey:      p.ProcessKey,
		ProcessName:     p.ProcessName,
		ProcessVersion:  p.ProcessVersion,
		ProcessType:     p.ProcessType,
		ModelContent:    p.ModelContent,
		ProcessForm:     processForm,
		FormData:        string(formBytes),
		CurrentNodeKey:  curKey,
		CurrentNodeName: curName,
		InstanceState:   state,
		CreateID:        userID,
		CreateBy:        userName,
		CreateDeptID:    deptID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	engine.SnapshotProcessSetting(inst, p.ProcessSetting)

	sel := &engine.LaunchSelection{
		Assignees: map[string][]engine.UserRef{},
		CCUsers:   map[string][]engine.UserRef{},
	}
	for k, list := range req.NodeAssignees {
		for _, it := range list {
			sel.Assignees[k] = append(sel.Assignees[k], engine.UserRef{ID: it.ID, Name: it.Name})
		}
	}
	for k, list := range req.NodeCCUsers {
		for _, it := range list {
			sel.CCUsers[k] = append(sel.CCUsers[k], engine.UserRef{ID: it.ID, Name: it.Name})
		}
	}

	if req.SaveAsDraft {
		engine.SnapshotLaunchSelection(inst, sel)
		if err := s.repo.CreateInstance(ctx, inst); err != nil {
			return 0, err
		}
		return inst.ID, nil
	}

	if err := s.engine().Start(ctx, inst, sel); err != nil {
		return 0, err
	}
	if err := s.bootstrapSubProcesses(ctx, inst.ID, userID, userName); err != nil {
		return inst.ID, fmt.Errorf("流程已发起，但子流程启动失败: %w", err)
	}
	return inst.ID, nil
}

func logFlowSideEffect(msg string, err error) {
	if err == nil || logger.Logger == nil {
		return
	}
	logger.Logger.Error(msg, zap.Error(err))
}

// ActivateDraft 暂存单正式发起审批
func (s *FlowService) ActivateDraft(ctx context.Context, req *dto.ActivateDraftRequest, userID uint64) error {
	inst, err := s.repo.GetInstance(ctx, req.InstanceID)
	if err != nil || inst == nil {
		return ErrTaskNotFound
	}
	if inst.InstanceState != model.InstDraft {
		return fmt.Errorf("当前不是暂存单")
	}
	if inst.CreateID != userID {
		return fmt.Errorf("仅发起人可正式提交暂存单")
	}
	if inst.ParentInstanceID > 0 {
		return fmt.Errorf("子流程实例不可作为暂存发起")
	}

	sel := engine.LoadLaunchSelection(inst)
	if req.NodeAssignees != nil {
		sel.Assignees = map[string][]engine.UserRef{}
		for k, list := range req.NodeAssignees {
			for _, it := range list {
				sel.Assignees[k] = append(sel.Assignees[k], engine.UserRef{ID: it.ID, Name: it.Name})
			}
		}
	}
	if req.NodeCCUsers != nil {
		sel.CCUsers = map[string][]engine.UserRef{}
		for k, list := range req.NodeCCUsers {
			for _, it := range list {
				sel.CCUsers[k] = append(sel.CCUsers[k], engine.UserRef{ID: it.ID, Name: it.Name})
			}
		}
	}

	launchReq := &dto.LaunchRequest{
		ProcessID:     inst.ProcessID,
		FormData:      req.FormData,
		NodeAssignees: req.NodeAssignees,
		NodeCCUsers:   req.NodeCCUsers,
	}
	// 若请求未带自选，用暂存里的回填校验
	if launchReq.NodeAssignees == nil {
		launchReq.NodeAssignees = map[string][]dto.IDName{}
		for k, list := range sel.Assignees {
			for _, u := range list {
				launchReq.NodeAssignees[k] = append(launchReq.NodeAssignees[k], dto.IDName{ID: u.ID, Name: u.Name})
			}
		}
	}
	if launchReq.NodeCCUsers == nil {
		launchReq.NodeCCUsers = map[string][]dto.IDName{}
		for k, list := range sel.CCUsers {
			for _, u := range list {
				launchReq.NodeCCUsers[k] = append(launchReq.NodeCCUsers[k], dto.IDName{ID: u.ID, Name: u.Name})
			}
		}
	}

	if mc, e := engine.ParseModel(inst.ModelContent); e == nil {
		for _, n := range collectSelectionNodes(mc.NodeConfig) {
			if !n.Required {
				continue
			}
			list := launchReq.NodeAssignees[n.NodeKey]
			if n.Kind == "cc" {
				list = launchReq.NodeCCUsers[n.NodeKey]
			}
			if len(list) == 0 {
				label := "审批人"
				if n.Kind == "cc" {
					label = "抄送人"
				}
				return fmt.Errorf("请为节点「%s」选择%s", n.NodeName, label)
			}
		}
		if err := validateLaunchCandidates(mc.NodeConfig, launchReq); err != nil {
			return err
		}
	}

	schema := s.resolveInstanceFormSchema(ctx, inst)
	if mc, e := engine.ParseModel(inst.ModelContent); e == nil && mc != nil && mc.NodeConfig != nil {
		schema = s.mergeActionURLForm(ctx, schema, mc.NodeConfig.ActionURL)
	}
	cfg := startFormConfig(inst.ModelContent)
	cfg = appendMissingFormConfig(cfg, schema, 1)
	formData := req.FormData
	if formData == nil {
		formData = unmarshalMap(inst.FormData)
	} else {
		formData = applyFormConfigWrite(unmarshalMap(inst.FormData), formData, cfg)
	}
	if err := validateRequiredFields(schema, formData, cfg, true); err != nil {
		return err
	}
	b, _ := json.Marshal(formData)
	inst.FormData = string(b)
	inst.UpdatedAt = time.Now()
	engine.SnapshotLaunchSelection(inst, sel)
	if err := s.repo.UpdateInstance(ctx, inst); err != nil {
		return err
	}
	if err := s.engine().ContinueFromStart(ctx, inst, sel); err != nil {
		return err
	}
	if err := s.bootstrapSubProcesses(ctx, inst.ID, inst.CreateID, inst.CreateBy); err != nil {
		return err
	}
	_, _ = s.TickDelays(ctx)
	return nil
}

// DeleteDraft 删除暂存单
func (s *FlowService) DeleteDraft(ctx context.Context, instanceID, userID uint64) error {
	inst, err := s.repo.GetInstance(ctx, instanceID)
	if err != nil || inst == nil {
		return ErrTaskNotFound
	}
	if inst.InstanceState != model.InstDraft {
		return fmt.Errorf("只能删除暂存单")
	}
	if inst.CreateID != userID {
		return fmt.Errorf("仅发起人可删除暂存单")
	}
	return s.repo.DeleteInstanceHard(ctx, instanceID)
}

// UpdateDraft 更新暂存单（保持暂存状态）
func (s *FlowService) UpdateDraft(ctx context.Context, req *dto.UpdateDraftRequest, userID uint64) error {
	inst, err := s.repo.GetInstance(ctx, req.InstanceID)
	if err != nil || inst == nil {
		return ErrTaskNotFound
	}
	if inst.InstanceState != model.InstDraft {
		return fmt.Errorf("当前不是暂存单")
	}
	if inst.CreateID != userID {
		return fmt.Errorf("仅发起人可更新暂存单")
	}
	if inst.ParentInstanceID > 0 {
		return fmt.Errorf("子流程实例不可作为暂存")
	}

	sel := engine.LoadLaunchSelection(inst)
	if req.NodeAssignees != nil {
		sel.Assignees = map[string][]engine.UserRef{}
		for k, list := range req.NodeAssignees {
			for _, it := range list {
				sel.Assignees[k] = append(sel.Assignees[k], engine.UserRef{ID: it.ID, Name: it.Name})
			}
		}
	}
	if req.NodeCCUsers != nil {
		sel.CCUsers = map[string][]engine.UserRef{}
		for k, list := range req.NodeCCUsers {
			for _, it := range list {
				sel.CCUsers[k] = append(sel.CCUsers[k], engine.UserRef{ID: it.ID, Name: it.Name})
			}
		}
	}

	schema := s.resolveInstanceFormSchema(ctx, inst)
	if mc, e := engine.ParseModel(inst.ModelContent); e == nil && mc != nil && mc.NodeConfig != nil {
		schema = s.mergeActionURLForm(ctx, schema, mc.NodeConfig.ActionURL)
	}
	cfg := startFormConfig(inst.ModelContent)
	cfg = appendMissingFormConfig(cfg, schema, 1)
	formData := req.FormData
	if formData == nil {
		formData = unmarshalMap(inst.FormData)
	} else {
		formData = applyFormConfigWrite(unmarshalMap(inst.FormData), formData, cfg)
	}
	// 暂存不校验必填
	b, _ := json.Marshal(formData)
	inst.FormData = string(b)
	inst.UpdatedAt = time.Now()
	engine.SnapshotLaunchSelection(inst, sel)
	return s.repo.UpdateInstance(ctx, inst)
}

func (s *FlowService) bootstrapSubProcesses(ctx context.Context, parentID uint64, userID uint64, userName string) error {
	tasks, err := s.repo.ListActiveTasksByInstance(ctx, parentID)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if t.TaskType != model.TaskSub {
			continue
		}
		var payload struct {
			CallProcess     string `json:"callProcess"`
			ChildInstanceID uint64 `json:"childInstanceId"`
		}
		_ = json.Unmarshal([]byte(t.Payload), &payload)
		// 幂等：已绑定子实例则跳过
		if payload.ChildInstanceID > 0 {
			continue
		}
		// 幂等：该节点仍有活动子实例
		if n, _ := s.repo.CountActiveChildren(ctx, parentID, t.NodeKey); n > 0 {
			continue
		}
		// 幂等：本 TaskSub 生命周期内已创建过子实例（含已结束）
		if latest, e := s.repo.GetLatestChildByParentNode(ctx, parentID, t.NodeKey); e == nil && latest != nil {
			if !latest.CreatedAt.Before(t.CreatedAt.Add(-time.Second)) {
				var raw map[string]interface{}
				_ = json.Unmarshal([]byte(t.Payload), &raw)
				if raw == nil {
					raw = map[string]interface{}{}
				}
				raw["childInstanceId"] = latest.ID
				b, _ := json.Marshal(raw)
				t.Payload = string(b)
				t.UpdatedAt = time.Now()
				_ = s.repo.UpdateTask(ctx, &t)
				continue
			}
		}

		ref := strings.TrimSpace(payload.CallProcess)
		if ref == "" {
			continue
		}
		// 支持 id 或 id:name
		idStr := ref
		if i := strings.Index(ref, ":"); i > 0 {
			idStr = ref[:i]
		}
		var pid uint64
		fmt.Sscanf(idStr, "%d", &pid)
		var childProc *model.FlowProcess
		if pid > 0 {
			childProc, err = s.repo.GetProcess(ctx, pid)
		} else {
			childProc, err = s.repo.GetProcessByKey(ctx, ref)
		}
		if err != nil || childProc == nil || childProc.ProcessState != 1 {
			return fmt.Errorf("子流程不可用: %s", ref)
		}
		parent, _ := s.repo.GetInstance(ctx, parentID)
		// 禁止自调用 / A→B→A：祖先链上已出现同一 processId 则形成环
		if err := s.ensureSubProcessNoCycle(ctx, parent, childProc.ID); err != nil {
			return err
		}
		formData := "{}"
		// 子实例发起人固定为父实例发起人（勿用当前审批人，否则部门负责人/发起人条件/撤销归属都会错）
		starterID := userID
		starterName := userName
		if parent != nil {
			if parent.CreateID > 0 {
				starterID = parent.CreateID
			}
			if parent.CreateBy != "" {
				starterName = parent.CreateBy
			}
		}
		// 子流程表单：优先子定义（含业务绑定表）；无字段时继承父实例快照
		processForm := s.resolveProcessFormJSON(ctx, childProc)
		if !formMapHasWidgets(unmarshalMap(processForm)) && parent != nil {
			if formMapHasWidgets(unmarshalMap(parent.ProcessForm)) {
				processForm = parent.ProcessForm
			} else if pp, e := s.repo.GetProcess(ctx, parent.ProcessID); e == nil && pp != nil {
				processForm = s.resolveProcessFormJSON(ctx, pp)
			}
		}
		// 仅下发子表单 schema 中存在的字段，避免把父单无关字段整包带进子实例
		if parent != nil {
			filtered := filterFormDataBySchema(unmarshalMap(parent.FormData), unmarshalMap(processForm))
			if len(filtered) == 0 {
				formData = "{}"
			} else if b, e := json.Marshal(filtered); e == nil {
				formData = string(b)
			}
		}
		child := &model.FlowInstance{
			ProcessID:        childProc.ID,
			ProcessKey:       childProc.ProcessKey,
			ProcessName:      childProc.ProcessName,
			ProcessVersion:   childProc.ProcessVersion,
			ProcessType:      childProc.ProcessType,
			ModelContent:     childProc.ModelContent,
			ProcessForm:      processForm,
			FormData:         formData,
			InstanceState:    model.InstActive,
			CreateID:         starterID,
			CreateBy:         starterName,
			CreateDeptID:     0,
			ParentInstanceID: parentID,
			ParentNodeKey:    t.NodeKey,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if parent != nil {
			child.CreateDeptID = parent.CreateDeptID
		}
		engine.SnapshotProcessSetting(child, childProc.ProcessSetting)
		if err := s.engine().Start(ctx, child, &engine.LaunchSelection{}); err != nil {
			return err
		}
		// 回写 childInstanceId，保留原 payload（含 gateStack）
		var raw map[string]interface{}
		_ = json.Unmarshal([]byte(t.Payload), &raw)
		if raw == nil {
			raw = map[string]interface{}{}
		}
		raw["callProcess"] = payload.CallProcess
		raw["childInstanceId"] = child.ID
		b, _ := json.Marshal(raw)
		t.Payload = string(b)
		t.UpdatedAt = time.Now()
		_ = s.repo.UpdateTask(ctx, &t)

		if err := s.bootstrapSubProcesses(ctx, child.ID, child.CreateID, child.CreateBy); err != nil {
			return fmt.Errorf("子流程实例 %d 启动嵌套子流程失败: %w", child.ID, err)
		}
	}
	return nil
}

// ensureSubProcessNoCycle 校验即将启动的子流程不会与祖先实例形成 processId 环（含直接选自己）
func (s *FlowService) ensureSubProcessNoCycle(ctx context.Context, parent *model.FlowInstance, childProcessID uint64) error {
	if childProcessID == 0 {
		return fmt.Errorf("子流程不可用")
	}
	const maxDepth = 16
	cur := parent
	for depth := 0; depth < maxDepth && cur != nil; depth++ {
		if cur.ProcessID == childProcessID {
			if depth == 0 {
				return fmt.Errorf("子流程不能调用自身，请改选其他子流程")
			}
			return fmt.Errorf("子流程调用形成循环（已出现在上级实例链），请检查子流程互相引用")
		}
		if cur.ParentInstanceID == 0 {
			break
		}
		next, err := s.repo.GetInstance(ctx, cur.ParentInstanceID)
		if err != nil || next == nil {
			break
		}
		cur = next
	}
	return nil
}

// checkParentResume 子流程结束后回写父流程：仅成功才继续；失败则终止父流程在子流程节点
func (s *FlowService) checkParentResume(ctx context.Context, inst *model.FlowInstance) {
	if inst == nil || inst.ID == 0 {
		return
	}
	fresh, err := s.repo.GetInstance(ctx, inst.ID)
	if err != nil || fresh == nil {
		logFlowSideEffect("checkParentResume load child instance", err)
		return
	}
	if fresh.ParentInstanceID == 0 || fresh.InstanceState == model.InstActive {
		return
	}
	n, err := s.repo.CountActiveChildren(ctx, fresh.ParentInstanceID, fresh.ParentNodeKey)
	if err != nil {
		logFlowSideEffect("checkParentResume count active children", err)
		return
	}
	if n > 0 {
		return
	}
	parentID := fresh.ParentInstanceID
	switch fresh.InstanceState {
	case model.InstComplete:
		parent, pe := s.repo.GetInstance(ctx, parentID)
		if pe != nil || parent == nil {
			logFlowSideEffect("checkParentResume load parent instance", pe)
			return
		}
		if mergeChildFormIntoParent(parent, fresh) {
			parent.UpdatedAt = time.Now()
			if ue := s.repo.UpdateInstance(ctx, parent); ue != nil {
				logFlowSideEffect("checkParentResume merge child form into parent", ue)
			}
		}
		if err := s.engine().ResumeAfterSubProcess(ctx, parentID, fresh.ParentNodeKey); err != nil {
			logFlowSideEffect("ResumeAfterSubProcess", err)
			return
		}
		parent, pe = s.repo.GetInstance(ctx, parentID)
		if pe != nil || parent == nil {
			logFlowSideEffect("checkParentResume reload parent after resume", pe)
			return
		}
		if err := s.bootstrapSubProcesses(ctx, parentID, parent.CreateID, parent.CreateBy); err != nil {
			logFlowSideEffect("bootstrapSubProcesses after sub-process complete", err)
		}
		// 父流程可能也是子流程：成功同样向上冒泡（与失败路径一致）
		s.checkParentResume(ctx, parent)
	case model.InstReject, model.InstRevoke, model.InstTimeout, model.InstTerminate:
		opinion := "子流程未通过"
		switch fresh.InstanceState {
		case model.InstRevoke:
			opinion = "子流程已撤销"
		case model.InstTimeout:
			opinion = "子流程已超时"
		case model.InstTerminate:
			opinion = "子流程已终止"
		case model.InstReject:
			opinion = "子流程已拒绝"
		}
		if err := s.engine().FailAfterSubProcess(ctx, parentID, fresh.ParentNodeKey, fresh.InstanceState, opinion); err != nil {
			logFlowSideEffect("FailAfterSubProcess", err)
			return
		}
		// 父流程可能也是子流程：继续向上传递失败
		if parent, pe := s.repo.GetInstance(ctx, parentID); pe == nil && parent != nil {
			s.checkParentResume(ctx, parent)
		}
	}
}

func (s *FlowService) ListPending(ctx context.Context, userID uint64, f repository.ApproveListFilter, page, pageSize int) ([]dto.PendingTaskItem, int64, error) {
	principals := s.principalsOf(ctx, userID)
	rows, total, err := s.repo.ListPendingTasks(ctx, userID, principals, f, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PendingTaskItem, 0, len(rows))
	for _, r := range rows {
		isDelegate := r.ActorUserID != userID
		instStub := &model.FlowInstance{Variable: r.Variable}
		allowDelegate := engine.ProcessSettingBool(instStub, "allowDelegate", true)
		item := dto.PendingTaskItem{
			ActorID:           r.ActorID,
			ActorUserID:       r.ActorUserID,
			TaskID:            r.TaskID,
			InstanceID:        r.InstanceID,
			ParentInstanceID:  r.ParentInstanceID,
			NodeName:          r.NodeName,
			WorkNodeName:      r.WorkNodeName,
			NodeKey:           r.NodeKey,
			ExamineMode:       r.ExamineMode,
			ProcessName:       r.ProcessName,
			SubProcessName:    r.SubProcessName,
			IsSubProcess:      r.IsSubProcess,
			CreateBy:          r.CreateBy,
			CreateID:          r.CreateID,
			CreatedAt:         r.CreatedAt,
			CanHandle:         true,
			AllowBatchOperate: engine.ProcessSettingBool(instStub, "allowBatchOperate", true),
			IsDelegate:        isDelegate,
		}
		task, err := s.repo.GetTask(ctx, r.TaskID)
		if err == nil {
			actors, _ := s.repo.ListTaskActors(ctx, r.TaskID)
			ps := principals
			if !allowDelegate {
				ps = nil
			}
			item.CanHandle = engine.CanUserAct(task, actors, userID, ps...)
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *FlowService) ListMyApplications(ctx context.Context, userID uint64, f repository.ApproveListFilter, page, pageSize int) ([]dto.InstanceListItem, int64, error) {
	list, total, err := s.repo.ListMyInstances(ctx, userID, f, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.InstanceListItem, 0, len(list))
	for _, inst := range list {
		item := dto.InstanceListItem{
			InstanceID:      inst.ID,
			ProcessName:     inst.ProcessName,
			ProcessKey:      inst.ProcessKey,
			CurrentNodeName: inst.CurrentNodeName,
			InstanceState:   inst.InstanceState,
			CreateBy:        inst.CreateBy,
			CreatedAt:       inst.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if inst.FinishTime != nil {
			item.FinishTime = inst.FinishTime.Format("2006-01-02 15:04:05")
		}
		out = append(out, item)
	}
	return out, total, nil
}

// ListMonitor 流程监控：用户可管理的流程下的主实例（默认仅审批中）
func (s *FlowService) ListMonitor(ctx context.Context, userID uint64, f repository.ApproveListFilter, onlyActive bool, page, pageSize int) ([]dto.InstanceListItem, int64, error) {
	ids, all, err := s.listManagedProcessIDs(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	list, total, err := s.repo.ListMonitorInstances(ctx, ids, all, f, onlyActive, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.InstanceListItem, 0, len(list))
	for _, inst := range list {
		item := dto.InstanceListItem{
			InstanceID:      inst.ID,
			ProcessName:     inst.ProcessName,
			ProcessKey:      inst.ProcessKey,
			CurrentNodeName: inst.CurrentNodeName,
			InstanceState:   inst.InstanceState,
			CreateBy:        inst.CreateBy,
			CreatedAt:       inst.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if inst.FinishTime != nil {
			item.FinishTime = inst.FinishTime.Format("2006-01-02 15:04:05")
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *FlowService) ListApproved(ctx context.Context, userID uint64, f repository.ApproveListFilter, page, pageSize int) ([]dto.InstanceListItem, int64, error) {
	rows, total, err := s.repo.ListApprovedByUser(ctx, userID, f, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.InstanceListItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.InstanceListItem{
			ID:             r.ActorRowID,
			HisTaskID:      r.HisTaskID,
			InstanceID:     r.InstanceID,
			ProcessName:    r.ProcessName,
			SubProcessName: r.SubProcessName,
			IsSubProcess:   r.IsSubProcess,
			ProcessKey:     r.ProcessKey,
			NodeName:       r.NodeName,
			TaskID:         r.TaskID,
			ActorState:     r.ActorState,
			InstanceState:  r.InstanceState,
			CreateBy:       r.CreateBy,
			CreatedAt:      r.CreatedAt,
			FinishTime:     r.FinishTime,
			IsDelegate:     r.IsDelegate,
		})
	}
	return out, total, nil
}

func (s *FlowService) ListReceived(ctx context.Context, userID uint64, f repository.ApproveListFilter, page, pageSize int) ([]dto.InstanceListItem, int64, error) {
	rows, total, err := s.repo.ListReceivedCC(ctx, userID, f, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.InstanceListItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.InstanceListItem{
			ID:             r.ActorRowID,
			HisTaskID:      r.HisTaskID,
			InstanceID:     r.InstanceID,
			ProcessName:    r.ProcessName,
			SubProcessName: r.SubProcessName,
			IsSubProcess:   r.IsSubProcess,
			ProcessKey:     r.ProcessKey,
			NodeName:       r.NodeName,
			TaskID:         r.TaskID,
			ActorState:     r.ActorState,
			InstanceState:  r.InstanceState,
			CreateBy:       r.CreateBy,
			CreatedAt:      r.CreatedAt,
			FinishTime:     r.FinishTime,
			Read:           r.Read,
		})
	}
	return out, total, nil
}

// MarkReceivedRead 打开「我收到的」详情时标记抄送已读
func (s *FlowService) MarkReceivedRead(ctx context.Context, actorRowID, userID uint64) error {
	if actorRowID == 0 {
		return fmt.Errorf("无效抄送记录")
	}
	return s.repo.MarkReceivedCCRead(ctx, actorRowID, userID)
}

// ListClaimable 认领任务列表
func (s *FlowService) ListClaimable(ctx context.Context, userID uint64, f repository.ApproveListFilter, page, pageSize int) ([]dto.InstanceListItem, int64, error) {
	roleIDs, err := s.org().UserRoleIDs(userID)
	if err != nil {
		return nil, 0, err
	}
	rows, total, err := s.repo.ListClaimableByRoles(ctx, roleIDs, f, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.InstanceListItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.InstanceListItem{
			ID:             r.ActorRowID,
			InstanceID:     r.InstanceID,
			ProcessName:    r.ProcessName,
			SubProcessName: r.SubProcessName,
			IsSubProcess:   r.IsSubProcess,
			ProcessKey:     r.ProcessKey,
			NodeName:       r.NodeName,
			TaskID:         r.TaskID,
			ActorState:     r.ActorState,
			InstanceState:  r.InstanceState,
			CreateBy:       r.CreateBy,
			CreatedAt:      r.CreatedAt,
			FinishTime:     r.FinishTime,
		})
	}
	return out, total, nil
}

func (s *FlowService) Claim(ctx context.Context, taskID, userID uint64) error {
	return s.engine().Claim(ctx, taskID, userID)
}

func (s *FlowService) Consent(ctx context.Context, req *dto.ConsentRequest, userID uint64) error {
	task, err := s.repo.GetTask(ctx, req.TaskID)
	if err != nil {
		return ErrTaskNotFound
	}
	if task.NodeType == engine.NodeStart {
		return fmt.Errorf("请使用重新提交")
	}
	inst, _ := s.repo.GetInstance(ctx, task.InstanceID)
	actors, _ := s.repo.ListTaskActors(ctx, req.TaskID)
	ps := s.actPrincipals(ctx, inst, userID)
	if !engine.CanUserAct(task, actors, userID, ps...) {
		return ErrCannotHandle
	}
	var formData map[string]interface{}
	if inst != nil {
		if req.FormData != nil {
			formData = s.sanitizeInstanceForm(ctx, inst, task.NodeKey, req.FormData, false)
			req.FormData = formData
		} else {
			formData = unmarshalMap(inst.FormData)
		}
		if err := s.validateInstanceNodeForm(ctx, inst, task.NodeKey, formData, false); err != nil {
			return err
		}
	}
	if err := s.engine().Consent(ctx, req.TaskID, userID, req.Opinion, req.FormData); err != nil {
		return err
	}
	if inst != nil {
		s.checkParentResume(ctx, inst)
		if err := s.bootstrapSubProcesses(ctx, inst.ID, inst.CreateID, inst.CreateBy); err != nil {
			logFlowSideEffect("bootstrapSubProcesses after consent", err)
		}
	}
	_, _ = s.TickDelays(ctx)
	return nil
}

func (s *FlowService) Resubmit(ctx context.Context, req *dto.ResubmitRequest, userID uint64) error {
	task, err := s.repo.GetTask(ctx, req.TaskID)
	if err != nil {
		return ErrTaskNotFound
	}
	actors, _ := s.repo.ListTaskActors(ctx, req.TaskID)
	if !engine.CanUserAct(task, actors, userID) {
		return ErrCannotHandle
	}
	inst, _ := s.repo.GetInstance(ctx, task.InstanceID)
	var formData map[string]interface{}
	if inst != nil {
		if req.FormData != nil {
			// 发起人改单：无 formConfig 时允许写全表
			formData = s.sanitizeInstanceForm(ctx, inst, task.NodeKey, req.FormData, true)
			req.FormData = formData
		} else {
			formData = unmarshalMap(inst.FormData)
		}
		if err := s.validateInstanceNodeForm(ctx, inst, task.NodeKey, formData, true); err != nil {
			return err
		}
	}
	if err := s.engine().Resubmit(ctx, req.TaskID, userID, req.FormData); err != nil {
		return err
	}
	if inst == nil {
		inst, _ = s.repo.GetInstance(ctx, task.InstanceID)
	}
	if inst != nil {
		s.checkParentResume(ctx, inst)
		if err := s.bootstrapSubProcesses(ctx, inst.ID, inst.CreateID, inst.CreateBy); err != nil {
			logFlowSideEffect("bootstrapSubProcesses after resubmit", err)
		}
	}
	_, _ = s.TickDelays(ctx)
	return nil
}

func (s *FlowService) Reject(ctx context.Context, req *dto.RejectRequest, userID uint64) error {
	task, err := s.repo.GetTask(ctx, req.TaskID)
	if err != nil {
		return ErrTaskNotFound
	}
	inst, _ := s.repo.GetInstance(ctx, task.InstanceID)
	actors, _ := s.repo.ListTaskActors(ctx, req.TaskID)
	ps := s.actPrincipals(ctx, inst, userID)
	if !engine.CanUserAct(task, actors, userID, ps...) {
		return ErrCannotHandle
	}
	if req.FormData != nil && inst != nil {
		req.FormData = s.sanitizeInstanceForm(ctx, inst, task.NodeKey, req.FormData, false)
	}
	if err := s.engine().Reject(ctx, req.TaskID, userID, req.Opinion, req.FormData, req.RejectStrategy, req.RejectNodeKey); err != nil {
		return err
	}
	if inst != nil {
		s.checkParentResume(ctx, inst)
	}
	_, _ = s.TickDelays(ctx)
	return nil
}

func (s *FlowService) GetInstanceDetail(ctx context.Context, instanceID, userID uint64, taskID uint64) (*dto.InstanceDetail, error) {
	workInst, err := s.repo.GetInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if !s.canAccessInstance(ctx, workInst, userID) {
		return nil, fmt.Errorf("无权查看该流程")
	}
	viewInst := workInst
	mainView := false
	// 子流程待办（带 taskId）：详情用主流程视角（流转/流程图），办理仍落在子任务
	if workInst.ParentInstanceID > 0 && taskID > 0 {
		if task, e := s.repo.GetTask(ctx, taskID); e == nil && task.InstanceID == workInst.ID {
			if root := s.resolveRootInstance(ctx, workInst); root != nil && root.ID != workInst.ID {
				viewInst = root
				mainView = true
			}
		}
	}
	inst := viewInst
	viewID := viewInst.ID
	workID := workInst.ID

	formData := unmarshalMap(workInst.FormData)
	if len(formData) == 0 {
		formData = unmarshalMap(viewInst.FormData)
	}
	detail := &dto.InstanceDetail{
		InstanceID:          viewID,
		ProcessID:           viewInst.ProcessID,
		ProcessName:         viewInst.ProcessName,
		ProcessKey:          viewInst.ProcessKey,
		ProcessVersion:      viewInst.ProcessVersion,
		InstanceState:       viewInst.InstanceState,
		CurrentNodeKey:      viewInst.CurrentNodeKey,
		CurrentNodeName:     viewInst.CurrentNodeName,
		CreateID:            viewInst.CreateID,
		CreateBy:            viewInst.CreateBy,
		CreatedAt:           viewInst.CreatedAt.Format("2006-01-02 15:04:05"),
		FormSchema:          nil, // 下方按办理节点合并子表单后再赋值
		FormData:            formData,
		FormRenderType:      "designer",
		ModelContent:        unmarshalMap(viewInst.ModelContent),
		SecondOperatePrompt: engine.ProcessSettingBool(viewInst, "secondOperatePrompt", true),
		AllowDelegate:       engine.ProcessSettingBool(workInst, "allowDelegate", true),
	}
	if p, e := s.repo.GetProcess(ctx, workInst.ProcessID); e == nil && p != nil {
		rt, comp := s.resolveBoundFormRender(ctx, p)
		detail.FormRenderType = rt
		detail.FormComponent = comp
	}
	if mainView {
		detail.IsSubProcess = true
		detail.SubProcessName = workInst.ProcessName
		detail.WorkInstanceID = workID
		detail.ParentInstanceID = viewID
		if detail.SubProcessName == "" {
			detail.SubProcessName = "子流程"
		}
	} else {
		s.applyRootProcessDisplay(ctx, workInst, &detail.ProcessName, &detail.SubProcessName, &detail.IsSubProcess)
		if workInst.ParentInstanceID > 0 {
			detail.WorkInstanceID = workID
			detail.ParentInstanceID = workInst.ParentInstanceID
		}
	}
	if viewInst.FinishTime != nil {
		detail.FinishTime = viewInst.FinishTime.Format("2006-01-02 15:04:05")
	}

	// 自愈：子流程占位任务若尚未创建子实例，打开详情时补启动（避免一直「等待子流程」却无法查看）
	if workInst.InstanceState == model.InstActive {
		if err := s.bootstrapSubProcesses(ctx, workID, workInst.CreateID, workInst.CreateBy); err != nil {
			logFlowSideEffect("bootstrapSubProcesses on detail", err)
		}
	}

	mc, _ := engine.ParseModel(viewInst.ModelContent)
	workMc, _ := engine.ParseModel(workInst.ModelContent)
	if workMc == nil {
		workMc = mc
	}
	var curTaskNodeKey string
	if taskID > 0 {
		task, err := s.repo.GetTask(ctx, taskID)
		if err == nil && task.InstanceID == workID {
			detail.TaskID = task.ID
			detail.TaskNodeType = task.NodeType
			curTaskNodeKey = task.NodeKey
			detail.WorkNodeKey = task.NodeKey
			detail.WorkNodeName = task.NodeName
			actors, _ := s.repo.ListTaskActors(ctx, task.ID)
			ps := s.actPrincipals(ctx, workInst, userID)
			can := engine.CanUserAct(task, actors, userID, ps...)
			if task.NodeType == engine.NodeStart {
				// 发起人改单重提：可编辑表单并重新提交，不可同意/驳回/转交等
				detail.CanResubmit = can && workInst.CreateID == userID
				detail.CanConsent = false
				detail.CanReject = false
				if workMc != nil && workMc.NodeConfig != nil {
					detail.FormConfig = formConfigOf(workMc.NodeConfig)
				}
			} else {
				detail.CanConsent = can
				detail.CanReject = can
				if workMc != nil {
					if n := engine.FindNode(workMc.NodeConfig, task.NodeKey); n != nil {
						detail.RejectStrategy = n.RejectStrategy
						if detail.RejectStrategy == 0 {
							detail.RejectStrategy = 3 // 与引擎默认「指定节点」一致
						}
						detail.RejectStart = n.RejectStart
						detail.AllowRollback = n.AllowRollback
						detail.AllowTransfer = n.AllowTransfer
						detail.AllowAppendNode = n.AllowAppendNode
						detail.CanTransfer = can && n.AllowTransfer
						detail.CanAppend = can && n.AllowAppendNode && n.SetType != 6
						detail.CanRollback = can && n.AllowRollback
						detail.AllowCc = n.AllowCc
						detail.CanCc = can && n.AllowCc
						detail.FormConfig = formConfigOf(n)
						if n.ExtendConfig != nil {
							if v, ok := n.ExtendConfig["rejectNodeKey"].(string); ok {
								detail.RejectNodeKey = v
							}
						}
					}
				}
			}
			for _, a := range actors {
				detail.TaskActors = append(detail.TaskActors, dto.TaskActorItem{
					ID: a.ID, ActorID: a.ActorID, ActorName: a.ActorName,
					ActorType: a.ActorType, ActorState: a.ActorState, Weight: a.Weight,
				})
			}
			if detail.CanAppend {
				pending := 0
				removable := 0
				for _, a := range actors {
					if a.ActorState != model.ActorPending {
						continue
					}
					pending++
					if a.ActorType == model.ActorTypeAppend && a.ActorID != userID {
						removable++
					}
				}
				detail.CanRemove = pending > 1 && removable > 0
			}
		}
	} else if workMc != nil {
		// 无 taskId：发起人打开「我的申请」时，自动挂上改单重提待办
		if workInst.CreateID == userID && workInst.InstanceState == model.InstActive {
			actives, _ := s.repo.ListActiveTasksByInstance(ctx, workID)
			for _, t := range actives {
				if t.NodeType != engine.NodeStart {
					continue
				}
				actors, _ := s.repo.ListTaskActors(ctx, t.ID)
				if !engine.CanUserAct(&t, actors, userID) {
					continue
				}
				detail.TaskID = t.ID
				detail.TaskNodeType = t.NodeType
				curTaskNodeKey = t.NodeKey
				detail.WorkNodeKey = t.NodeKey
				detail.WorkNodeName = t.NodeName
				detail.CanResubmit = true
				detail.FormConfig = formConfigOf(workMc.NodeConfig)
				for _, a := range actors {
					detail.TaskActors = append(detail.TaskActors, dto.TaskActorItem{
						ID: a.ID, ActorID: a.ActorID, ActorName: a.ActorName,
						ActorType: a.ActorType, ActorState: a.ActorState, Weight: a.Weight,
					})
				}
				break
			}
		}
		if detail.FormConfig == nil {
			if n := engine.FindNode(workMc.NodeConfig, workInst.CurrentNodeKey); n != nil {
				detail.FormConfig = formConfigOf(n)
				detail.RejectStrategy = n.RejectStrategy
				if detail.RejectStrategy == 0 {
					detail.RejectStrategy = 3
				}
			} else if workMc.NodeConfig != nil {
				detail.FormConfig = formConfigOf(workMc.NodeConfig)
			}
		}
	}

	// 暂存待审：发起人可编辑表单并正式发起 / 删除
	if workInst.InstanceState == model.InstDraft && workInst.CreateID == userID {
		detail.CanActivateDraft = true
		detail.CanUpdateDraft = true
		detail.CanDeleteDraft = true
		if workMc != nil && workMc.NodeConfig != nil {
			detail.FormConfig = formConfigOf(workMc.NodeConfig)
			detail.WorkNodeKey = workMc.NodeConfig.NodeKey
			detail.WorkNodeName = workMc.NodeConfig.NodeName
			if detail.WorkNodeName == "" {
				detail.WorkNodeName = "发起人"
			}
			detail.SelectionNodes = collectSelectionNodes(workMc.NodeConfig)
		}
		sel := engine.LoadLaunchSelection(workInst)
		payload := &dto.LaunchSelectionPayload{
			Assignees: map[string][]dto.IDName{},
			CCUsers:   map[string][]dto.IDName{},
		}
		if sel != nil {
			for k, list := range sel.Assignees {
				for _, u := range list {
					payload.Assignees[k] = append(payload.Assignees[k], dto.IDName{ID: u.ID, Name: u.Name})
				}
			}
			for k, list := range sel.CCUsers {
				for _, u := range list {
					payload.CCUsers[k] = append(payload.CCUsers[k], dto.IDName{ID: u.ID, Name: u.Name})
				}
			}
		}
		detail.LaunchSelection = payload
	}

	// 表单 schema / 权限
	{
		var formNode *engine.Node
		var root *engine.Node
		if workMc != nil {
			root = workMc.NodeConfig
			key := curTaskNodeKey
			if key == "" {
				key = workInst.CurrentNodeKey
			}
			if key != "" {
				formNode = engine.FindNode(root, key)
			}
			if formNode == nil {
				formNode = root
			}
		}
		if workInst.InstanceState == model.InstDraft {
			// 暂存与发起页一致：仅主表单 + 发起节点子表单，不拼其它审批节点补录字段
			detail.FormSchema = s.resolveNodeFormSchema(ctx, workInst, root)
			if root != nil && strings.TrimSpace(root.ActionURL) != "" {
				detail.FormConfig = appendMissingFormConfig(
					detail.FormConfig,
					s.mergeActionURLForm(ctx, map[string]interface{}{}, root.ActionURL),
					1,
				)
			}
			detail.FormConfig = appendMissingFormConfig(detail.FormConfig, detail.FormSchema, 1)
		} else {
			// 主表单 + 历史已办子表；当前节点子表仅在可办理编辑时拼入（认领/只读不提前露空字段）
			includeCurrent := detail.CanConsent || detail.CanResubmit
			reached := s.collectDisplayFormNodeKeys(ctx, workInst, root, curTaskNodeKey, includeCurrent)
			detail.FormSchema = s.resolveDisplayFormSchema(ctx, workInst, root, reached)
			if includeCurrent && formNode != nil && strings.TrimSpace(formNode.ActionURL) != "" {
				detail.FormConfig = appendMissingFormConfig(
					detail.FormConfig,
					s.mergeActionURLForm(ctx, map[string]interface{}{}, formNode.ActionURL),
					1,
				)
			}
			detail.FormConfig = appendMissingFormConfig(detail.FormConfig, detail.FormSchema, 0)
		}
		detail.FormData = applyFormConfigRead(detail.FormData, detail.FormConfig)
	}

	// 流程图办理人：无 taskId（如发起人看「我的申请」）时也挂上当前活动节点参与人，否则加签不可见
	if len(detail.TaskActors) == 0 {
		actives, _ := s.repo.ListActiveTasksByInstance(ctx, viewID)
		for _, t := range actives {
			if inst.CurrentNodeKey != "" && t.NodeKey != inst.CurrentNodeKey {
				continue
			}
			actors, _ := s.repo.ListTaskActors(ctx, t.ID)
			for _, a := range actors {
				detail.TaskActors = append(detail.TaskActors, dto.TaskActorItem{
					ID: a.ID, ActorID: a.ActorID, ActorName: a.ActorName,
					ActorType: a.ActorType, ActorState: a.ActorState, Weight: a.Weight,
				})
			}
			if len(detail.TaskActors) > 0 {
				break
			}
		}
	}

	his, _ := s.repo.ListHisTasks(ctx, viewID)
	hisActors, _ := s.repo.ListHisActorsByInstance(ctx, viewID)
	// 必须按 his_task_id 归组。按 task_id 会把多次自动通过（task_id 常为 0）等历史串在一起，回退后流转记录爆炸
	actorMap := map[uint64][]model.FlowHisTaskActor{}
	for _, a := range hisActors {
		actorMap[a.HisTaskID] = append(actorMap[a.HisTaskID], a)
	}

	// 发起节点本身不落任务，时间线最前补一条
	startName := "发起人"
	startKey := ""
	if mc != nil && mc.NodeConfig != nil {
		if mc.NodeConfig.NodeName != "" {
			startName = mc.NodeConfig.NodeName
		}
		startKey = mc.NodeConfig.NodeKey
	}
	startItem := dto.TimelineItem{
		NodeName:   startName,
		NodeKey:    startKey,
		NodeType:   engine.NodeStart,
		TaskState:  model.TaskComplete,
		ActorName:  inst.CreateBy,
		ActorState: model.ActorAgree,
		FinishTime: inst.CreatedAt.Format("2006-01-02 15:04:05"),
		CreatedAt:  inst.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if viewInst.InstanceState == model.InstDraft {
		startItem.NodeName = startName + "（暂存）"
		startItem.TaskState = model.TaskActive
		startItem.ActorState = model.ActorPending
		startItem.FinishTime = ""
		startItem.Opinion = "暂存待审"
	}
	detail.Timeline = append(detail.Timeline, startItem)

	// 驳回到指定节点：仅发起人 + 已走过的审批节点（不含当前），避免列出全部分支里同名「审核人」
	// 驳回目标取自办理实例（子流程内）历史
	rejectHis := his
	if mainView {
		rejectHis, _ = s.repo.ListHisTasks(ctx, workID)
	}
	if detail.TaskID > 0 && detail.RejectNodeKey == "" {
		seen := map[string]bool{}
		rejectStartName, rejectStartKey := startName, startKey
		if mainView && workMc != nil && workMc.NodeConfig != nil {
			rejectStartName = workMc.NodeConfig.NodeName
			if rejectStartName == "" {
				rejectStartName = "发起人"
			}
			rejectStartKey = workMc.NodeConfig.NodeKey
		}
		if rejectStartKey != "" && rejectStartKey != curTaskNodeKey {
			detail.RejectTargets = append(detail.RejectTargets, dto.RejectTarget{
				NodeKey: rejectStartKey, NodeName: rejectStartName,
			})
			seen[rejectStartKey] = true
		}
		for _, h := range rejectHis {
			if h.NodeKey == "" || h.NodeKey == curTaskNodeKey || seen[h.NodeKey] {
				continue
			}
			if h.NodeType != engine.NodeApproval && h.NodeType != engine.NodeStart {
				continue
			}
			name := h.NodeName
			if name == "" {
				name = h.NodeKey
			}
			detail.RejectTargets = append(detail.RejectTargets, dto.RejectTarget{
				NodeKey: h.NodeKey, NodeName: name,
			})
			seen[h.NodeKey] = true
		}
		detail.RejectTargets = dedupeRejectTargetNames(detail.RejectTargets)
		detail.RollbackTargets = append([]dto.RejectTarget{}, detail.RejectTargets...)
	}

	// 撤销 / 催办：仍仅发起人（管理员干预走「流程监控」终止/转办）
	detail.CanRevoke = viewInst.InstanceState == model.InstActive &&
		viewInst.CreateID == userID &&
		engine.ProcessSettingBool(viewInst, "allowRevocation", true)
	detail.CanUrge = viewInst.InstanceState == model.InstActive &&
		(viewInst.CreateID == userID || detail.CanConsent)
	isMgr := s.isInstanceProcessManager(ctx, userID, viewInst)
	detail.CanTerminate = viewInst.InstanceState == model.InstActive && isMgr
	if detail.CanTerminate {
		actives, _ := s.repo.ListActiveTasksByInstance(ctx, workID)
		for _, t := range actives {
			if t.TaskType != model.TaskMajor || t.TaskState != model.TaskActive {
				continue
			}
			detail.CanAdminTransfer = true
			if detail.TaskID == 0 {
				detail.TaskID = t.ID
				detail.WorkNodeKey = t.NodeKey
				detail.WorkNodeName = t.NodeName
			}
			if len(detail.TaskActors) == 0 {
				actors, _ := s.repo.ListTaskActors(ctx, t.ID)
				for _, a := range actors {
					detail.TaskActors = append(detail.TaskActors, dto.TaskActorItem{
						ID: a.ID, ActorID: a.ActorID, ActorName: a.ActorName,
						ActorType: a.ActorType, ActorState: a.ActorState, Weight: a.Weight,
					})
				}
			}
			break
		}
	}
	// 暂存单不评论；其余能看详情即可评论
	detail.CanComment = viewInst.InstanceState != model.InstDraft

	appendComments := func(id uint64) {
		comments, _ := s.repo.ListInstanceComments(ctx, id)
		for _, c := range comments {
			detail.Comments = append(detail.Comments, dto.CommentItem{
				ID: c.ID, UserID: c.UserID, UserName: c.UserName, Content: c.Content,
				CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}
	appendComments(viewID)
	if mainView && workID != viewID {
		appendComments(workID)
	}

	for _, h := range his {
		as := actorMap[h.ID]
		if len(as) == 0 {
			// 子流程占位等无参与人历史：按任务终态展示，避免被当成「进行中」
			item := dto.TimelineItem{
				NodeName:  h.NodeName,
				NodeKey:   h.NodeKey,
				NodeType:  h.NodeType,
				TaskState: h.TaskState,
				CreatedAt: h.CreatedAt.Format("2006-01-02 15:04:05"),
			}
			if h.FinishTime != nil {
				item.FinishTime = h.FinishTime.Format("2006-01-02 15:04:05")
			}
			switch h.TaskState {
			case model.TaskComplete:
				item.ActorName = "系统"
				item.ActorState = model.ActorAgree
				if h.TaskType == model.TaskSub || h.NodeType == engine.NodeSubProcess {
					item.Opinion = "子流程完成"
				}
			case model.TaskReject, model.TaskTerminate:
				item.ActorName = "系统"
				item.ActorState = model.ActorRefuse
				if item.Opinion == "" {
					if h.NodeType == engine.NodeAutoReject {
						item.Opinion = "自动拒绝"
					} else if h.TaskState == model.TaskTerminate {
						item.Opinion = "已终止"
					} else {
						item.Opinion = "已拒绝"
					}
				}
			default:
				item.ActorName = "系统"
				item.ActorState = model.ActorPending
			}
			detail.Timeline = append(detail.Timeline, item)
			continue
		}
		for _, a := range as {
			// 跳过未真正办理的人，避免回退/或签后流转记录刷一堆「空操作」
			if a.ActorState == model.ActorSkip {
				continue
			}
			ft := ""
			if a.FinishTime != nil {
				ft = a.FinishTime.Format("2006-01-02 15:04:05")
			}
			detail.Timeline = append(detail.Timeline, s.timelineItem(
				h.NodeName, h.NodeKey, h.NodeType, h.TaskState,
				a.ActorID, a.ActorName, a.AgentID, a.ActorState, a.Opinion, ft,
				a.CreatedAt.Format("2006-01-02 15:04:05"),
			))
		}
	}

	// 活动任务：只展示「已处理」+「当前轮到」的人，避免依次/后加签把未轮到的人提前写进流转记录
	actives, _ := s.repo.ListActiveTasksByInstance(ctx, viewID)
	for _, t := range actives {
		// 子流程占位无审批人：必须写入流转，且不能走依次审批分支（ExamineMode 默认值也会 continue）
		if t.TaskType == model.TaskSub {
			detail.Timeline = append(detail.Timeline, dto.TimelineItem{
				NodeName:   t.NodeName,
				NodeKey:    t.NodeKey,
				NodeType:   t.NodeType,
				TaskState:  t.TaskState,
				ActorName:  "系统",
				ActorState: model.ActorPending,
				Opinion:    "等待子流程",
				CreatedAt:  t.CreatedAt.Format("2006-01-02 15:04:05"),
			})
			continue
		}

		actors, _ := s.repo.ListTaskActors(ctx, t.ID)
		appendActiveActor := func(a model.FlowTaskActor) {
			ft := ""
			if a.FinishTime != nil {
				ft = a.FinishTime.Format("2006-01-02 15:04:05")
			}
			detail.Timeline = append(detail.Timeline, s.timelineItem(
				t.NodeName, t.NodeKey, t.NodeType, t.TaskState,
				a.ActorID, a.ActorName, a.AgentID, a.ActorState, a.Opinion, ft,
				a.CreatedAt.Format("2006-01-02 15:04:05"),
			))
		}

		// 依次审批：未轮到的 pending 不展示
		if t.ExamineMode == 1 {
			for _, a := range actors {
				if a.ActorState != model.ActorPending && a.ActorState != model.ActorSkip {
					appendActiveActor(a)
				}
			}
			minW := 0
			for _, a := range actors {
				if a.ActorState != model.ActorPending {
					continue
				}
				if minW == 0 || a.Weight < minW {
					minW = a.Weight
				}
			}
			if minW != 0 {
				for _, a := range actors {
					if a.ActorState == model.ActorPending && a.Weight == minW {
						appendActiveActor(a)
					}
				}
			}
			continue
		}

		// 会签 / 或签：节点上的人都已到达，全部展示（跳过被或签跳过的）
		for _, a := range actors {
			if a.ActorState == model.ActorSkip {
				continue
			}
			appendActiveActor(a)
		}
	}

	for i := range detail.Timeline {
		s.enrichSubProcessTimeline(ctx, inst, &detail.Timeline[i])
	}
	return detail, nil
}

// resolveRootInstance 向上找到根主流程实例
// canAccessInstance 详情/评论可见：发起人、办理/抄送/认领池、父实例相关人、流程管理员
func (s *FlowService) canAccessInstance(ctx context.Context, inst *model.FlowInstance, userID uint64) bool {
	if inst == nil || userID == 0 {
		return false
	}
	if inst.InstanceState == model.InstDraft {
		return inst.CreateID == userID
	}
	seen := map[uint64]struct{}{}
	var walk func(*model.FlowInstance) bool
	walk = func(i *model.FlowInstance) bool {
		if i == nil {
			return false
		}
		if _, ok := seen[i.ID]; ok {
			return false
		}
		seen[i.ID] = struct{}{}
		if i.CreateID == userID || s.userParticipated(ctx, i, userID) {
			return true
		}
		if i.ParentInstanceID > 0 {
			p, err := s.repo.GetInstance(ctx, i.ParentInstanceID)
			if err == nil && walk(p) {
				return true
			}
		}
		return false
	}
	if walk(inst) {
		return true
	}
	ids, all, err := s.listManagedProcessIDs(ctx, userID)
	if err != nil {
		return false
	}
	if all {
		return true
	}
	for _, id := range ids {
		if id == inst.ProcessID {
			return true
		}
	}
	return false
}

func (s *FlowService) userParticipated(ctx context.Context, inst *model.FlowInstance, userID uint64) bool {
	if inst == nil || userID == 0 {
		return false
	}
	allow := map[uint64]struct{}{userID: {}}
	for _, id := range s.actPrincipals(ctx, inst, userID) {
		if id > 0 {
			allow[id] = struct{}{}
		}
	}
	roleSet := map[uint64]struct{}{}
	if roles, err := s.org().UserRoleIDs(userID); err == nil {
		for _, id := range roles {
			roleSet[id] = struct{}{}
		}
	}
	actives, _ := s.repo.ListActiveTasksByInstance(ctx, inst.ID)
	for _, t := range actives {
		actors, _ := s.repo.ListTaskActors(ctx, t.ID)
		for _, a := range actors {
			if a.ActorType == model.ActorTypeRole {
				if _, ok := roleSet[a.ActorID]; ok {
					return true
				}
				continue
			}
			if _, ok := allow[a.ActorID]; ok {
				return true
			}
		}
	}
	hisActors, _ := s.repo.ListHisActorsByInstance(ctx, inst.ID)
	for _, a := range hisActors {
		if a.ActorType == model.ActorTypeRole {
			continue
		}
		if _, ok := allow[a.ActorID]; ok {
			return true
		}
	}
	return false
}

func (s *FlowService) resolveRootInstance(ctx context.Context, inst *model.FlowInstance) *model.FlowInstance {
	if inst == nil {
		return nil
	}
	cur := inst
	for depth := 0; depth < 8 && cur.ParentInstanceID > 0; depth++ {
		parent, err := s.repo.GetInstance(ctx, cur.ParentInstanceID)
		if err != nil || parent == nil {
			break
		}
		cur = parent
	}
	return cur
}

// applyRootProcessDisplay 子实例详情/列表标题改用主流程名，并保留子流程名供角标
func (s *FlowService) applyRootProcessDisplay(ctx context.Context, inst *model.FlowInstance, processName, subProcessName *string, isSub *bool) {
	if inst == nil || processName == nil || subProcessName == nil || isSub == nil {
		return
	}
	if inst.ParentInstanceID == 0 {
		return
	}
	*isSub = true
	*subProcessName = inst.ProcessName
	root := s.resolveRootInstance(ctx, inst)
	if root != nil && root.ProcessName != "" {
		*processName = root.ProcessName
	}
	if *processName == "" {
		*processName = inst.ProcessName
	}
}

// enrichSubProcessTimeline 子流程节点挂上子实例，便于主流程流转记录点开查看
func (s *FlowService) enrichSubProcessTimeline(ctx context.Context, parent *model.FlowInstance, item *dto.TimelineItem) {
	if parent == nil || item == nil || item.NodeKey == "" {
		return
	}
	if item.NodeType != engine.NodeSubProcess {
		return
	}
	child, err := s.repo.GetLatestChildByParentNode(ctx, parent.ID, item.NodeKey)
	if err != nil || child == nil {
		// 回退：从活动 TaskSub 的 payload.childInstanceId 读取
		actives, _ := s.repo.ListActiveTasksByInstance(ctx, parent.ID)
		for i := range actives {
			t := &actives[i]
			if t.NodeKey != item.NodeKey || t.TaskType != model.TaskSub {
				continue
			}
			var payload struct {
				ChildInstanceID uint64 `json:"childInstanceId"`
			}
			_ = json.Unmarshal([]byte(t.Payload), &payload)
			if payload.ChildInstanceID > 0 {
				child, _ = s.repo.GetInstance(ctx, payload.ChildInstanceID)
			}
			break
		}
	}
	if child == nil {
		return
	}
	item.ChildInstanceID = child.ID
	item.ChildProcessName = child.ProcessName
	if item.ChildProcessName == "" {
		item.ChildProcessName = item.NodeName
	}
	item.ChildInstanceState = child.InstanceState
	starter := child.CreateBy
	if starter == "" {
		starter = parent.CreateBy
	}
	item.Opinion = fmt.Sprintf("%s 发起的子流程", starter)
	if item.ActorName == "" || item.ActorName == "系统" {
		item.ActorName = starter
		if item.ActorName == "" {
			item.ActorName = "系统"
		}
	}
	if child.InstanceState == model.InstActive {
		item.ActorState = model.ActorPending
		item.TaskState = model.TaskActive
	} else if child.InstanceState == model.InstComplete && item.ActorState == model.ActorPending {
		item.ActorState = model.ActorAgree
		item.TaskState = model.TaskComplete
	}
}

// timelineItem 组装流转记录项；仅真实代批（agent≠槽位本人）时附带代批人
func (s *FlowService) timelineItem(
	nodeName, nodeKey string, nodeType int, taskState int8,
	actorID uint64, actorName string, agentID uint64, actorState int8, opinion, finishTime, createdAt string,
) dto.TimelineItem {
	item := dto.TimelineItem{
		NodeName:   nodeName,
		NodeKey:    nodeKey,
		NodeType:   nodeType,
		TaskState:  taskState,
		ActorName:  actorName,
		ActorState: actorState,
		Opinion:    opinion,
		FinishTime: finishTime,
		CreatedAt:  createdAt,
	}
	if agentID > 0 && agentID != actorID {
		if n, _, ok := s.org().GetUser(agentID); ok && n != "" {
			item.AgentName = n
		} else {
			item.AgentName = fmt.Sprintf("用户#%d", agentID)
		}
	}
	return item
}

// dedupeRejectTargetNames 同名节点加序号，避免下拉里全是「审核人」
func dedupeRejectTargetNames(list []dto.RejectTarget) []dto.RejectTarget {
	if len(list) <= 1 {
		return list
	}
	counts := map[string]int{}
	for _, t := range list {
		counts[t.NodeName]++
	}
	seen := map[string]int{}
	out := make([]dto.RejectTarget, 0, len(list))
	for _, t := range list {
		name := t.NodeName
		if counts[name] > 1 {
			seen[name]++
			name = fmt.Sprintf("%s（%d）", name, seen[name])
		}
		out = append(out, dto.RejectTarget{NodeKey: t.NodeKey, NodeName: name})
	}
	return out
}

func formConfigOf(n *engine.Node) []map[string]interface{} {
	if n == nil || n.ExtendConfig == nil {
		return nil
	}
	raw, ok := n.ExtendConfig["formConfig"]
	if !ok {
		return nil
	}
	b, _ := json.Marshal(raw)
	var list []map[string]interface{}
	_ = json.Unmarshal(b, &list)
	return list
}

func (s *FlowService) TickDelays(ctx context.Context) (int, error) {
	n, _, _, err := s.TickTimers(ctx)
	return n, err
}

// TickTimers 延时节点 + 审批超时 + 审批提醒
// 返回：延时推进数、超时处理数、提醒发送数
func (s *FlowService) TickTimers(ctx context.Context) (delayN, timeoutN, remindN int, err error) {
	eng := s.engine()
	now := time.Now()
	touched := map[uint64]struct{}{}

	delays, err := s.repo.ListExpiredDelayTasks(ctx, now, 50)
	if err != nil {
		return 0, 0, 0, err
	}
	for _, t := range delays {
		if e := eng.CompleteDelayTask(ctx, t.ID); e == nil {
			delayN++
			touched[t.InstanceID] = struct{}{}
		}
	}

	timeouts, err := s.repo.ListExpiredMajorTimeoutTasks(ctx, now, 50)
	if err != nil {
		return delayN, 0, 0, err
	}
	for _, t := range timeouts {
		if e := eng.ExecuteTimeout(ctx, t.ID); e == nil {
			timeoutN++
			touched[t.InstanceID] = struct{}{}
		}
	}

	for id := range touched {
		inst, ie := s.repo.GetInstance(ctx, id)
		if ie != nil || inst == nil {
			continue
		}
		if err := s.bootstrapSubProcesses(ctx, id, inst.CreateID, inst.CreateBy); err != nil {
			logFlowSideEffect("bootstrapSubProcesses in tick", err)
		}
		s.checkParentResume(ctx, inst)
	}

	// 扫描卡住的子流程占位：尚未创建子实例的 TaskSub（详情自愈之外的兜底）
	if subs, se := s.repo.ListActiveSubTasks(ctx, 100); se == nil {
		seen := map[uint64]struct{}{}
		for i := range subs {
			id := subs[i].InstanceID
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			inst, ie := s.repo.GetInstance(ctx, id)
			if ie != nil || inst == nil || inst.InstanceState != model.InstActive {
				continue
			}
			if err := s.bootstrapSubProcesses(ctx, id, inst.CreateID, inst.CreateBy); err != nil {
				logFlowSideEffect("bootstrapSubProcesses for active TaskSub", err)
			}
		}
	}

	majors, err := s.repo.ListActiveMajorTasks(ctx, 200)
	if err != nil {
		return delayN, timeoutN, 0, err
	}
	msgSvc := msgsvc.NewMessageService(msgrepo.NewMessageRepository())
	for i := range majors {
		t := &majors[i]
		receivers, e := eng.RemindDue(ctx, t)
		if e != nil || len(receivers) == 0 {
			continue
		}
		inst, ie := s.repo.GetInstance(ctx, t.InstanceID)
		if ie != nil || inst == nil {
			continue
		}
		title := fmt.Sprintf("审批提醒：%s", inst.ProcessName)
		content := fmt.Sprintf("流程「%s」节点「%s」待您处理，请尽快审批。", inst.ProcessName, t.NodeName)
		if se := msgSvc.Send(ctx, &msgdto.SendMessageRequest{
			Title:       title,
			Content:     content,
			Type:        1,
			ReceiverIDs: receivers,
		}, 0, "系统"); se != nil {
			continue
		}
		if me := eng.MarkReminded(ctx, t); me == nil {
			remindN++
		}
	}
	return delayN, timeoutN, remindN, nil
}

func (s *FlowService) Revoke(ctx context.Context, req *dto.RevokeRequest, userID uint64) error {
	inst, err := s.repo.GetInstance(ctx, req.InstanceID)
	if err != nil {
		return err
	}
	if err := s.engine().Revoke(ctx, req.InstanceID, userID, req.Opinion); err != nil {
		return err
	}
	// 直接撤销子实例时，父流程按失败收尾（父级联撤销子实例时父自己已结束）
	if inst.ParentInstanceID > 0 {
		s.checkParentResume(ctx, inst)
	}
	return nil
}

func (s *FlowService) Transfer(ctx context.Context, req *dto.TransferRequest, userID uint64) error {
	return s.engine().Transfer(ctx, req.TaskID, userID, req.ToUserID, req.ToName, req.Opinion)
}

// Terminate 流程管理员强制终止
func (s *FlowService) Terminate(ctx context.Context, req *dto.TerminateRequest, userID uint64) error {
	inst, err := s.repo.GetInstance(ctx, req.InstanceID)
	if err != nil {
		return err
	}
	if !s.isInstanceProcessManager(ctx, userID, inst) {
		return fmt.Errorf("无权终止该流程")
	}
	if err := s.engine().Terminate(ctx, req.InstanceID, userID, req.Opinion); err != nil {
		return err
	}
	if inst.ParentInstanceID > 0 {
		s.checkParentResume(ctx, inst)
	}
	return nil
}

// AdminTransfer 流程管理员转办当前待办
func (s *FlowService) AdminTransfer(ctx context.Context, req *dto.TransferRequest, userID uint64) error {
	task, err := s.repo.GetTask(ctx, req.TaskID)
	if err != nil {
		return err
	}
	inst, err := s.repo.GetInstance(ctx, task.InstanceID)
	if err != nil {
		return err
	}
	if !s.isInstanceProcessManager(ctx, userID, inst) {
		return fmt.Errorf("无权转办该任务")
	}
	return s.engine().AdminTransfer(ctx, req.TaskID, userID, req.ToUserID, req.ToName, req.Opinion, req.FromActorID)
}

func (s *FlowService) Rollback(ctx context.Context, req *dto.RollbackRequest, userID uint64) error {
	return s.engine().Rollback(ctx, req.TaskID, userID, req.Opinion, req.TargetNodeKey)
}

func (s *FlowService) AppendActor(ctx context.Context, req *dto.AppendRequest, userID uint64) error {
	pos := req.Position
	if pos != 1 {
		pos = 2
	}
	return s.engine().AppendActor(ctx, req.TaskID, userID, req.ToUserID, req.ToName, pos)
}

func (s *FlowService) RemoveActor(ctx context.Context, req *dto.RemoveActorRequest, userID uint64) error {
	return s.engine().RemoveActor(ctx, req.TaskID, userID, req.ActorID)
}

func (s *FlowService) AddRuntimeCC(ctx context.Context, req *dto.RuntimeCCRequest, userID uint64, userName string) error {
	users := make([]engine.UserRef, 0, len(req.Users))
	for _, u := range req.Users {
		users = append(users, engine.UserRef{ID: u.ID, Name: u.Name})
	}
	if err := s.engine().AddRuntimeCC(ctx, req.TaskID, userID, users); err != nil {
		return err
	}
	// 站内信通知抄送人
	ids := make([]uint64, 0, len(users))
	for _, u := range users {
		if u.ID > 0 && u.ID != userID {
			ids = append(ids, u.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	task, err := s.repo.GetTask(ctx, req.TaskID)
	instName := "流程"
	nodeName := ""
	if err == nil {
		nodeName = task.NodeName
		if inst, ie := s.repo.GetInstance(ctx, task.InstanceID); ie == nil {
			instName = inst.ProcessName
		}
	}
	msgSvc := msgsvc.NewMessageService(msgrepo.NewMessageRepository())
	title := fmt.Sprintf("抄送：%s", instName)
	content := fmt.Sprintf("%s 在节点「%s」抄送了流程「%s」给您，请知悉。", userName, nodeName, instName)
	_ = msgSvc.Send(ctx, &msgdto.SendMessageRequest{
		Title: title, Content: content, Type: 1, ReceiverIDs: ids,
	}, userID, userName)
	return nil
}

func (s *FlowService) AddComment(ctx context.Context, req *dto.CommentRequest, userID uint64, userName string) error {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return fmt.Errorf("请输入评论内容")
	}
	inst, err := s.repo.GetInstance(ctx, req.InstanceID)
	if err != nil || inst == nil {
		return ErrTaskNotFound
	}
	if inst.InstanceState == model.InstDraft {
		return fmt.Errorf("暂存单不可评论")
	}
	if !s.canAccessInstance(ctx, inst, userID) {
		return fmt.Errorf("无权评论")
	}
	c := &model.FlowInstanceComment{
		InstanceID: req.InstanceID,
		UserID:     userID,
		UserName:   userName,
		Content:    content,
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreateInstanceComment(ctx, c); err != nil {
		return err
	}
	// 站内信：@的人 + 发起人 + 当前待办处理人（不含评论者本人）
	s.notifyComment(ctx, inst, userID, userName, content, req.MentionUserIDs)
	return nil
}

func (s *FlowService) notifyComment(ctx context.Context, inst *model.FlowInstance, fromID uint64, fromName, content string, mentionIDs []uint64) {
	if inst == nil {
		return
	}
	snippet := content
	if len([]rune(snippet)) > 80 {
		r := []rune(snippet)
		snippet = string(r[:80]) + "…"
	}
	msgSvc := msgsvc.NewMessageService(msgrepo.NewMessageRepository())

	mentionSet := map[uint64]struct{}{}
	for _, id := range mentionIDs {
		if id > 0 && id != fromID {
			mentionSet[id] = struct{}{}
		}
	}
	if len(mentionSet) > 0 {
		ids := make([]uint64, 0, len(mentionSet))
		for id := range mentionSet {
			ids = append(ids, id)
		}
		title := fmt.Sprintf("有人@了你：%s", inst.ProcessName)
		body := fmt.Sprintf("%s 在流程「%s」的评论中提到了你：%s", fromName, inst.ProcessName, snippet)
		_ = msgSvc.Send(ctx, &msgdto.SendMessageRequest{
			Title: title, Content: body, Type: 1, ReceiverIDs: ids,
		}, fromID, fromName)
	}

	recv := map[uint64]struct{}{}
	if inst.CreateID > 0 && inst.CreateID != fromID {
		recv[inst.CreateID] = struct{}{}
	}
	eng := s.engine()
	if ids, err := eng.ListUrgeReceiverIDs(ctx, inst.ID); err == nil {
		for _, id := range ids {
			if id > 0 && id != fromID {
				recv[id] = struct{}{}
			}
		}
	}
	// 已单独发过 @ 通知的不再重复发普通评论信
	for id := range mentionSet {
		delete(recv, id)
	}
	if len(recv) == 0 {
		return
	}
	ids := make([]uint64, 0, len(recv))
	for id := range recv {
		ids = append(ids, id)
	}
	title := fmt.Sprintf("流程评论：%s", inst.ProcessName)
	body := fmt.Sprintf("%s 评论了流程「%s」：%s", fromName, inst.ProcessName, snippet)
	_ = msgSvc.Send(ctx, &msgdto.SendMessageRequest{
		Title: title, Content: body, Type: 1, ReceiverIDs: ids,
	}, fromID, fromName)
}

func (s *FlowService) Urge(ctx context.Context, req *dto.UrgeRequest, userID uint64, userName string) error {
	inst, err := s.repo.GetInstance(ctx, req.InstanceID)
	if err != nil {
		return err
	}
	if inst.InstanceState != model.InstActive {
		return fmt.Errorf("流程已结束")
	}
	// 发起人或当前任务处理人（管理员催办走流程监控场景较少，监控以终止/转办为主）
	allowed := inst.CreateID == userID
	if !allowed {
		actives, _ := s.repo.ListActiveTasksByInstance(ctx, inst.ID)
		ps := s.actPrincipals(ctx, inst, userID)
		for _, t := range actives {
			actors, _ := s.repo.ListTaskActors(ctx, t.ID)
			if engine.CanUserAct(&t, actors, userID, ps...) {
				allowed = true
				break
			}
		}
	}
	if !allowed {
		return fmt.Errorf("无权催办")
	}
	eng := s.engine()
	ok, wait := eng.MarkUrged(inst)
	if !ok {
		return fmt.Errorf("催办过于频繁，请 %d 秒后再试", wait)
	}
	_ = s.repo.UpdateInstance(ctx, inst)
	ids, err := eng.ListUrgeReceiverIDs(ctx, inst.ID)
	if err != nil {
		return err
	}
	// 不催自己
	filtered := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id != userID {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == 0 {
		return fmt.Errorf("当前没有可催办的审批人")
	}
	msgSvc := msgsvc.NewMessageService(msgrepo.NewMessageRepository())
	title := fmt.Sprintf("催办：%s", inst.ProcessName)
	content := fmt.Sprintf("%s 催促您尽快处理流程「%s」（当前节点：%s）", userName, inst.ProcessName, inst.CurrentNodeName)
	return msgSvc.Send(ctx, &msgdto.SendMessageRequest{
		Title:       title,
		Content:     content,
		Type:        1,
		ReceiverIDs: filtered,
	}, userID, userName)
}

func (s *FlowService) BatchConsent(ctx context.Context, req *dto.BatchConsentRequest, userID uint64) (*dto.BatchOperateResult, error) {
	res := &dto.BatchOperateResult{}
	opinion := strings.TrimSpace(req.Opinion)
	if opinion == "" {
		opinion = "同意"
	}
	for _, tid := range req.TaskIDs {
		if tid == 0 {
			continue
		}
		task, err := s.repo.GetTask(ctx, tid)
		if err != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("任务%d不存在", tid))
			continue
		}
		inst, err := s.repo.GetInstance(ctx, task.InstanceID)
		if err != nil || inst == nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("任务%d实例不存在", tid))
			continue
		}
		if task.NodeType == engine.NodeStart {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("「%s」改单任务请单独重新提交", inst.ProcessName))
			continue
		}
		if !engine.ProcessSettingBool(inst, "allowBatchOperate", true) {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("「%s」未开启批量操作", inst.ProcessName))
			continue
		}
		// 节点有可写必填且当前值为空：不可批量空表通过
		if s.nodeHasUnfilledWritableRequired(ctx, inst, task.NodeKey) {
			res.Failed++
			name := inst.ProcessName
			if task.NodeName != "" {
				name = fmt.Sprintf("%s / %s", inst.ProcessName, task.NodeName)
			}
			res.Errors = append(res.Errors, fmt.Sprintf("「%s」含未填必填项，请单独打开审批填写", name))
			continue
		}
		if err := s.Consent(ctx, &dto.ConsentRequest{TaskID: tid, Opinion: opinion}, userID); err != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("「%s」：%v", inst.ProcessName, err))
			continue
		}
		res.Success++
	}
	return res, nil
}

func (s *FlowService) BatchReject(ctx context.Context, req *dto.BatchRejectRequest, userID uint64) (*dto.BatchOperateResult, error) {
	res := &dto.BatchOperateResult{}
	opinion := strings.TrimSpace(req.Opinion)
	if opinion == "" {
		opinion = "驳回"
	}
	for _, tid := range req.TaskIDs {
		if tid == 0 {
			continue
		}
		task, err := s.repo.GetTask(ctx, tid)
		if err != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("任务%d不存在", tid))
			continue
		}
		inst, err := s.repo.GetInstance(ctx, task.InstanceID)
		if err != nil || inst == nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("任务%d实例不存在", tid))
			continue
		}
		if task.NodeType == engine.NodeStart {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("「%s」改单任务请单独处理", inst.ProcessName))
			continue
		}
		if !engine.ProcessSettingBool(inst, "allowBatchOperate", true) {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("「%s」未开启批量操作", inst.ProcessName))
			continue
		}
		strategy := req.RejectStrategy
		rejectKey := ""
		// 0=按节点配置：策略 3 且未预置目标时批量无法选节点，提前失败并提示
		if strategy == 0 {
			st, key, nodeName := resolveRejectMeta(inst.ModelContent, task.NodeKey)
			if st == 3 && key == "" {
				name := nodeName
				if name == "" {
					name = task.NodeName
				}
				res.Failed++
				res.Errors = append(res.Errors, fmt.Sprintf("「%s」节点「%s」需指定驳回目标，请单独驳回或勾选统一终止", inst.ProcessName, name))
				continue
			}
			rejectKey = key
		}
		if err := s.Reject(ctx, &dto.RejectRequest{
			TaskID: tid, Opinion: opinion, RejectStrategy: strategy, RejectNodeKey: rejectKey,
		}, userID); err != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("「%s」：%v", inst.ProcessName, err))
			continue
		}
		res.Success++
	}
	return res, nil
}

// resolveRejectMeta 读取节点驳回策略与预置目标
func resolveRejectMeta(modelJSON, nodeKey string) (strategy int, rejectKey, nodeName string) {
	strategy = 3
	mc, err := engine.ParseModel(modelJSON)
	if err != nil || mc == nil {
		return
	}
	n := engine.FindNode(mc.NodeConfig, nodeKey)
	if n == nil {
		return
	}
	nodeName = n.NodeName
	strategy = n.RejectStrategy
	if strategy == 0 {
		strategy = 3 // 与设计器默认「指定节点」一致
	}
	if n.ExtendConfig != nil {
		if v, ok := n.ExtendConfig["rejectNodeKey"].(string); ok {
			rejectKey = strings.TrimSpace(v)
		}
	}
	return
}

func (s *FlowService) GetMyDelegate(ctx context.Context, userID uint64) (*dto.DelegateSetting, error) {
	d, err := s.repo.GetUserDelegate(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dto.DelegateSetting{}, nil
		}
		return nil, err
	}
	return &dto.DelegateSetting{
		ToUserID:   d.ToUserID,
		ToUserName: d.ToUserName,
		Enabled:    d.Enabled == 1,
		Remark:     d.Remark,
	}, nil
}

func (s *FlowService) SetMyDelegate(ctx context.Context, userID uint64, req *dto.SetDelegateRequest) error {
	if req.Clear || req.ToUserID == 0 {
		return s.repo.ClearUserDelegate(ctx, userID)
	}
	if req.ToUserID == userID {
		return fmt.Errorf("不能委托给自己")
	}
	name := strings.TrimSpace(req.ToUserName)
	if name == "" {
		if n, _, ok := s.org().GetUser(req.ToUserID); ok {
			name = n
		}
	}
	enabled := int8(0)
	if req.Enabled {
		enabled = 1
	}
	return s.repo.UpsertUserDelegate(ctx, &model.FlowUserDelegate{
		UserID:     userID,
		ToUserID:   req.ToUserID,
		ToUserName: name,
		Enabled:    enabled,
		Remark:     req.Remark,
	})
}
