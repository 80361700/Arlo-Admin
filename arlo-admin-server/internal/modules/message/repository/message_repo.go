package repository

import (
	"context"
	"fmt"
	"time"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/message/model"
	"arlo-admin/pkg/datascope"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	DeleteSideReceived = "received" // 我的消息：仅对自己隐藏
	DeleteSideSent     = "sent"     // 发送记录：仅对发送方隐藏
)

// MessageRepository 站内信数据访问层
type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository() *MessageRepository {
	return &MessageRepository{db: database.DB}
}

func (r *MessageRepository) Create(ctx context.Context, msg *model.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *MessageRepository) CreateBatch(ctx context.Context, receiverIDs []uint64, title, content string, msgType int8, senderID uint64, sender string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rid := range receiverIDs {
			msg := &model.Message{
				Title:      title,
				Content:    content,
				Type:       msgType,
				SenderID:   senderID,
				Sender:     sender,
				ReceiverID: rid,
			}
			if err := tx.Create(msg).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// batchWhere 与 ListSent 分组条件一致
func batchWhere(msg *model.Message) (string, []interface{}) {
	return "sender_id = ? AND title = ? AND IFNULL(content,'') = IFNULL(?, '') AND type = ? AND DATE_FORMAT(created_at, '%Y-%m-%d %H:%i') = DATE_FORMAT(?, '%Y-%m-%d %H:%i')",
		[]interface{}{msg.SenderID, msg.Title, msg.Content, msg.Type, msg.CreatedAt}
}

// DeleteBySide 按场景删除（互不影响对方可见性）
func (r *MessageRepository) DeleteBySide(ctx context.Context, id, userID uint64, side string, dataScopeAll bool) error {
	var msg model.Message
	if err := r.db.WithContext(ctx).First(&msg, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("消息不存在")
		}
		return err
	}

	switch side {
	case DeleteSideSent:
		return r.deleteSentSide(ctx, &msg, userID, dataScopeAll)
	case DeleteSideReceived:
		return r.deleteReceivedSide(ctx, &msg, userID)
	default:
		return fmt.Errorf("无效的删除场景")
	}
}

func (r *MessageRepository) deleteSentSide(ctx context.Context, msg *model.Message, userID uint64, dataScopeAll bool) error {
	can := msg.SenderID == userID
	if !can && msg.SenderID == 0 && dataScopeAll {
		can = true
	}
	if !can && dataScopeAll {
		// 全部数据权限：可清理发送记录列表中可见的批次
		can = true
	}
	if !can {
		return fmt.Errorf("无权删除该发送记录")
	}

	where, args := batchWhere(msg)
	result := r.db.WithContext(ctx).Model(&model.Message{}).
		Where(where, args...).
		Where("sender_deleted = 0").
		Update("sender_deleted", 1)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("发送记录不存在或已删除")
	}
	return nil
}

func (r *MessageRepository) deleteReceivedSide(ctx context.Context, msg *model.Message, userID uint64) error {
	// 广播：个人隐藏表
	if msg.ReceiverID == 0 {
		rec := model.MessageHide{MessageID: msg.ID, UserID: userID, CreatedAt: time.Now()}
		return r.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "message_id"}, {Name: "user_id"}},
			DoNothing: true,
		}).Create(&rec).Error
	}

	if msg.ReceiverID != userID {
		return fmt.Errorf("无权删除该消息")
	}

	result := r.db.WithContext(ctx).Model(&model.Message{}).
		Where("id = ? AND receiver_id = ? AND receiver_deleted = 0", msg.ID, userID).
		Update("receiver_deleted", 1)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("消息不存在或已删除")
	}
	return nil
}

// MarkAsRead 标记已读：私信改本行；广播写入个人已读回执
func (r *MessageRepository) MarkAsRead(ctx context.Context, id, userID uint64) error {
	var msg model.Message
	if err := r.db.WithContext(ctx).First(&msg, id).Error; err != nil {
		return err
	}

	now := time.Now()
	if msg.ReceiverID == 0 {
		rec := model.MessageRead{MessageID: id, UserID: userID, ReadAt: now}
		return r.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "message_id"}, {Name: "user_id"}},
			DoNothing: true,
		}).Create(&rec).Error
	}

	if msg.ReceiverID != userID {
		return gorm.ErrRecordNotFound
	}
	return r.db.WithContext(ctx).Model(&model.Message{}).
		Where("id = ? AND receiver_id = ? AND is_read = 0 AND receiver_deleted = 0", id, userID).
		Updates(map[string]interface{}{
			"is_read": 1,
			"read_at": &now,
		}).Error
}

func (r *MessageRepository) MarkAllAsRead(ctx context.Context, userID uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Message{}).
			Where("receiver_id = ? AND is_read = 0 AND receiver_deleted = 0", userID).
			Updates(map[string]interface{}{
				"is_read": 1,
				"read_at": &now,
			}).Error; err != nil {
			return err
		}

		var broadcastIDs []uint64
		if err := tx.Model(&model.Message{}).
			Where(`receiver_id = 0 AND deleted_at IS NULL
				AND id NOT IN (?)
				AND id NOT IN (?)`,
				tx.Model(&model.MessageRead{}).Select("message_id").Where("user_id = ?", userID),
				tx.Model(&model.MessageHide{}).Select("message_id").Where("user_id = ?", userID),
			).
			Pluck("id", &broadcastIDs).Error; err != nil {
			return err
		}
		if len(broadcastIDs) == 0 {
			return nil
		}
		reads := make([]model.MessageRead, 0, len(broadcastIDs))
		for _, mid := range broadcastIDs {
			reads = append(reads, model.MessageRead{MessageID: mid, UserID: userID, ReadAt: now})
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "message_id"}, {Name: "user_id"}},
			DoNothing: true,
		}).Create(&reads).Error
	})
}

func (r *MessageRepository) FindByID(ctx context.Context, id uint64) (*model.Message, error) {
	var msg model.Message
	err := r.db.WithContext(ctx).First(&msg, id).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *MessageRepository) List(ctx context.Context, userID uint64, isRead *int8, direction *int8, page, pageSize int) ([]model.Message, int64, error) {
	var msgs []model.Message
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Message{}).Where("deleted_at IS NULL")

	// direction: 0=全部 1=我收到的 2=我发送的
	if direction != nil {
		switch *direction {
		case 1:
			// 我的消息：指定给我且未删，或广播且未个人隐藏
			q = q.Where(`(
				(receiver_id = ? AND receiver_deleted = 0) OR
				(receiver_id = 0 AND NOT EXISTS (
					SELECT 1 FROM sys_message_hide h
					WHERE h.message_id = sys_message.id AND h.user_id = ?
				))
			)`, userID, userID)
		case 2:
			q = q.Where("sender_id = ? AND sender_deleted = 0", userID)
		default:
			q = q.Where(`(
				(sender_id = ? AND sender_deleted = 0) OR
				(receiver_id = ? AND receiver_deleted = 0) OR
				(receiver_id = 0 AND NOT EXISTS (
					SELECT 1 FROM sys_message_hide h
					WHERE h.message_id = sys_message.id AND h.user_id = ?
				))
			)`, userID, userID, userID)
		}
	} else {
		q = q.Where(`(
			(sender_id = ? AND sender_deleted = 0) OR
			(receiver_id = ? AND receiver_deleted = 0) OR
			(receiver_id = 0 AND NOT EXISTS (
				SELECT 1 FROM sys_message_hide h
				WHERE h.message_id = sys_message.id AND h.user_id = ?
			))
		)`, userID, userID, userID)
	}

	if isRead != nil {
		if *isRead == 1 {
			q = q.Where(`(
				(receiver_id = ? AND is_read = 1) OR
				(receiver_id = 0 AND EXISTS (
					SELECT 1 FROM sys_message_read r
					WHERE r.message_id = sys_message.id AND r.user_id = ?
				))
			)`, userID, userID)
		} else {
			q = q.Where(`(
				(receiver_id = ? AND is_read = 0) OR
				(receiver_id = 0 AND NOT EXISTS (
					SELECT 1 FROM sys_message_read r
					WHERE r.message_id = sys_message.id AND r.user_id = ?
				))
			)`, userID, userID)
		}
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&msgs).Error; err != nil {
		return nil, 0, err
	}

	if err := r.overlayBroadcastRead(ctx, userID, msgs); err != nil {
		return nil, 0, err
	}

	return msgs, total, nil
}

func (r *MessageRepository) overlayBroadcastRead(ctx context.Context, userID uint64, msgs []model.Message) error {
	ids := make([]uint64, 0)
	for _, m := range msgs {
		if m.ReceiverID == 0 {
			ids = append(ids, m.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var reads []model.MessageRead
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND message_id IN ?", userID, ids).
		Find(&reads).Error; err != nil {
		return err
	}
	readMap := make(map[uint64]model.MessageRead, len(reads))
	for _, rd := range reads {
		readMap[rd.MessageID] = rd
	}
	for i := range msgs {
		if msgs[i].ReceiverID != 0 {
			continue
		}
		if rd, ok := readMap[msgs[i].ID]; ok {
			msgs[i].IsRead = 1
			t := rd.ReadAt
			msgs[i].ReadAt = &t
		} else {
			msgs[i].IsRead = 0
			msgs[i].ReadAt = nil
		}
	}
	return nil
}

func (r *MessageRepository) ListSent(ctx context.Context, userID uint64, scope *datascope.Provider, page, pageSize int) ([]model.Message, int64, error) {
	var msgs []model.Message
	var total int64

	senderFilter, args := buildSenderFilterSQL(scope, userID)
	listArgs := append(append([]interface{}{}, args...), pageSize, (page-1)*pageSize)

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			MIN(id) AS id,
			title,
			content,
			type,
			sender_id,
			MAX(sender) AS sender,
			MIN(receiver_id) AS receiver_id,
			0 AS is_read,
			NULL AS read_at,
			MAX(created_at) AS created_at,
			COUNT(*) AS receiver_count
		FROM sys_message
		WHERE `+senderFilter+` AND deleted_at IS NULL AND sender_deleted = 0
		GROUP BY sender_id, title, content, type, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i')
		ORDER BY MAX(created_at) DESC
		LIMIT ? OFFSET ?
	`, listArgs...).Scan(&msgs).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM (
			SELECT 1
			FROM sys_message
			WHERE `+senderFilter+` AND deleted_at IS NULL AND sender_deleted = 0
			GROUP BY sender_id, title, content, type, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i')
		) t
	`, args...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	return msgs, total, nil
}

func buildSenderFilterSQL(scope *datascope.Provider, userID uint64) (string, []interface{}) {
	if scope == nil || scope.Scope == datascope.ScopeSelf || scope.Scope == 0 {
		return "sender_id = ?", []interface{}{userID}
	}
	switch scope.Scope {
	case datascope.ScopeAll:
		return "1 = 1", nil
	case datascope.ScopeCustom, datascope.ScopeDeptAndChild, datascope.ScopeDept:
		if len(scope.DeptIDs) == 0 {
			return "1 = 0", nil
		}
		return "sender_id IN (SELECT id FROM sys_user WHERE dept_id IN ? AND deleted_at IS NULL)", []interface{}{scope.DeptIDs}
	default:
		return "sender_id = ?", []interface{}{userID}
	}
}

func (r *MessageRepository) UnreadCount(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM sys_message m
		WHERE m.deleted_at IS NULL AND (
			(m.receiver_id = ? AND m.is_read = 0 AND m.receiver_deleted = 0) OR
			(m.receiver_id = 0 AND NOT EXISTS (
				SELECT 1 FROM sys_message_read r
				WHERE r.message_id = m.id AND r.user_id = ?
			) AND NOT EXISTS (
				SELECT 1 FROM sys_message_hide h
				WHERE h.message_id = m.id AND h.user_id = ?
			))
		)
	`, userID, userID, userID).Scan(&count).Error
	return count, err
}
