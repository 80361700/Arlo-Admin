package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"arlo-admin/internal/modules/flow/dto"
	"arlo-admin/internal/modules/flow/engine"
	"arlo-admin/internal/modules/flow/model"
	"arlo-admin/internal/modules/flow/repository"

	"gorm.io/gorm"
)

var (
	ErrCategoryHasProcess = errors.New("请先删除关联流程")
	ErrProcessKeyExists   = errors.New("流程标识已存在")
	ErrInvalidModel       = errors.New("流程模型无效")
	ErrCategoryHasForm    = errors.New("请先删除关联表单")
	ErrFormCodeExists     = errors.New("模板编码已存在")
	ErrFormBound          = errors.New("表单已被流程引用，无法删除")
	ErrHistoryMismatch    = errors.New("历史版本与流程不匹配")
)

type FlowService struct {
	repo *repository.FlowRepository
}

func NewFlowService(repo *repository.FlowRepository) *FlowService {
	return &FlowService{repo: repo}
}

func (s *FlowService) ListCategoryTree(ctx context.Context, userID uint64) ([]dto.CategoryListItem, error) {
	cats, err := s.repo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	procs, err := s.repo.ListProcesses(ctx, nil, "")
	if err != nil {
		return nil, err
	}
	byCat := map[uint64][]dto.ProcessBrief{}
	for i := range procs {
		if !s.isProcessManager(ctx, userID, &procs[i]) {
			continue
		}
		b := toBrief(&procs[i])
		byCat[procs[i].CategoryID] = append(byCat[procs[i].CategoryID], b)
	}
	out := make([]dto.CategoryListItem, 0, len(cats))
	for _, c := range cats {
		list := byCat[c.ID]
		if list == nil {
			list = []dto.ProcessBrief{}
		}
		out = append(out, dto.CategoryListItem{
			CategoryID:     c.ID,
			CategoryName:   c.Name,
			CategoryRemark: c.Remark,
			CategorySort:   c.Sort,
			ProcessList:    list,
		})
	}
	return out, nil
}

func (s *FlowService) ListCategoryOptions(ctx context.Context) ([]dto.CategoryOption, error) {
	cats, err := s.repo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.CategoryOption, 0, len(cats))
	for _, c := range cats {
		out = append(out, dto.CategoryOption{ID: c.ID, Name: c.Name})
	}
	return out, nil
}

func (s *FlowService) CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*model.FlowCategory, error) {
	c := &model.FlowCategory{Name: strings.TrimSpace(req.Name), Sort: req.Sort, Remark: req.Remark}
	if err := s.repo.CreateCategory(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *FlowService) UpdateCategory(ctx context.Context, req *dto.UpdateCategoryRequest) error {
	c, err := s.repo.GetCategory(ctx, req.ID)
	if err != nil {
		return err
	}
	c.Name = strings.TrimSpace(req.Name)
	c.Sort = req.Sort
	c.Remark = req.Remark
	return s.repo.UpdateCategory(ctx, c)
}

func (s *FlowService) DeleteCategory(ctx context.Context, id uint64) error {
	n, err := s.repo.CountProcessByCategory(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrCategoryHasProcess
	}
	return s.repo.DeleteCategory(ctx, id)
}

func (s *FlowService) GetProcess(ctx context.Context, id uint64, userID uint64) (*dto.ProcessDetail, error) {
	p, err := s.repo.GetProcess(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureProcessManager(ctx, userID, p); err != nil {
		return nil, err
	}
	return toDetail(p), nil
}

func (s *FlowService) SaveProcess(ctx context.Context, req *dto.SaveProcessRequest, userID uint64) (*dto.ProcessDetail, error) {
	if err := validateModelContent(req.ModelContent); err != nil {
		return nil, err
	}
	iconJSON := encodeIcon(req.ProcessIcon, req.ProcessBgcolor)
	modelStr, err := marshalJSON(req.ModelContent)
	if err != nil {
		return nil, err
	}
	formStr, err := marshalJSON(req.ProcessForm)
	if err != nil {
		return nil, err
	}
	if formStr == "" || formStr == "null" {
		formStr = `{"widgetList":[]}`
	}
	settingStr, err := marshalJSON(req.ProcessSetting)
	if err != nil {
		return nil, err
	}
	if settingStr == "" || settingStr == "null" {
		settingStr = `{}`
	}
	permStr, err := marshalJSON(req.ProcessPermissionList)
	if err != nil {
		return nil, err
	}
	if permStr == "" || permStr == "null" {
		permStr = `[]`
	}
	procType := req.ProcessType
	if procType == "" {
		procType = "main"
	}
	if procType != "main" && procType != "business" && procType != "child" {
		procType = "main"
	}
	key := strings.TrimSpace(req.ProcessKey)

	if req.ProcessID != nil && *req.ProcessID > 0 {
		p, err := s.repo.GetProcess(ctx, *req.ProcessID)
		if err != nil {
			return nil, err
		}
		if err := s.ensureProcessManager(ctx, userID, p); err != nil {
			return nil, err
		}
		// 编辑时未传 key：沿用原标识（前端禁用修改）
		if key == "" {
			key = p.ProcessKey
		}
		if p.ProcessKey != key {
			if other, e := s.repo.GetProcessByKey(ctx, key); e == nil && other.ID != p.ID {
				return nil, ErrProcessKeyExists
			} else if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
				return nil, e
			}
		}
		// 先把「即将被顶替」的当前版写入历史，新版本只留在主表
		prevSnap := snapshotHistory(p, userID)
		p.CategoryID = req.CategoryID
		p.ProcessKey = key
		p.ProcessName = strings.TrimSpace(req.ProcessName)
		p.ProcessIcon = iconJSON
		// 类型创建后不可改
		p.Remark = req.Remark
		p.ModelContent = modelStr
		p.ProcessForm = formStr
		p.ProcessSetting = settingStr
		p.ProcessPermission = permStr
		p.UpdatedBy = userID
		p.ProcessVersion = p.ProcessVersion + 1
		err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
			if e := s.repo.EnsureProcessHistoryTx(tx, prevSnap); e != nil {
				return e
			}
			return s.repo.UpdateProcessTx(tx, p)
		})
		if err != nil {
			return nil, err
		}
		return toDetail(p), nil
	}

	if key == "" {
		var err error
		key, err = s.allocProcessKey(ctx)
		if err != nil {
			return nil, err
		}
		modelStr = patchModelContentKey(modelStr, key)
	}

	if _, err := s.repo.GetProcessByKey(ctx, key); err == nil {
		return nil, ErrProcessKeyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	p := &model.FlowProcess{
		CategoryID:        req.CategoryID,
		ProcessKey:        key,
		ProcessName:       strings.TrimSpace(req.ProcessName),
		ProcessIcon:       iconJSON,
		ProcessType:       procType,
		ProcessVersion:    1,
		ProcessState:      0,
		Remark:            req.Remark,
		ModelContent:      modelStr,
		ProcessForm:       formStr,
		ProcessSetting:    settingStr,
		ProcessPermission: permStr,
		CreatedBy:         userID,
		UpdatedBy:         userID,
	}
	if err := s.repo.CreateProcess(ctx, p); err != nil {
		return nil, err
	}
	return toDetail(p), nil
}

func (s *FlowService) UpdateState(ctx context.Context, id uint64, state int8, userID uint64) error {
	p, err := s.repo.GetProcess(ctx, id)
	if err != nil {
		return err
	}
	if err := s.ensureProcessManager(ctx, userID, p); err != nil {
		return err
	}
	if state == 1 {
		if err := validateProcessForEnable(p); err != nil {
			return err
		}
	}
	p.ProcessState = state
	return s.repo.UpdateProcess(ctx, p)
}

func (s *FlowService) DeleteProcess(ctx context.Context, id uint64, userID uint64) error {
	p, err := s.repo.GetProcess(ctx, id)
	if err != nil {
		return err
	}
	if err := s.ensureProcessManager(ctx, userID, p); err != nil {
		return err
	}
	// 软删前改写 key，避免唯一索引阻止同名重建
	p.ProcessKey = fmt.Sprintf("%s__del_%d", p.ProcessKey, p.ID)
	if err := s.repo.UpdateProcess(ctx, p); err != nil {
		return err
	}
	return s.repo.DeleteProcess(ctx, id)
}

// ListProcessOptions 子流程节点下拉：仅启用中的「子流程」类型
func (s *FlowService) ListProcessOptions(ctx context.Context, excludeID uint64) ([]dto.ProcessOption, error) {
	list, err := s.repo.ListProcesses(ctx, nil, "")
	if err != nil {
		return nil, err
	}
	out := make([]dto.ProcessOption, 0, len(list))
	for i := range list {
		p := list[i]
		if excludeID > 0 && p.ID == excludeID {
			continue
		}
		if p.ProcessState != 1 {
			continue
		}
		if p.ProcessType != "child" {
			continue
		}
		out = append(out, dto.ProcessOption{
			ID:   p.ID,
			Name: p.ProcessName,
			Key:  p.ProcessKey,
		})
	}
	return out, nil
}

// allocProcessKey 生成未占用的流程标识（flw + 毫秒时间戳 + 4 位随机数字）
func (s *FlowService) allocProcessKey(ctx context.Context) (string, error) {
	for i := 0; i < 8; i++ {
		key := fmt.Sprintf("flw%d%04d", time.Now().UnixMilli(), time.Now().Nanosecond()%10000)
		if len(key) > 64 {
			key = key[:64]
		}
		_, err := s.repo.GetProcessByKey(ctx, key)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return key, nil
		}
		if err != nil {
			return "", err
		}
		time.Sleep(time.Millisecond)
	}
	return "", errors.New("生成流程标识失败")
}

func patchModelContentKey(modelStr, key string) string {
	if modelStr == "" || modelStr == "null" || key == "" {
		return modelStr
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(modelStr), &m); err != nil {
		return modelStr
	}
	m["key"] = key
	b, err := json.Marshal(m)
	if err != nil {
		return modelStr
	}
	return string(b)
}

func (s *FlowService) CloneProcess(ctx context.Context, id uint64, userID uint64) (*dto.ProcessDetail, error) {
	src, err := s.repo.GetProcess(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureProcessManager(ctx, userID, src); err != nil {
		return nil, err
	}
	newKey := fmt.Sprintf("%s_copy_%d", src.ProcessKey, time.Now().Unix()%100000)
	for i := 0; i < 5; i++ {
		if _, e := s.repo.GetProcessByKey(ctx, newKey); errors.Is(e, gorm.ErrRecordNotFound) {
			break
		}
		newKey = fmt.Sprintf("%s_copy_%d", src.ProcessKey, time.Now().UnixNano()%1000000)
	}
	p := &model.FlowProcess{
		CategoryID:        src.CategoryID,
		ProcessKey:        newKey,
		ProcessName:       src.ProcessName + "（副本）",
		ProcessIcon:       src.ProcessIcon,
		ProcessType:       src.ProcessType,
		ProcessVersion:    1,
		ProcessState:      0,
		Remark:            src.Remark,
		ModelContent:      src.ModelContent,
		ProcessForm:       src.ProcessForm,
		ProcessSetting:    src.ProcessSetting,
		ProcessPermission: src.ProcessPermission,
		CreatedBy:         userID,
		UpdatedBy:         userID,
	}
	if err := s.repo.CreateProcess(ctx, p); err != nil {
		return nil, err
	}
	return toDetail(p), nil
}

func snapshotHistory(p *model.FlowProcess, userID uint64) *model.FlowProcessHistory {
	return &model.FlowProcessHistory{
		ProcessID:         p.ID,
		ProcessKey:        p.ProcessKey,
		ProcessName:       p.ProcessName,
		ProcessIcon:       p.ProcessIcon,
		ProcessType:       p.ProcessType,
		ProcessVersion:    p.ProcessVersion,
		Remark:            p.Remark,
		ModelContent:      p.ModelContent,
		ProcessForm:       p.ProcessForm,
		ProcessSetting:    p.ProcessSetting,
		ProcessPermission: p.ProcessPermission,
		CreatedBy:         userID,
	}
}

func (s *FlowService) ListProcessHistories(ctx context.Context, processID uint64, page, pageSize int, userID uint64) ([]dto.ProcessHistoryBrief, int64, error) {
	p, err := s.repo.GetProcess(ctx, processID)
	if err != nil {
		return nil, 0, err
	}
	if err := s.ensureProcessManager(ctx, userID, p); err != nil {
		return nil, 0, err
	}
	// 当前版本只在主表，历史列表排除与当前相同的版本号
	list, total, err := s.repo.ListProcessHistories(ctx, processID, p.ProcessVersion, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.ProcessHistoryBrief, 0, len(list))
	for i := range list {
		h := list[i]
		out = append(out, dto.ProcessHistoryBrief{
			HistoryID:      h.ID,
			ProcessID:      h.ProcessID,
			ProcessName:    h.ProcessName,
			ProcessIcon:    h.ProcessIcon,
			ProcessVersion: h.ProcessVersion,
			Remark:         h.Remark,
			CreatedAt:      h.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return out, total, nil
}

func (s *FlowService) GetProcessHistory(ctx context.Context, processID, historyID uint64, userID uint64) (*dto.ProcessHistoryDetail, error) {
	p, err := s.repo.GetProcess(ctx, processID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureProcessManager(ctx, userID, p); err != nil {
		return nil, err
	}
	h, err := s.repo.GetProcessHistory(ctx, historyID)
	if err != nil {
		return nil, err
	}
	if h.ProcessID != processID {
		return nil, ErrHistoryMismatch
	}
	return toHistoryDetail(p, h), nil
}

func (s *FlowService) CheckoutProcessHistory(ctx context.Context, processID, historyID uint64, userID uint64) (*dto.ProcessDetail, error) {
	p, err := s.repo.GetProcess(ctx, processID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureProcessManager(ctx, userID, p); err != nil {
		return nil, err
	}
	h, err := s.repo.GetProcessHistory(ctx, historyID)
	if err != nil {
		return nil, err
	}
	if h.ProcessID != processID {
		return nil, ErrHistoryMismatch
	}
	if h.ProcessVersion == p.ProcessVersion {
		return nil, ErrHistoryMismatch
	}

	prevSnap := snapshotHistory(p, userID)

	p.ProcessName = h.ProcessName
	p.ProcessIcon = h.ProcessIcon
	// 类型不随历史覆盖，避免破坏创建后锁定约定；内容与配置迁出
	p.Remark = h.Remark
	p.ModelContent = h.ModelContent
	p.ProcessForm = h.ProcessForm
	p.ProcessSetting = h.ProcessSetting
	p.ProcessPermission = h.ProcessPermission
	p.UpdatedBy = userID
	p.ProcessVersion = p.ProcessVersion + 1

	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		if e := s.repo.EnsureProcessHistoryTx(tx, prevSnap); e != nil {
			return e
		}
		return s.repo.UpdateProcessTx(tx, p)
	})
	if err != nil {
		return nil, err
	}
	return toDetail(p), nil
}

func toHistoryDetail(p *model.FlowProcess, h *model.FlowProcessHistory) *dto.ProcessHistoryDetail {
	icon, color := decodeIcon(h.ProcessIcon)
	return &dto.ProcessHistoryDetail{
		HistoryID:             h.ID,
		ProcessID:             h.ProcessID,
		CategoryID:            p.CategoryID,
		ProcessKey:            p.ProcessKey,
		ProcessName:           h.ProcessName,
		ProcessIcon:           icon,
		ProcessBgcolor:        color,
		ProcessType:           p.ProcessType,
		ProcessVersion:        h.ProcessVersion,
		ProcessState:          p.ProcessState,
		Remark:                h.Remark,
		ModelContent:          unmarshalMap(h.ModelContent),
		ProcessForm:           unmarshalMap(h.ProcessForm),
		ProcessSetting:        unmarshalMap(h.ProcessSetting),
		ProcessPermissionList: unmarshalMapSlice(h.ProcessPermission),
		CreatedAt:             h.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toBrief(p *model.FlowProcess) dto.ProcessBrief {
	return dto.ProcessBrief{
		ProcessID:      p.ID,
		CategoryID:     p.CategoryID,
		ProcessKey:     p.ProcessKey,
		ProcessName:    p.ProcessName,
		ProcessIcon:    p.ProcessIcon,
		ProcessType:    p.ProcessType,
		ProcessVersion: p.ProcessVersion,
		ProcessState:   p.ProcessState,
		Remark:         p.Remark,
		CreatedAt:      p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toDetail(p *model.FlowProcess) *dto.ProcessDetail {
	icon, color := decodeIcon(p.ProcessIcon)
	return &dto.ProcessDetail{
		ProcessID:             p.ID,
		CategoryID:            p.CategoryID,
		ProcessKey:            p.ProcessKey,
		ProcessName:           p.ProcessName,
		ProcessIcon:           icon,
		ProcessBgcolor:        color,
		ProcessType:           p.ProcessType,
		ProcessVersion:        p.ProcessVersion,
		ProcessState:          p.ProcessState,
		Remark:                p.Remark,
		ModelContent:          unmarshalMap(p.ModelContent),
		ProcessForm:           unmarshalMap(p.ProcessForm),
		ProcessSetting:        unmarshalMap(p.ProcessSetting),
		ProcessPermissionList: unmarshalMapSlice(p.ProcessPermission),
		CreatedBy:             p.CreatedBy,
		CreatedAt:             p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:             p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func marshalJSON(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}
	if s, ok := v.(string); ok {
		return s, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func unmarshalMap(s string) map[string]interface{} {
	out := map[string]interface{}{}
	if strings.TrimSpace(s) == "" {
		return out
	}
	_ = json.Unmarshal([]byte(s), &out)
	return out
}

func unmarshalMapSlice(s string) []map[string]interface{} {
	out := []map[string]interface{}{}
	if strings.TrimSpace(s) == "" {
		return out
	}
	_ = json.Unmarshal([]byte(s), &out)
	return out
}

func encodeIcon(icon, color string) string {
	icon = strings.TrimSpace(icon)
	if icon == "" {
		return ""
	}
	// 已是 JSON
	if strings.HasPrefix(icon, "{") {
		return icon
	}
	if color == "" {
		color = "rgba(30, 144, 255, 1)"
	}
	b, _ := json.Marshal(map[string]string{"icon": icon, "color": color})
	return string(b)
}

func decodeIcon(raw string) (icon, color string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "rgba(30, 144, 255, 1)"
	}
	if strings.HasPrefix(raw, "{") {
		var m map[string]string
		if json.Unmarshal([]byte(raw), &m) == nil {
			return m["icon"], m["color"]
		}
	}
	return raw, "rgba(30, 144, 255, 1)"
}

func validateModelContent(v interface{}) error {
	if v == nil {
		return ErrInvalidModel
	}
	var m map[string]interface{}
	switch t := v.(type) {
	case map[string]interface{}:
		m = t
	case string:
		if err := json.Unmarshal([]byte(t), &m); err != nil {
			return ErrInvalidModel
		}
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ErrInvalidModel
		}
		if err := json.Unmarshal(b, &m); err != nil {
			return ErrInvalidModel
		}
	}
	if _, ok := m["nodeConfig"]; !ok {
		return ErrInvalidModel
	}
	return nil
}

// CheckProcess 启用前配置体检（不落库、不启停）
func (s *FlowService) CheckProcess(ctx context.Context, id uint64, userID uint64) (*dto.CheckProcessResult, error) {
	p, err := s.repo.GetProcess(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureProcessManager(ctx, userID, p); err != nil {
		return nil, err
	}
	issues, err := collectEnableIssues(p)
	if err != nil {
		return &dto.CheckProcessResult{Ok: false, Message: err.Error()}, nil
	}
	if len(issues) > 0 {
		return &dto.CheckProcessResult{
			Ok:      false,
			Message: strings.Join(issues, "；"),
			Issues:  issues,
		}, nil
	}
	return &dto.CheckProcessResult{Ok: true, Message: "配置检查通过，可以启用"}, nil
}

// validateProcessForEnable 启用前检查关键节点配置，避免运行期才失败
func validateProcessForEnable(p *model.FlowProcess) error {
	issues, err := collectEnableIssues(p)
	if err != nil {
		return err
	}
	if len(issues) == 0 {
		return nil
	}
	const maxShow = 5
	show := issues
	if len(show) > maxShow {
		show = append(show[:maxShow], fmt.Sprintf("…另有 %d 项", len(issues)-maxShow))
	}
	return fmt.Errorf("%w: %s", ErrInvalidModel, strings.Join(show, "；"))
}

func collectEnableIssues(p *model.FlowProcess) ([]string, error) {
	if p == nil {
		return nil, ErrInvalidModel
	}
	mc, err := engine.ParseModel(p.ModelContent)
	if err != nil || mc == nil || mc.NodeConfig == nil {
		return nil, fmt.Errorf("%w: 缺少有效流程设计", ErrInvalidModel)
	}
	var issues []string
	var walk func(n *engine.Node)
	walk = func(n *engine.Node) {
		if n == nil {
			return
		}
		name := strings.TrimSpace(n.NodeName)
		if name == "" {
			name = n.NodeKey
		}
		switch n.Type {
		case engine.NodeApproval:
			switch n.SetType {
			case 1: // 指定成员
				if len(n.NodeAssigneeList) == 0 {
					issues = append(issues, fmt.Sprintf("审批节点「%s」未指定成员", name))
				}
			case 3: // 角色
				if len(n.NodeAssigneeList) == 0 {
					issues = append(issues, fmt.Sprintf("审批节点「%s」未指定角色", name))
				}
			}
		case engine.NodeSubProcess:
			if strings.TrimSpace(n.CallProcess) == "" && strings.TrimSpace(n.SubProcessValue) == "" {
				issues = append(issues, fmt.Sprintf("子流程节点「%s」未选择子流程", name))
			}
		case engine.NodeRoute:
			if len(n.RouteNodes) == 0 {
				issues = append(issues, fmt.Sprintf("路由节点「%s」未配置路由分支", name))
			} else {
				for _, r := range n.RouteNodes {
					if r == nil || strings.TrimSpace(r.NodeKey) == "" {
						issues = append(issues, fmt.Sprintf("路由节点「%s」存在空跳转目标", name))
						break
					}
				}
			}
		}
		walk(n.ChildNode)
		for _, b := range n.ConditionNodes {
			walk(b)
			if b != nil {
				walk(b.ChildNode)
			}
		}
		for _, b := range n.ParallelNodes {
			walk(b)
			if b != nil {
				walk(b.ChildNode)
			}
		}
		for _, b := range n.InclusiveNodes {
			walk(b)
			if b != nil {
				walk(b.ChildNode)
			}
		}
	}
	walk(mc.NodeConfig)
	return issues, nil
}

// ---------- 表单模板 ----------

func (s *FlowService) ListFormCategoryTree(ctx context.Context) ([]dto.FormCategoryListItem, error) {
	cats, err := s.repo.ListFormCategories(ctx)
	if err != nil {
		return nil, err
	}
	forms, err := s.repo.ListForms(ctx)
	if err != nil {
		return nil, err
	}
	bound, err := s.collectBoundFormIDs(ctx)
	if err != nil {
		return nil, err
	}
	byCat := map[uint64][]dto.FormBrief{}
	for i := range forms {
		b := toFormBrief(&forms[i], bound[forms[i].ID])
		byCat[forms[i].CategoryID] = append(byCat[forms[i].CategoryID], b)
	}
	out := make([]dto.FormCategoryListItem, 0, len(cats))
	for _, c := range cats {
		list := byCat[c.ID]
		if list == nil {
			list = []dto.FormBrief{}
		}
		out = append(out, dto.FormCategoryListItem{
			CategoryID:     c.ID,
			CategoryName:   c.Name,
			CategoryRemark: c.Remark,
			CategorySort:   c.Sort,
			FormList:       list,
		})
	}
	return out, nil
}

func (s *FlowService) ListFormCategoryOptions(ctx context.Context) ([]dto.FormCategoryOption, error) {
	cats, err := s.repo.ListFormCategories(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.FormCategoryOption, 0, len(cats))
	for _, c := range cats {
		out = append(out, dto.FormCategoryOption{ID: c.ID, Name: c.Name})
	}
	return out, nil
}

func (s *FlowService) CreateFormCategory(ctx context.Context, req *dto.CreateFormCategoryRequest) (*model.FlowFormCategory, error) {
	c := &model.FlowFormCategory{Name: strings.TrimSpace(req.Name), Sort: req.Sort, Remark: req.Remark}
	if err := s.repo.CreateFormCategory(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *FlowService) UpdateFormCategory(ctx context.Context, req *dto.UpdateFormCategoryRequest) error {
	c, err := s.repo.GetFormCategory(ctx, req.ID)
	if err != nil {
		return err
	}
	c.Name = strings.TrimSpace(req.Name)
	c.Sort = req.Sort
	c.Remark = req.Remark
	return s.repo.UpdateFormCategory(ctx, c)
}

func (s *FlowService) DeleteFormCategory(ctx context.Context, id uint64) error {
	n, err := s.repo.CountFormByCategory(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrCategoryHasForm
	}
	return s.repo.DeleteFormCategory(ctx, id)
}

func (s *FlowService) ListFormOptions(ctx context.Context) ([]dto.FormOption, error) {
	list, err := s.repo.ListEnabledForms(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.FormOption, 0, len(list))
	for _, f := range list {
		out = append(out, dto.FormOption{
			ID: f.ID, Name: f.Name, Code: f.Code, FormType: f.FormType, PcURL: f.PcURL,
		})
	}
	return out, nil
}

func (s *FlowService) GetForm(ctx context.Context, id uint64) (*dto.FormDetail, error) {
	f, err := s.repo.GetForm(ctx, id)
	if err != nil {
		return nil, err
	}
	bound, err := s.collectBoundFormIDs(ctx)
	if err != nil {
		return nil, err
	}
	return toFormDetail(f, bound[f.ID]), nil
}

func (s *FlowService) SaveForm(ctx context.Context, req *dto.SaveFormRequest, userID uint64) (*dto.FormDetail, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	formType := req.FormType
	if formType == 0 {
		formType = 1
	}
	pcURL := strings.TrimSpace(req.PcURL)
	pcURL = strings.TrimPrefix(pcURL, "/")
	pcURL = strings.TrimSuffix(pcURL, ".vue")
	if formType == 2 && pcURL == "" {
		return nil, fmt.Errorf("系统表单请填写组件路径 pcUrl，例如 business/demo/form")
	}
	if formType == 1 {
		pcURL = ""
	}

	if req.FormID != nil && *req.FormID > 0 {
		f, err := s.repo.GetForm(ctx, *req.FormID)
		if err != nil {
			return nil, err
		}
		if other, err := s.repo.GetFormByCode(ctx, code); err == nil && other.ID != f.ID {
			return nil, ErrFormCodeExists
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		f.CategoryID = req.CategoryID
		f.Name = name
		f.Code = code
		f.FormType = formType
		f.PcURL = pcURL
		f.Status = status
		f.Sort = req.Sort
		f.Remark = req.Remark
		f.UpdatedBy = userID
		if err := s.repo.UpdateForm(ctx, f); err != nil {
			return nil, err
		}
		return s.GetForm(ctx, f.ID)
	}

	if _, err := s.repo.GetFormByCode(ctx, code); err == nil {
		return nil, ErrFormCodeExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	f := &model.FlowForm{
		CategoryID: req.CategoryID,
		Name:       name,
		Code:       code,
		FormType:   formType,
		PcURL:      pcURL,
		Status:     status,
		Sort:       req.Sort,
		Remark:     req.Remark,
		FormSchema: `{}`,
		CreatedBy:  userID,
		UpdatedBy:  userID,
	}
	if err := s.repo.CreateForm(ctx, f); err != nil {
		return nil, err
	}
	return s.GetForm(ctx, f.ID)
}

func (s *FlowService) UpdateFormState(ctx context.Context, id uint64, status int8) error {
	f, err := s.repo.GetForm(ctx, id)
	if err != nil {
		return err
	}
	f.Status = status
	return s.repo.UpdateForm(ctx, f)
}

func (s *FlowService) SaveFormSchema(ctx context.Context, id uint64, schema interface{}, userID uint64) (*dto.FormDetail, error) {
	f, err := s.repo.GetForm(ctx, id)
	if err != nil {
		return nil, err
	}
	str, err := marshalJSON(schema)
	if err != nil {
		return nil, err
	}
	if str == "" || str == "null" {
		str = `{}`
	}
	f.FormSchema = str
	f.UpdatedBy = userID
	if err := s.repo.UpdateForm(ctx, f); err != nil {
		return nil, err
	}
	return s.GetForm(ctx, f.ID)
}

func (s *FlowService) DeleteForm(ctx context.Context, id uint64) error {
	bound, err := s.collectBoundFormIDs(ctx)
	if err != nil {
		return err
	}
	if bound[id] {
		return ErrFormBound
	}
	return s.repo.DeleteForm(ctx, id)
}

func (s *FlowService) collectBoundFormIDs(ctx context.Context) (map[uint64]bool, error) {
	list, err := s.repo.ListProcessFormBindSources(ctx)
	if err != nil {
		return nil, err
	}
	out := map[uint64]bool{}
	for _, p := range list {
		if p.ModelContent != "" {
			for _, part := range extractActionURLs(p.ModelContent) {
				if id, ok := parseFormActionID(part); ok {
					out[id] = true
				}
			}
		}
		if id, ok := parseBindFormID(p.ProcessSetting); ok {
			out[id] = true
		}
	}
	return out, nil
}

func parseBindFormID(settingJSON string) (uint64, bool) {
	if settingJSON == "" || settingJSON == "null" {
		return 0, false
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(settingJSON), &m); err != nil {
		return 0, false
	}
	switch v := m["bindFormId"].(type) {
	case float64:
		if v > 0 {
			return uint64(v), true
		}
	case json.Number:
		n, err := v.Int64()
		if err == nil && n > 0 {
			return uint64(n), true
		}
	}
	return 0, false
}

func extractActionURLs(modelJSON string) []string {
	var urls []string
	var walk func(v interface{})
	walk = func(v interface{}) {
		switch t := v.(type) {
		case map[string]interface{}:
			if u, ok := t["actionUrl"].(string); ok && u != "" {
				urls = append(urls, u)
			}
			for _, child := range t {
				walk(child)
			}
		case []interface{}:
			for _, child := range t {
				walk(child)
			}
		}
	}
	var root interface{}
	if err := json.Unmarshal([]byte(modelJSON), &root); err != nil {
		return nil
	}
	walk(root)
	return urls
}

func parseFormActionID(actionURL string) (uint64, bool) {
	actionURL = strings.TrimSpace(actionURL)
	if actionURL == "" {
		return 0, false
	}
	idx := strings.Index(actionURL, ":")
	idPart := actionURL
	if idx >= 0 {
		idPart = actionURL[:idx]
	}
	id, err := strconv.ParseUint(idPart, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return id, true
}

func toFormBrief(f *model.FlowForm, bound bool) dto.FormBrief {
	return dto.FormBrief{
		FormID:     f.ID,
		CategoryID: f.CategoryID,
		Name:       f.Name,
		Code:       f.Code,
		FormType:   f.FormType,
		PcURL:      f.PcURL,
		Status:     f.Status,
		Sort:       f.Sort,
		Remark:     f.Remark,
		Bound:      bound,
		CreatedAt:  f.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  f.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toFormDetail(f *model.FlowForm, bound bool) *dto.FormDetail {
	return &dto.FormDetail{
		FormID:     f.ID,
		CategoryID: f.CategoryID,
		Name:       f.Name,
		Code:       f.Code,
		FormType:   f.FormType,
		PcURL:      f.PcURL,
		Status:     f.Status,
		Sort:       f.Sort,
		Remark:     f.Remark,
		FormSchema: unmarshalMap(f.FormSchema),
		Bound:      bound,
		CreatedAt:  f.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  f.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
