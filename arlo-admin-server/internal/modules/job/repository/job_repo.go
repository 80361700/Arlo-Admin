package repository

import (
	"context"
	"time"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/job/model"

	"gorm.io/gorm"
)

type JobRepository struct{}

func NewJobRepository() *JobRepository { return &JobRepository{} }

func (r *JobRepository) db() *gorm.DB {
	return database.DB
}

func (r *JobRepository) List(ctx context.Context, name, handler string, status *int8, page, pageSize int) ([]model.SysJob, int64, error) {
	tx := r.db().WithContext(ctx).Model(&model.SysJob{})
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}
	if handler != "" {
		tx = tx.Where("handler = ?", handler)
	}
	if status != nil {
		tx = tx.Where("status = ?", *status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.SysJob
	err := tx.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *JobRepository) ListEnabled(ctx context.Context) ([]model.SysJob, error) {
	var list []model.SysJob
	err := r.db().WithContext(ctx).Where("status = 1").Find(&list).Error
	return list, err
}

func (r *JobRepository) GetByID(ctx context.Context, id uint64) (*model.SysJob, error) {
	var j model.SysJob
	if err := r.db().WithContext(ctx).First(&j, id).Error; err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *JobRepository) Create(ctx context.Context, j *model.SysJob) error {
	return r.db().WithContext(ctx).Create(j).Error
}

func (r *JobRepository) Update(ctx context.Context, j *model.SysJob) error {
	return r.db().WithContext(ctx).Save(j).Error
}

func (r *JobRepository) Delete(ctx context.Context, id uint64) error {
	return r.db().WithContext(ctx).Delete(&model.SysJob{}, id).Error
}

func (r *JobRepository) UpdateLastRun(ctx context.Context, id uint64, status int8, at time.Time) error {
	return r.db().WithContext(ctx).Model(&model.SysJob{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_run_at": at,
			"last_status": status,
		}).Error
}

func (r *JobRepository) CreateLog(ctx context.Context, log *model.SysJobLog) error {
	return r.db().WithContext(ctx).Create(log).Error
}

func (r *JobRepository) ListLogs(ctx context.Context, jobID *uint64, status, triggerType *int8, page, pageSize int) ([]model.SysJobLog, int64, error) {
	tx := r.db().WithContext(ctx).Model(&model.SysJobLog{})
	if jobID != nil {
		tx = tx.Where("job_id = ?", *jobID)
	}
	if status != nil {
		tx = tx.Where("status = ?", *status)
	}
	if triggerType != nil {
		tx = tx.Where("trigger_type = ?", *triggerType)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.SysJobLog
	err := tx.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *JobRepository) GetLogByID(ctx context.Context, id uint64) (*model.SysJobLog, error) {
	var log model.SysJobLog
	if err := r.db().WithContext(ctx).First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}
