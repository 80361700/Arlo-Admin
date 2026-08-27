package datascope

import (
	"gorm.io/gorm"
)

// Scope 数据范围常量
const (
	ScopeAll          = 1 // 全部数据权限
	ScopeCustom       = 2 // 自定义部门
	ScopeDeptAndChild = 3 // 本部门及以下
	ScopeDept         = 4 // 本部门
	ScopeSelf         = 5 // 仅本人
)

// Provider 数据权限范围注入器
// 由 Service 层根据用户角色（取最宽松 data_scope）构建，传入查询链。
//
// 已接入（业务数据）：用户列表/导出、文件、操作/登录日志、公告、站内信（发送端）。
// 不接入（全局主数据，无部门/归属字段或全员共享）：字典、岗位、部门、角色、菜单、系统配置、会员。
type Provider struct {
	Scope   int8     // 当前生效的数据范围
	DeptIDs []uint64 // 可见部门ID列表（Scope=Custom 或 Scope=DeptAndChild 时使用）
	UserID  uint64   // 当前用户ID（Scope=Self 时使用）
}

// ApplyOptions 数据权限查询选项
// 不同业务表可能用不同的列名来标识部门和创建者
type ApplyOptions struct {
	DeptColumn   string // 部门列名，默认 "dept_id"
	CreatorColumn string // 创建者列名，默认 "created_by"（ScopeSelf 时使用）
}

// Apply 将数据权限 WHERE 条件注入 GORM 查询（使用默认列名 dept_id / created_by）
//
// 调用方式: provider.Apply(db).Find(&records)
func (p *Provider) Apply(tx *gorm.DB) *gorm.DB {
	return p.ApplyWith(tx, ApplyOptions{})
}

// ApplyWith 将数据权限 WHERE 条件注入 GORM 查询（使用自定义列名）
//
// 调用方式: provider.ApplyWith(db, datascope.ApplyOptions{DeptColumn: "dept_id", CreatorColumn: "created_by"}).Find(&records)
func (p *Provider) ApplyWith(tx *gorm.DB, opts ApplyOptions) *gorm.DB {
	deptCol := opts.DeptColumn
	if deptCol == "" {
		deptCol = "dept_id"
	}
	creatorCol := opts.CreatorColumn
	if creatorCol == "" {
		creatorCol = "created_by"
	}

	switch p.Scope {
	case ScopeAll:
		// 全部数据 — 不加任何过滤
		return tx

	case ScopeCustom:
		// 自定义部门
		if len(p.DeptIDs) == 0 {
			return tx.Where("1 = 0") // 安全：未配置部门时查不出任何数据
		}
		return tx.Where(deptCol+" IN ?", p.DeptIDs)

	case ScopeDeptAndChild:
		// 本部门及以下
		if len(p.DeptIDs) == 0 {
			return tx.Where("1 = 0")
		}
		return tx.Where(deptCol+" IN ?", p.DeptIDs)

	case ScopeDept:
		// 本部门 — DeptIDs[0] 是用户所属部门
		if len(p.DeptIDs) == 0 {
			return tx.Where("1 = 0")
		}
		return tx.Where(deptCol+" = ?", p.DeptIDs[0])

	case ScopeSelf:
		// 仅本人 — 按创建者过滤
		return tx.Where(creatorCol+" = ?", p.UserID)

	default:
		// 未知范围 — 安全兜底，仅本人
		return tx.Where(creatorCol+" = ?", p.UserID)
	}
}
