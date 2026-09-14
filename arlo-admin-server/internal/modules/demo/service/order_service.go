package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"arlo-admin/internal/modules/demo/dto"
	"arlo-admin/internal/modules/demo/model"
	"arlo-admin/internal/modules/demo/repository"
	flowdto "arlo-admin/internal/modules/flow/dto"
	flowmodel "arlo-admin/internal/modules/flow/model"
	flowrepo "arlo-admin/internal/modules/flow/repository"
	flowsvc "arlo-admin/internal/modules/flow/service"
	"arlo-admin/pkg/response"

	"gorm.io/gorm"
)

type OrderService struct {
	repo     *repository.OrderRepository
	flow     *flowsvc.FlowService
	flowRepo *flowrepo.FlowRepository
}

func NewOrderService(repo *repository.OrderRepository, flow *flowsvc.FlowService, flowRepo *flowrepo.FlowRepository) *OrderService {
	return &OrderService{repo: repo, flow: flow, flowRepo: flowRepo}
}

// ListProcessOptions 已启用的业务流程（创建时绑定，不按发起人过滤，便于先绑后控）
func (s *OrderService) ListProcessOptions(ctx context.Context) ([]dto.ProcessOption, error) {
	list, err := s.flowRepo.ListLaunchProcesses(ctx, "")
	if err != nil {
		return nil, err
	}
	out := make([]dto.ProcessOption, 0, len(list))
	for _, p := range list {
		if p.ProcessType != "business" {
			continue
		}
		out = append(out, dto.ProcessOption{
			ProcessID:   p.ID,
			ProcessKey:  p.ProcessKey,
			ProcessName: p.ProcessName,
			ProcessType: p.ProcessType,
			Remark:      p.Remark,
		})
	}
	return out, nil
}

// GetProcessPreview 未发起时按单据预览已绑定流程（不校验发起人权限）
func (s *OrderService) GetProcessPreview(ctx context.Context, orderID uint64) (*dto.ProcessPreview, error) {
	o, err := s.repo.Get(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("单据不存在")
		}
		return nil, err
	}
	if o.ProcessID == 0 {
		return nil, fmt.Errorf("单据未绑定业务流程")
	}
	p, err := s.flowRepo.GetProcess(ctx, o.ProcessID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("流程不存在")
		}
		return nil, err
	}
	if p.ProcessType != "business" {
		return nil, fmt.Errorf("仅支持业务流程预览")
	}
	mc := map[string]interface{}{}
	if strings.TrimSpace(p.ModelContent) != "" {
		if e := json.Unmarshal([]byte(p.ModelContent), &mc); e != nil {
			return nil, fmt.Errorf("流程模型解析失败")
		}
	}
	return &dto.ProcessPreview{
		ProcessID:    p.ID,
		ProcessKey:   p.ProcessKey,
		ProcessName:  p.ProcessName,
		ModelContent: mc,
	}, nil
}

func (s *OrderService) List(ctx context.Context, q *dto.ListQuery, userID uint64) (*response.PageData, error) {
	list, total, err := s.repo.List(ctx, strings.TrimSpace(q.Title), q.Status, q.Page, q.PageSize)
	if err != nil {
		return nil, err
	}
	launchable := map[uint64]struct{}{}
	if procs, e := s.flow.ListLaunchProcesses(ctx, "", userID); e == nil {
		for _, p := range procs {
			if p.ProcessType == "business" {
				launchable[p.ProcessID] = struct{}{}
			}
		}
	}
	nameCache := map[uint64]string{}
	items := make([]dto.OrderItem, 0, len(list))
	for i := range list {
		o := &list[i]
		s.syncStatusFromInstance(ctx, o)
		item := toItem(o)
		if o.ProcessID > 0 {
			if _, ok := launchable[o.ProcessID]; ok {
				item.CanLaunch = true
			}
			if n, ok := nameCache[o.ProcessID]; ok {
				item.ProcessName = n
			} else if p, e := s.flowRepo.GetProcess(ctx, o.ProcessID); e == nil && p != nil {
				nameCache[o.ProcessID] = p.ProcessName
				item.ProcessName = p.ProcessName
			}
		}
		items = append(items, item)
	}
	return &response.PageData{List: items, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

func (s *OrderService) Create(ctx context.Context, req *dto.SaveRequest, userID uint64, userName string) (uint64, error) {
	p, err := s.flowRepo.GetProcess(ctx, req.ProcessID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("流程不存在")
		}
		return 0, err
	}
	if p.ProcessType != "business" {
		return 0, fmt.Errorf("请选择业务流程")
	}
	if p.ProcessState != 1 {
		return 0, fmt.Errorf("流程未启用")
	}

	o := &model.DemoPurchaseOrder{
		Title:      strings.TrimSpace(req.Title),
		Content:    strings.TrimSpace(req.Content),
		Status:     model.OrderPending,
		ProcessID:  p.ID,
		ProcessKey: p.ProcessKey,
		CreateID:   userID,
		CreateBy:   userName,
	}
	if err := s.repo.Create(ctx, o); err != nil {
		return 0, err
	}
	return o.ID, nil
}

func (s *OrderService) Delete(ctx context.Context, ids []uint64) error {
	for _, id := range ids {
		o, err := s.repo.Get(ctx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}
		if o.Status == model.OrderApproving {
			return fmt.Errorf("审批中的单据不可删除")
		}
	}
	return s.repo.Delete(ctx, ids)
}

// GetLaunchForm 按单据已绑定流程取发起表单（会校验设计器发起人权限）
func (s *OrderService) GetLaunchForm(ctx context.Context, orderID, userID uint64) (*flowdto.LaunchFormDetail, error) {
	o, err := s.repo.Get(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("单据不存在")
		}
		return nil, err
	}
	if o.ProcessID == 0 {
		return nil, fmt.Errorf("单据未绑定业务流程")
	}
	p, err := s.flowRepo.GetProcess(ctx, o.ProcessID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("流程不存在")
		}
		return nil, err
	}
	if p.ProcessType != "business" {
		return nil, fmt.Errorf("演示页仅支持业务流程")
	}
	return s.flow.GetLaunchForm(ctx, p.ID, userID)
}

// Launch 使用单据已绑定流程发起
func (s *OrderService) Launch(ctx context.Context, id uint64, req *dto.LaunchRequest, userID uint64, userName string) (uint64, error) {
	o, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("单据不存在")
		}
		return 0, err
	}
	if o.Status != model.OrderPending {
		return 0, fmt.Errorf("仅「待审批」状态可发起")
	}
	if o.ProcessID == 0 {
		return 0, fmt.Errorf("单据未绑定业务流程")
	}

	p, err := s.flowRepo.GetProcess(ctx, o.ProcessID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("流程不存在")
		}
		return 0, err
	}
	if p.ProcessType != "business" {
		return 0, fmt.Errorf("演示页仅支持业务流程")
	}

	formData := map[string]interface{}{
		"title":   o.Title,
		"content": o.Content,
		"remark":  o.Content,
	}
	for k, v := range req.FormData {
		formData[k] = v
	}
	if t, ok := formData["title"].(string); ok && strings.TrimSpace(t) != "" {
		o.Title = strings.TrimSpace(t)
	}
	if c, ok := formData["content"].(string); ok {
		o.Content = strings.TrimSpace(c)
	} else if r, ok := formData["remark"].(string); ok {
		o.Content = strings.TrimSpace(r)
	}

	launchReq := &flowdto.LaunchRequest{
		ProcessID:     p.ID,
		FormData:      formData,
		NodeAssignees: req.NodeAssignees,
		NodeCCUsers:   req.NodeCCUsers,
	}
	instanceID, err := s.flow.Launch(ctx, launchReq, userID, userName)
	if err != nil {
		return 0, err
	}

	o.InstanceID = instanceID
	o.Status = model.OrderApproving
	o.ProcessID = p.ID
	o.ProcessKey = p.ProcessKey
	if err := s.repo.Update(ctx, o); err != nil {
		return 0, err
	}
	return instanceID, nil
}

func (s *OrderService) syncStatusFromInstance(ctx context.Context, o *model.DemoPurchaseOrder) {
	if o == nil || o.InstanceID == 0 {
		return
	}
	inst, err := s.flowRepo.GetInstance(ctx, o.InstanceID)
	if err != nil || inst == nil {
		return
	}
	next := mapInstanceState(inst.InstanceState)
	if next == o.Status {
		return
	}
	o.Status = next
	_ = s.repo.Update(ctx, o)
}

func mapInstanceState(state int8) int8 {
	switch state {
	case flowmodel.InstActive, flowmodel.InstDraft:
		return model.OrderApproving
	case flowmodel.InstComplete:
		return model.OrderPassed
	case flowmodel.InstReject, flowmodel.InstRevoke, flowmodel.InstTerminate, flowmodel.InstTimeout:
		return model.OrderRejected
	default:
		return model.OrderApproving
	}
}

func toItem(o *model.DemoPurchaseOrder) dto.OrderItem {
	return dto.OrderItem{
		ID:         o.ID,
		Title:      o.Title,
		Content:    o.Content,
		Status:     o.Status,
		InstanceID: o.InstanceID,
		ProcessID:  o.ProcessID,
		ProcessKey: o.ProcessKey,
		CreateBy:   o.CreateBy,
		CreatedAt:  o.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
