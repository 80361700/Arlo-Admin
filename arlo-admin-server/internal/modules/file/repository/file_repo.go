package repository

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/file/model"

	"arlo-admin/pkg/datascope"

	"gorm.io/gorm"
)

// FileRepository 文件仓储
type FileRepository struct {
	db *gorm.DB
}

// NewFileRepository 创建 FileRepository
func NewFileRepository() *FileRepository {
	return &FileRepository{db: database.DB}
}

// Create 创建文件记录
func (r *FileRepository) Create(ctx context.Context, f *model.SysFile) error {
	return r.db.WithContext(ctx).Create(f).Error
}

// Update 更新文件记录
func (r *FileRepository) Update(ctx context.Context, f *model.SysFile) error {
	return r.db.WithContext(ctx).Save(f).Error
}

// FindByID 按 ID 查询（含软删除）
func (r *FileRepository) FindByID(ctx context.Context, id uint64) (*model.SysFile, error) {
	var f model.SysFile
	if err := r.db.WithContext(ctx).Unscoped().First(&f, id).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// FindByAccessKey 按访问钥查询未删除文件
func (r *FileRepository) FindByAccessKey(ctx context.Context, accessKey string) (*model.SysFile, error) {
	var f model.SysFile
	if err := r.db.WithContext(ctx).Where("access_key = ?", accessKey).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// FindPublicByID 查询未删除且已公开的文件
func (r *FileRepository) FindPublicByID(ctx context.Context, id uint64) (*model.SysFile, error) {
	var f model.SysFile
	err := r.db.WithContext(ctx).Where("id = ? AND is_public = 1", id).First(&f).Error
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// MarkPublic 标记文件为公开
func (r *FileRepository) MarkPublic(ctx context.Context, id uint64) error {
	return r.SetPublic(ctx, id, true)
}

// SetPublic 设置公开/私有
func (r *FileRepository) SetPublic(ctx context.Context, id uint64, isPublic bool) error {
	val := int8(0)
	if isPublic {
		val = 1
	}
	return r.db.WithContext(ctx).Model(&model.SysFile{}).Where("id = ?", id).
		Update("is_public", val).Error
}

// MarkPublicIDs 批量标记公开
func (r *FileRepository) MarkPublicIDs(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.SysFile{}).Where("id IN ?", ids).
		Update("is_public", 1).Error
}

// FindByMD5 按 MD5 查询已存在的文件（用于去重）
func (r *FileRepository) FindByMD5(ctx context.Context, md5 string) (*model.SysFile, error) {
	var f model.SysFile
	if err := r.db.WithContext(ctx).Where("md5 = ?", md5).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// Delete 软删除
func (r *FileRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.SysFile{}, id).Error
}

// List 分页查询
func (r *FileRepository) List(ctx context.Context, name, mimeType, category string, isPublic *int8, scope *datascope.Provider, page, pageSize int) ([]model.SysFile, int64, error) {
	var (
		files []model.SysFile
		total int64
	)

	tx := r.db.WithContext(ctx).Model(&model.SysFile{})
	if scope != nil {
		tx = scope.ApplyByOwner(tx, "uploader_id")
	}
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}
	if mimeType != "" {
		tx = tx.Where("mime_type LIKE ?", "%"+mimeType+"%")
	}
	if category != "" {
		tx = tx.Where("category = ?", category)
	}
	if isPublic != nil {
		tx = tx.Where("is_public = ?", *isPublic)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := tx.Order("id DESC").Offset(offset).Limit(pageSize).Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}
