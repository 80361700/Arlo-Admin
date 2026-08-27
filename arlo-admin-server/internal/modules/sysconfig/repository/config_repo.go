package repository

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/sysconfig/model"

	"gorm.io/gorm"
)

// ConfigRepository 配置仓储
type ConfigRepository struct {
	db *gorm.DB
}

// NewConfigRepository 创建 ConfigRepository
func NewConfigRepository() *ConfigRepository {
	return &ConfigRepository{db: database.DB}
}

// FindAll 查询所有配置（支持可选筛选）
func (r *ConfigRepository) FindAll(ctx context.Context, query map[string]interface{}) ([]model.SysConfig, error) {
	var configs []model.SysConfig
	tx := r.db.WithContext(ctx).Model(&model.SysConfig{})
	if name, ok := query["name"]; ok && name != "" {
		tx = tx.Where("name LIKE ?", "%"+name.(string)+"%")
	}
	if key, ok := query["key"]; ok && key != "" {
		tx = tx.Where("`key` LIKE ?", "%"+key.(string)+"%")
	}
	if t, ok := query["type"]; ok && t.(int8) > 0 {
		tx = tx.Where("type = ?", t)
	}
	if err := tx.Order("id ASC").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// FindByKey 按 key 查询
func (r *ConfigRepository) FindByKey(ctx context.Context, key string) (*model.SysConfig, error) {
	var config model.SysConfig
	if err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// FindByID 按 ID 查询
func (r *ConfigRepository) FindByID(ctx context.Context, id uint64) (*model.SysConfig, error) {
	var config model.SysConfig
	if err := r.db.WithContext(ctx).First(&config, id).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// Create 创建配置
func (r *ConfigRepository) Create(ctx context.Context, config *model.SysConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// Update 更新配置
func (r *ConfigRepository) Update(ctx context.Context, config *model.SysConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete 软删除配置
func (r *ConfigRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.SysConfig{}, id).Error
}

// ExistsByKey 检查 key 是否已存在（排除指定 ID）
func (r *ConfigRepository) ExistsByKey(ctx context.Context, key string, excludeID uint64) (bool, error) {
	var count int64
	tx := r.db.WithContext(ctx).Model(&model.SysConfig{}).Where("`key` = ?", key)
	if excludeID > 0 {
		tx = tx.Where("id != ?", excludeID)
	}
	if err := tx.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
