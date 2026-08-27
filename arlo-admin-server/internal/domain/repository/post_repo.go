package repository

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/domain/model"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository() *PostRepository {
	return &PostRepository{db: database.DB}
}

func (r *PostRepository) Create(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *PostRepository) Update(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *PostRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", id).Delete(&model.UserPost{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Post{}, id).Error
	})
}

func (r *PostRepository) FindByID(ctx context.Context, id uint64) (*model.Post, error) {
	var post model.Post
	err := r.db.WithContext(ctx).First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) FindAll(ctx context.Context) ([]model.Post, error) {
	var posts []model.Post
	err := r.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&posts).Error
	return posts, err
}

func (r *PostRepository) ExistsByCode(ctx context.Context, code string, excludeID uint64) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&model.Post{}).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	err := q.Count(&count).Error
	return count > 0, err
}

func (r *PostRepository) List(ctx context.Context, code, name string, status *int8, page, pageSize int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Post{})
	if code != "" {
		q = q.Where("code LIKE ?", "%"+code+"%")
	}
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Order("sort ASC, id ASC").Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}
