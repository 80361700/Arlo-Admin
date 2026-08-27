package model

import (
	"time"

	"gorm.io/gorm"
)

// User 系统用户模型，对应 sys_user 表
type User struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:32;not null" json:"username"`
	Password  string         `gorm:"size:128;not null" json:"-"`
	Nickname  string         `gorm:"size:32;not null" json:"nickname"`
	Avatar    string         `gorm:"size:255;not null" json:"avatar"`
	Email     string         `gorm:"size:64;not null" json:"email"`
	Phone     string         `gorm:"size:20;not null" json:"phone"`
	Gender    int8           `gorm:"not null;default:0" json:"gender"`
	DeptID    uint64         `gorm:"not null;default:0" json:"deptId"`
	Status    int8           `gorm:"not null;default:1;index" json:"status"`
	Remark        string         `gorm:"size:255;not null" json:"remark"`
	LastLogin     *time.Time     `json:"lastLogin"`
	PwdUpdatedAt  *time.Time     `json:"pwdUpdatedAt"`
	MustChangePwd int8           `gorm:"not null;default:0" json:"mustChangePwd"` // 1=强制改密
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return "sys_user" }

// IsEnabled 判断用户是否启用
func (u *User) IsEnabled() bool {
	return u.Status == 1
}

// UserRole 用户角色关联模型，对应 sys_user_role 表
type UserRole struct {
	ID     uint64 `gorm:"primaryKey" json:"id"`
	UserID uint64 `gorm:"uniqueIndex:uk_user_role;not null" json:"userId"`
	RoleID uint64 `gorm:"uniqueIndex:uk_user_role;not null" json:"roleId"`
}

func (UserRole) TableName() string { return "sys_user_role" }

// UserPost 用户岗位关联模型，对应 sys_user_post 表
type UserPost struct {
	ID     uint64 `gorm:"primaryKey" json:"id"`
	UserID uint64 `gorm:"uniqueIndex:uk_user_post;not null" json:"userId"`
	PostID uint64 `gorm:"uniqueIndex:uk_user_post;not null" json:"postId"`
}

func (UserPost) TableName() string { return "sys_user_post" }
