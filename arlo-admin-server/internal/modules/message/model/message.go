package model

import (
	"time"

	"gorm.io/gorm"
)

// Message 站内信模型，对应 sys_message 表
type Message struct {
	ID              uint64         `gorm:"primaryKey" json:"id"`
	Title           string         `gorm:"size:128;not null;default:''" json:"title"`
	Content         string         `gorm:"type:text" json:"content"`
	Type            int8           `gorm:"not null;default:1" json:"type"` // 1=系统消息, 2=通知, 3=私信
	SenderID        uint64         `gorm:"not null;default:0" json:"senderId"`
	Sender          string         `gorm:"size:32;not null;default:''" json:"sender"`
	ReceiverID      uint64         `gorm:"not null;default:0" json:"receiverId"` // 0=全部用户
	ReceiverName    string         `gorm:"-" json:"receiverName"`
	IsRead          int8           `gorm:"not null;default:0" json:"isRead"`
	ReadAt          *time.Time     `json:"readAt"`
	SenderDeleted   int8           `gorm:"not null;default:0" json:"senderDeleted"`   // 发送方已删
	ReceiverDeleted int8           `gorm:"not null;default:0" json:"receiverDeleted"` // 指定收件方已删
	CreatedAt       time.Time      `json:"createdAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	ReceiverCount   int            `gorm:"->;column:receiver_count" json:"receiverCount"`
}

func (Message) TableName() string { return "sys_message" }

// MessageRead 广播站内信的个人已读记录
type MessageRead struct {
	MessageID uint64    `gorm:"primaryKey;not null" json:"messageId"`
	UserID    uint64    `gorm:"primaryKey;not null" json:"userId"`
	ReadAt    time.Time `gorm:"not null" json:"readAt"`
}

func (MessageRead) TableName() string { return "sys_message_read" }

// MessageHide 广播站内信的个人隐藏（从「我的消息」删除）
type MessageHide struct {
	MessageID uint64    `gorm:"primaryKey;not null" json:"messageId"`
	UserID    uint64    `gorm:"primaryKey;not null" json:"userId"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
}

func (MessageHide) TableName() string { return "sys_message_hide" }
