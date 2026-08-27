package datascope

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// RoleScope 构建数据权限所需的角色信息
type RoleScope struct {
	ID        uint64 `gorm:"column:id"`
	DataScope int8   `gorm:"column:data_scope"`
}

// Source 数据权限构建所需的数据源（由各模块注入仓储实现）
type Source interface {
	FindUserRoles(ctx context.Context, userID uint64) ([]RoleScope, error)
	FindRoleDeptIDs(ctx context.Context, roleID uint64) ([]uint64, error)
	FindUserDeptID(ctx context.Context, userID uint64) (uint64, error)
	FindDescendantDeptIDs(ctx context.Context, deptID uint64) ([]uint64, error)
}

// Build 根据用户角色 data_scope 构建 Provider（多角色取最宽松）
func Build(ctx context.Context, src Source, userID uint64) (*Provider, error) {
	roles, err := src.FindUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return &Provider{Scope: ScopeSelf, UserID: userID}, nil
	}

	minScope := int8(ScopeSelf)
	for _, r := range roles {
		ds := r.DataScope
		if ds == 0 {
			ds = ScopeAll
		}
		if ds < minScope {
			minScope = ds
		}
	}

	switch minScope {
	case ScopeAll:
		return &Provider{Scope: ScopeAll}, nil

	case ScopeCustom:
		var deptIDs []uint64
		seen := make(map[uint64]bool)
		for _, r := range roles {
			if r.DataScope != ScopeCustom {
				continue
			}
			ids, _ := src.FindRoleDeptIDs(ctx, r.ID)
			for _, id := range ids {
				if !seen[id] {
					seen[id] = true
					deptIDs = append(deptIDs, id)
				}
			}
		}
		return &Provider{Scope: ScopeCustom, DeptIDs: deptIDs}, nil

	case ScopeDeptAndChild:
		deptID, err := src.FindUserDeptID(ctx, userID)
		if err != nil || deptID == 0 {
			return &Provider{Scope: ScopeSelf, UserID: userID}, nil
		}
		ids, _ := src.FindDescendantDeptIDs(ctx, deptID)
		return &Provider{Scope: ScopeDeptAndChild, DeptIDs: ids}, nil

	case ScopeDept:
		deptID, err := src.FindUserDeptID(ctx, userID)
		if err != nil || deptID == 0 {
			return &Provider{Scope: ScopeSelf, UserID: userID}, nil
		}
		return &Provider{Scope: ScopeDept, DeptIDs: []uint64{deptID}}, nil

	case ScopeSelf:
		return &Provider{Scope: ScopeSelf, UserID: userID}, nil

	default:
		return &Provider{Scope: ScopeSelf, UserID: userID}, nil
	}
}

// ApplyByOwner 按「归属用户」过滤（表无 dept_id，仅有 user_id / uploader_id 等）
// 部门类范围通过子查询关联 sys_user.dept_id
func (p *Provider) ApplyByOwner(tx *gorm.DB, ownerColumn string) *gorm.DB {
	if ownerColumn == "" {
		ownerColumn = "user_id"
	}
	switch p.Scope {
	case ScopeAll:
		return tx
	case ScopeSelf:
		return tx.Where(ownerColumn+" = ?", p.UserID)
	case ScopeCustom, ScopeDeptAndChild, ScopeDept:
		if len(p.DeptIDs) == 0 {
			return tx.Where("1 = 0")
		}
		return tx.Where(fmt.Sprintf(
			"%s IN (SELECT id FROM sys_user WHERE dept_id IN ? AND deleted_at IS NULL)",
			ownerColumn,
		), p.DeptIDs)
	default:
		return tx.Where(ownerColumn+" = ?", p.UserID)
	}
}

// ApplyByUsername 按用户名归属过滤（登录日志等无 user_id 的表）
func (p *Provider) ApplyByUsername(tx *gorm.DB, usernameColumn string) *gorm.DB {
	if usernameColumn == "" {
		usernameColumn = "username"
	}
	switch p.Scope {
	case ScopeAll:
		return tx
	case ScopeSelf:
		return tx.Where(fmt.Sprintf(
			"%s IN (SELECT username FROM sys_user WHERE id = ? AND deleted_at IS NULL)",
			usernameColumn,
		), p.UserID)
	case ScopeCustom, ScopeDeptAndChild, ScopeDept:
		if len(p.DeptIDs) == 0 {
			return tx.Where("1 = 0")
		}
		return tx.Where(fmt.Sprintf(
			"%s IN (SELECT username FROM sys_user WHERE dept_id IN ? AND deleted_at IS NULL)",
			usernameColumn,
		), p.DeptIDs)
	default:
		return tx.Where(fmt.Sprintf(
			"%s IN (SELECT username FROM sys_user WHERE id = ? AND deleted_at IS NULL)",
			usernameColumn,
		), p.UserID)
	}
}
