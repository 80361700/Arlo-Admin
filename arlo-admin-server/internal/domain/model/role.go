package model

import (
	"time"

	"gorm.io/gorm"
)

// Role 系统角色模型，对应 sys_role 表
type Role struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:32;not null" json:"name"`
	Code      string         `gorm:"uniqueIndex;size:32;not null" json:"code"`
	Sort      int            `gorm:"not null;default:0" json:"sort"`
	Status    int8           `gorm:"not null;default:1" json:"status"`
	Remark    string         `gorm:"size:255;not null" json:"remark"`
	DataScope int8           `gorm:"not null;default:1" json:"dataScope"` // 数据范围: 1=全部 2=自定义 3=本部门及以下 4=本部门 5=仅本人
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Role) TableName() string { return "sys_role" }

// RoleMenu 角色菜单关联模型，对应 sys_role_menu 表
type RoleMenu struct {
	ID     uint64 `gorm:"primaryKey" json:"id"`
	RoleID uint64 `gorm:"uniqueIndex:uk_role_menu;not null" json:"roleId"`
	MenuID uint64 `gorm:"uniqueIndex:uk_role_menu;not null" json:"menuId"`
}

func (RoleMenu) TableName() string { return "sys_role_menu" }

// RoleDept 角色部门关联模型（data_scope=2 自定义数据时使用），对应 sys_role_dept 表
type RoleDept struct {
	ID     uint64 `gorm:"primaryKey" json:"id"`
	RoleID uint64 `gorm:"uniqueIndex:uk_role_dept;not null" json:"roleId"`
	DeptID uint64 `gorm:"uniqueIndex:uk_role_dept;not null" json:"deptId"`
}

func (RoleDept) TableName() string { return "sys_role_dept" }
