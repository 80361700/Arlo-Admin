package service

import (
	"context"
	"errors"
	"time"

	corejob "arlo-admin/internal/job"
	"arlo-admin/internal/modules/job/dto"
	"arlo-admin/internal/modules/job/model"
	"arlo-admin/internal/modules/job/repository"

	"gorm.io/gorm"
)

type JobService struct {
	repo *repository.JobRepository
	sch  *corejob.Scheduler
}

func NewJobService(repo *repository.JobRepository, sch *corejob.Scheduler) *JobService {
	return &JobService{repo: repo, sch: sch}
}

func (s *JobService) List(ctx context.Context, q *dto.JobListQuery) (*dto.JobListResponse, error) {
	page, size := normalizePage(q.Page, q.PageSize)
	list, total, err := s.repo.List(ctx, q.Name, q.Handler, q.Status, page, size)
	if err != nil {
		return nil, err
	}
	out := make([]dto.JobResponse, 0, len(list))
	for _, j := range list {
		out = append(out, s.toJobResp(&j))
	}
	return &dto.JobListResponse{List: out, Total: total, Page: page, PageSize: size}, nil
}

func (s *JobService) Get(ctx context.Context, id uint64) (*dto.JobResponse, error) {
	j, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	resp := s.toJobResp(j)
	return &resp, nil
}

func (s *JobService) Create(ctx context.Context, req *dto.CreateJobRequest) (*dto.JobResponse, error) {
	if _, ok := corejob.GetHandler(req.Handler); !ok {
		return nil, errBadHandler
	}
	if _, err := corejob.ParseCron(req.Cron); err != nil {
		return nil, err
	}
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	j := &model.SysJob{
		Name:    req.Name,
		Handler: req.Handler,
		Cron:    req.Cron,
		Params:  req.Params,
		Status:  status,
		Remark:  req.Remark,
	}
	if err := s.repo.Create(ctx, j); err != nil {
		return nil, err
	}
	_ = s.sch.Reload(ctx)
	resp := s.toJobResp(j)
	return &resp, nil
}

func (s *JobService) Update(ctx context.Context, id uint64, req *dto.UpdateJobRequest) error {
	j, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := corejob.ParseCron(req.Cron); err != nil {
		return err
	}
	j.Name = req.Name
	j.Cron = req.Cron
	j.Params = req.Params
	j.Remark = req.Remark
	if err := s.repo.Update(ctx, j); err != nil {
		return err
	}
	return s.sch.Reload(ctx)
}

func (s *JobService) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	j, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	j.Status = status
	if err := s.repo.Update(ctx, j); err != nil {
		return err
	}
	return s.sch.Reload(ctx)
}

func (s *JobService) Delete(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return s.sch.Reload(ctx)
}

func (s *JobService) RunOnce(ctx context.Context, id uint64) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}
	return s.sch.RunJob(ctx, id, corejob.TriggerManual)
}

func (s *JobService) ListHandlers() []corejob.HandlerInfo {
	return corejob.ListHandlers()
}

func (s *JobService) ListLogs(ctx context.Context, q *dto.JobLogQuery) (*dto.JobLogListResponse, error) {
	page, size := normalizePage(q.Page, q.PageSize)
	list, total, err := s.repo.ListLogs(ctx, q.JobID, q.Status, q.TriggerType, page, size)
	if err != nil {
		return nil, err
	}
	out := make([]dto.JobLogResponse, 0, len(list))
	for _, l := range list {
		out = append(out, toLogResp(&l))
	}
	return &dto.JobLogListResponse{List: out, Total: total, Page: page, PageSize: size}, nil
}

func (s *JobService) GetLog(ctx context.Context, id uint64) (*dto.JobLogResponse, error) {
	l, err := s.repo.GetLogByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toLogResp(l)
	return &resp, nil
}

var errBadHandler = errors.New("未注册的处理器")

func (s *JobService) toJobResp(j *model.SysJob) dto.JobResponse {
	var next string
	if s.sch != nil {
		next = s.sch.NextRunAt(j.Cron)
	}
	return dto.JobResponse{
		ID:         j.ID,
		Name:       j.Name,
		Handler:    j.Handler,
		Cron:       j.Cron,
		Params:     j.Params,
		Status:     j.Status,
		Remark:     j.Remark,
		LastRunAt:  formatTimePtr(j.LastRunAt),
		LastStatus: j.LastStatus,
		NextRunAt:  next,
		CreatedAt:  j.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  j.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toLogResp(l *model.SysJobLog) dto.JobLogResponse {
	return dto.JobLogResponse{
		ID:          l.ID,
		JobID:       l.JobID,
		JobName:     l.JobName,
		Handler:     l.Handler,
		TriggerType: l.TriggerType,
		Status:      l.Status,
		Result:      l.Result,
		ErrorMsg:    l.ErrorMsg,
		DurationMs:  l.DurationMs,
		CreatedAt:   l.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func normalizePage(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return page, size
}

// IsBadHandler 是否处理器错误
func IsBadHandler(err error) bool {
	return errors.Is(err, errBadHandler)
}
