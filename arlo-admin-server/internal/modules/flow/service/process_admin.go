package service

import (
	"context"
	"encoding/json"
	"errors"

	domainrepo "arlo-admin/internal/domain/repository"
	"arlo-admin/internal/modules/flow/model"
)

var ErrProcessForbidden = errors.New("无权管理该流程")

// parseProcessManagerIDs 从 processPermission JSON 解析管理员用户 id
func parseProcessManagerIDs(permJSON string) []uint64 {
	if permJSON == "" || permJSON == "null" || permJSON == "[]" {
		return nil
	}
	var list []map[string]interface{}
	if err := json.Unmarshal([]byte(permJSON), &list); err != nil {
		return nil
	}
	seen := map[uint64]struct{}{}
	out := make([]uint64, 0, len(list))
	for _, m := range list {
		id := jsonUint64(m["userId"])
		if id == 0 {
			id = jsonUint64(m["id"])
		}
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func jsonUint64(v interface{}) uint64 {
	switch t := v.(type) {
	case float64:
		if t > 0 {
			return uint64(t)
		}
	case int:
		if t > 0 {
			return uint64(t)
		}
	case int64:
		if t > 0 {
			return uint64(t)
		}
	case uint64:
		return t
	case json.Number:
		n, _ := t.Int64()
		if n > 0 {
			return uint64(n)
		}
	case string:
		var n uint64
		for _, c := range t {
			if c < '0' || c > '9' {
				return 0
			}
			n = n*10 + uint64(c-'0')
		}
		return n
	}
	return 0
}

func (s *FlowService) userIsSuperAdmin(ctx context.Context, userID uint64) bool {
	if userID == 0 {
		return false
	}
	roles, err := domainrepo.NewRoleRepository().FindRolesByUserID(ctx, userID)
	if err != nil {
		return false
	}
	for _, r := range roles {
		if r != nil && r.Code == "super_admin" {
			return true
		}
	}
	return false
}

// isProcessManager 流程管理员：超管 / 创建人 / 备注管理员列表中的人。
// 未配置备注管理员时，仅创建人 + 超管。
func (s *FlowService) isProcessManager(ctx context.Context, userID uint64, p *model.FlowProcess) bool {
	if userID == 0 || p == nil {
		return false
	}
	if s.userIsSuperAdmin(ctx, userID) {
		return true
	}
	if p.CreatedBy > 0 && p.CreatedBy == userID {
		return true
	}
	for _, id := range parseProcessManagerIDs(p.ProcessPermission) {
		if id == userID {
			return true
		}
	}
	return false
}

func (s *FlowService) ensureProcessManager(ctx context.Context, userID uint64, p *model.FlowProcess) error {
	if s.isProcessManager(ctx, userID, p) {
		return nil
	}
	return ErrProcessForbidden
}

func (s *FlowService) isProcessManagerByProcessID(ctx context.Context, userID, processID uint64) bool {
	if processID == 0 {
		return false
	}
	p, err := s.repo.GetProcess(ctx, processID)
	if err != nil || p == nil {
		return false
	}
	return s.isProcessManager(ctx, userID, p)
}

func (s *FlowService) isInstanceProcessManager(ctx context.Context, userID uint64, inst *model.FlowInstance) bool {
	if inst == nil {
		return false
	}
	if s.isProcessManagerByProcessID(ctx, userID, inst.ProcessID) {
		return true
	}
	if inst.ParentInstanceID > 0 {
		parent, err := s.repo.GetInstance(ctx, inst.ParentInstanceID)
		if err == nil && parent != nil {
			return s.isProcessManagerByProcessID(ctx, userID, parent.ProcessID)
		}
	}
	return false
}

// listManagedProcessIDs 当前用户可监控的流程定义 id；all=true 表示超管可看全部
func (s *FlowService) listManagedProcessIDs(ctx context.Context, userID uint64) (ids []uint64, all bool, err error) {
	if s.userIsSuperAdmin(ctx, userID) {
		return nil, true, nil
	}
	list, err := s.repo.ListProcesses(ctx, nil, "")
	if err != nil {
		return nil, false, err
	}
	out := make([]uint64, 0)
	for i := range list {
		if s.isProcessManager(ctx, userID, &list[i]) {
			out = append(out, list[i].ID)
		}
	}
	return out, false, nil
}
