package model

import "time"

// 演示单据状态
const (
	OrderPending   int8 = 0 // 待审批
	OrderApproving int8 = 1 // 审批中
	OrderPassed    int8 = 2 // 已通过
	OrderRejected  int8 = 3 // 已拒绝
)

// DemoPurchaseOrder 演示采购单（创建时绑定业务流程）
type DemoPurchaseOrder struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Title      string    `gorm:"size:128;not null;default:''" json:"title"`
	Content    string    `gorm:"size:512;not null;default:''" json:"content"`
	Status     int8      `gorm:"not null;default:0;index" json:"status"`
	InstanceID uint64    `gorm:"not null;default:0;index" json:"instanceId"`
	ProcessID  uint64    `gorm:"not null;default:0" json:"processId"`
	ProcessKey string    `gorm:"size:64;not null;default:''" json:"processKey"`
	CreateID   uint64    `gorm:"not null;default:0;index" json:"createId"`
	CreateBy   string    `gorm:"size:64;not null;default:''" json:"createBy"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (DemoPurchaseOrder) TableName() string { return "demo_purchase_order" }
