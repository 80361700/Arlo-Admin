package repository

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/flow/model"

	"gorm.io/gorm"
)

type FlowRepository struct{}

func NewFlowRepository() *FlowRepository { return &FlowRepository{} }

func (r *FlowRepository) db() *gorm.DB { return database.DB }

// --- category ---

func (r *FlowRepository) ListCategories(ctx context.Context) ([]model.FlowCategory, error) {
	var list []model.FlowCategory
	err := r.db().WithContext(ctx).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *FlowRepository) GetCategory(ctx context.Context, id uint64) (*model.FlowCategory, error) {
	var c model.FlowCategory
	if err := r.db().WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *FlowRepository) CreateCategory(ctx context.Context, c *model.FlowCategory) error {
	return r.db().WithContext(ctx).Create(c).Error
}

func (r *FlowRepository) UpdateCategory(ctx context.Context, c *model.FlowCategory) error {
	return r.db().WithContext(ctx).Save(c).Error
}

func (r *FlowRepository) DeleteCategory(ctx context.Context, id uint64) error {
	return r.db().WithContext(ctx).Delete(&model.FlowCategory{}, id).Error
}

func (r *FlowRepository) CountProcessByCategory(ctx context.Context, categoryID uint64) (int64, error) {
	var n int64
	err := r.db().WithContext(ctx).Model(&model.FlowProcess{}).Where("category_id = ?", categoryID).Count(&n).Error
	return n, err
}

// --- process ---

func (r *FlowRepository) ListProcesses(ctx context.Context, categoryID *uint64, name string) ([]model.FlowProcess, error) {
	tx := r.db().WithContext(ctx).Model(&model.FlowProcess{})
	if categoryID != nil && *categoryID > 0 {
		tx = tx.Where("category_id = ?", *categoryID)
	}
	if name != "" {
		tx = tx.Where("process_name LIKE ?", "%"+name+"%")
	}
	var list []model.FlowProcess
	err := tx.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *FlowRepository) GetProcess(ctx context.Context, id uint64) (*model.FlowProcess, error) {
	var p model.FlowProcess
	if err := r.db().WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *FlowRepository) GetProcessByKey(ctx context.Context, key string) (*model.FlowProcess, error) {
	var p model.FlowProcess
	if err := r.db().WithContext(ctx).Where("process_key = ?", key).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *FlowRepository) CreateProcess(ctx context.Context, p *model.FlowProcess) error {
	return r.db().WithContext(ctx).Create(p).Error
}

func (r *FlowRepository) UpdateProcess(ctx context.Context, p *model.FlowProcess) error {
	return r.db().WithContext(ctx).Save(p).Error
}

func (r *FlowRepository) DeleteProcess(ctx context.Context, id uint64) error {
	return r.db().WithContext(ctx).Delete(&model.FlowProcess{}, id).Error
}

// --- form category ---

func (r *FlowRepository) ListFormCategories(ctx context.Context) ([]model.FlowFormCategory, error) {
	var list []model.FlowFormCategory
	err := r.db().WithContext(ctx).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *FlowRepository) GetFormCategory(ctx context.Context, id uint64) (*model.FlowFormCategory, error) {
	var c model.FlowFormCategory
	if err := r.db().WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *FlowRepository) CreateFormCategory(ctx context.Context, c *model.FlowFormCategory) error {
	return r.db().WithContext(ctx).Create(c).Error
}

func (r *FlowRepository) UpdateFormCategory(ctx context.Context, c *model.FlowFormCategory) error {
	return r.db().WithContext(ctx).Save(c).Error
}

func (r *FlowRepository) DeleteFormCategory(ctx context.Context, id uint64) error {
	return r.db().WithContext(ctx).Delete(&model.FlowFormCategory{}, id).Error
}

func (r *FlowRepository) CountFormByCategory(ctx context.Context, categoryID uint64) (int64, error) {
	var n int64
	err := r.db().WithContext(ctx).Model(&model.FlowForm{}).Where("category_id = ?", categoryID).Count(&n).Error
	return n, err
}

// --- form template ---

func (r *FlowRepository) ListForms(ctx context.Context) ([]model.FlowForm, error) {
	var list []model.FlowForm
	err := r.db().WithContext(ctx).Order("sort ASC, id DESC").Find(&list).Error
	return list, err
}

func (r *FlowRepository) ListEnabledForms(ctx context.Context) ([]model.FlowForm, error) {
	var list []model.FlowForm
	err := r.db().WithContext(ctx).Where("status = ?", 1).
		Order("sort ASC, id DESC").Find(&list).Error
	return list, err
}

func (r *FlowRepository) GetForm(ctx context.Context, id uint64) (*model.FlowForm, error) {
	var f model.FlowForm
	if err := r.db().WithContext(ctx).First(&f, id).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FlowRepository) GetFormByCode(ctx context.Context, code string) (*model.FlowForm, error) {
	var f model.FlowForm
	if err := r.db().WithContext(ctx).Where("code = ?", code).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FlowRepository) CreateForm(ctx context.Context, f *model.FlowForm) error {
	return r.db().WithContext(ctx).Create(f).Error
}

func (r *FlowRepository) UpdateForm(ctx context.Context, f *model.FlowForm) error {
	return r.db().WithContext(ctx).Save(f).Error
}

func (r *FlowRepository) DeleteForm(ctx context.Context, id uint64) error {
	return r.db().WithContext(ctx).Delete(&model.FlowForm{}, id).Error
}

// ListProcessFormBindSources 用于检测表单是否被流程引用（节点子表单 / 业务审批绑定）
func (r *FlowRepository) ListProcessFormBindSources(ctx context.Context) ([]model.FlowProcess, error) {
	var list []model.FlowProcess
	err := r.db().WithContext(ctx).Select("id", "model_content", "process_setting").Find(&list).Error
	return list, err
}

// --- process history ---

func (r *FlowRepository) CreateProcessHistory(ctx context.Context, h *model.FlowProcessHistory) error {
	return r.db().WithContext(ctx).Create(h).Error
}

func (r *FlowRepository) CreateProcessHistoryTx(tx *gorm.DB, h *model.FlowProcessHistory) error {
	return tx.Create(h).Error
}

// EnsureProcessHistoryTx 写入历史快照；同流程同版本已存在则跳过（历史不可变）
func (r *FlowRepository) EnsureProcessHistoryTx(tx *gorm.DB, h *model.FlowProcessHistory) error {
	var count int64
	if err := tx.Model(&model.FlowProcessHistory{}).
		Where("process_id = ? AND process_version = ?", h.ProcessID, h.ProcessVersion).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return tx.Create(h).Error
}

func (r *FlowRepository) UpdateProcessTx(tx *gorm.DB, p *model.FlowProcess) error {
	return tx.Save(p).Error
}

func (r *FlowRepository) CreateProcessTx(tx *gorm.DB, p *model.FlowProcess) error {
	return tx.Create(p).Error
}

func (r *FlowRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db().WithContext(ctx).Transaction(fn)
}

func (r *FlowRepository) ListProcessHistories(ctx context.Context, processID uint64, excludeVersion int, page, pageSize int) ([]model.FlowProcessHistory, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	tx := r.db().WithContext(ctx).Model(&model.FlowProcessHistory{}).Where("process_id = ?", processID)
	if excludeVersion > 0 {
		tx = tx.Where("process_version <> ?", excludeVersion)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FlowProcessHistory
	err := tx.Order("process_version DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *FlowRepository) GetProcessHistory(ctx context.Context, id uint64) (*model.FlowProcessHistory, error) {
	var h model.FlowProcessHistory
	if err := r.db().WithContext(ctx).First(&h, id).Error; err != nil {
		return nil, err
	}
	return &h, nil
}
