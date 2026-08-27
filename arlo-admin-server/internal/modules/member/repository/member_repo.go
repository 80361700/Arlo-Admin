package repository

import (
	"context"
	"time"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/member/model"

	"gorm.io/gorm"
)

type MemberRepository struct {
	db *gorm.DB
}

func NewMemberRepository() *MemberRepository {
	return &MemberRepository{db: database.DB}
}

// FindByPhone 根据手机号查找会员
func (r *MemberRepository) FindByPhone(ctx context.Context, phone string) (*model.Member, error) {
	var m model.Member
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// FindByID 根据ID查找会员
func (r *MemberRepository) FindByID(ctx context.Context, id uint64) (*model.Member, error) {
	var m model.Member
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Create 创建会员（openid 为空时写入 NULL，避免 uk_openid 空串冲突）
func (r *MemberRepository) Create(ctx context.Context, m *model.Member) error {
	db := r.db.WithContext(ctx)
	if m.Openid == "" {
		return db.Omit("Openid").Create(m).Error
	}
	return db.Create(m).Error
}

// ExistsByPhone 手机号是否已存在（excludeID>0 时排除自身）
func (r *MemberRepository) ExistsByPhone(ctx context.Context, phone string, excludeID uint64) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&model.Member{}).Where("phone = ?", phone)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	err := q.Count(&count).Error
	return count > 0, err
}

// Update 更新会员
func (r *MemberRepository) Update(ctx context.Context, m *model.Member) error {
	return r.db.WithContext(ctx).Save(m).Error
}

// UpdatePassword 更新会员密码
func (r *MemberRepository) UpdatePassword(ctx context.Context, id uint64, password string) error {
	return r.db.WithContext(ctx).Model(&model.Member{}).Where("id = ?", id).
		Update("password", password).Error
}

// UpdateLastLogin 更新最后登录时间
func (r *MemberRepository) UpdateLastLogin(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.Member{}).Where("id = ?", id).
		Update("last_login", time.Now()).Error
}

// List 管理员分页查询会员列表
func (r *MemberRepository) List(ctx context.Context, page, pageSize int, phone, nickname, source string, status *int8) ([]model.Member, int64, error) {
	var total int64
	var members []model.Member

	query := r.db.WithContext(ctx).Model(&model.Member{})
	if phone != "" {
		query = query.Where("phone LIKE ?", "%"+phone+"%")
	}
	if nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+nickname+"%")
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&members).Error; err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

// UpdateStatus 管理员更新会员状态（启用/禁用）
func (r *MemberRepository) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	return r.db.WithContext(ctx).Model(&model.Member{}).Where("id = ?", id).
		Update("status", status).Error
}

// Delete 软删除会员
func (r *MemberRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Member{}, id).Error
}
