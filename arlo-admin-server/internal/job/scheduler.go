package job

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"arlo-admin/internal/modules/job/model"
	"arlo-admin/internal/modules/job/repository"
	"arlo-admin/pkg/logger"

	"go.uber.org/zap"
)

const (
	TriggerCron   int8 = 0
	TriggerManual int8 = 1
)

// Scheduler 进程内轻量调度（分钟精度；暂无分布式锁）
type Scheduler struct {
	repo   *repository.JobRepository
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu       sync.Mutex
	specs    map[uint64]*CronSpec // jobID → cron
	running  map[uint64]bool      // 同任务禁止并发
	lastFire map[uint64]string    // jobID → "200601021504" 防同分钟重复
}

// Start 启动调度；repo 为空或 DB 不可用时 no-op
func Start(repo *repository.JobRepository) *Scheduler {
	s := &Scheduler{
		repo:     repo,
		specs:    map[uint64]*CronSpec{},
		running:  map[uint64]bool{},
		lastFire: map[uint64]string{},
	}
	RegisterBuiltin()
	if repo == nil {
		logger.Logger.Info("job scheduler skipped: repository unavailable")
		return s
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	if err := s.Reload(ctx); err != nil {
		logger.Logger.Warn("job scheduler initial reload failed", zap.Error(err))
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.loop(ctx)
	}()
	logger.Logger.Info("job scheduler started")
	return s
}

// Stop 停止
func (s *Scheduler) Stop() {
	if s == nil || s.cancel == nil {
		return
	}
	s.cancel()
	s.wg.Wait()
	logger.Logger.Info("job scheduler stopped")
}

// Reload 从库加载启用中的任务
func (s *Scheduler) Reload(ctx context.Context) error {
	if s == nil || s.repo == nil {
		return nil
	}
	list, err := s.repo.ListEnabled(ctx)
	if err != nil {
		return err
	}
	next := make(map[uint64]*CronSpec, len(list))
	for _, j := range list {
		spec, err := ParseCron(j.Cron)
		if err != nil {
			logger.Logger.Warn("skip invalid job cron",
				zap.Uint64("jobId", j.ID),
				zap.String("cron", j.Cron),
				zap.Error(err),
			)
			continue
		}
		next[j.ID] = spec
	}
	s.mu.Lock()
	s.specs = next
	s.mu.Unlock()
	logger.Logger.Info("job scheduler reloaded", zap.Int("enabled", len(next)))
	return nil
}

// NextRunAt 计算下次执行时间
func (s *Scheduler) NextRunAt(cronExpr string) string {
	spec, err := ParseCron(cronExpr)
	if err != nil {
		return ""
	}
	n := spec.Next(time.Now())
	if n.IsZero() {
		return ""
	}
	return n.Format("2006-01-02 15:04:05")
}

func (s *Scheduler) loop(ctx context.Context) {
	// 对齐到下一分钟
	now := time.Now()
	wait := time.Until(now.Truncate(time.Minute).Add(time.Minute))
	timer := time.NewTimer(wait)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			s.tick(ctx, time.Now())
			timer.Reset(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
		}
	}
}

func (s *Scheduler) tick(ctx context.Context, now time.Time) {
	s.mu.Lock()
	ids := make([]uint64, 0, len(s.specs))
	for id, spec := range s.specs {
		if spec.Matches(now) {
			key := now.Format("200601021504")
			if s.lastFire[id] == key {
				continue
			}
			s.lastFire[id] = key
			ids = append(ids, id)
		}
	}
	s.mu.Unlock()

	for _, id := range ids {
		go s.RunJob(ctx, id, TriggerCron)
	}
}

// RunJob 执行任务（调度或手动）；同任务串行
func (s *Scheduler) RunJob(ctx context.Context, jobID uint64, triggerType int8) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("调度器未就绪")
	}

	s.mu.Lock()
	if s.running[jobID] {
		s.mu.Unlock()
		return fmt.Errorf("任务正在执行中")
	}
	s.running[jobID] = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.running, jobID)
		s.mu.Unlock()
	}()

	j, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}
	if triggerType == TriggerCron && j.Status != 1 {
		return nil
	}

	fn, ok := GetHandler(j.Handler)
	start := time.Now()
	var (
		result   string
		runErr   error
		status   int8 = 1
		errorMsg string
	)
	if !ok {
		runErr = fmt.Errorf("未注册的处理器: %s", j.Handler)
	} else {
		// 与请求生命周期解耦，避免手动执行时客户端断开导致任务中断
		runCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		result, runErr = fn(runCtx, j.Params)
	}
	logCtx := context.Background()
	if runErr != nil {
		status = 0
		errorMsg = runErr.Error()
		if len(errorMsg) > 1000 {
			errorMsg = errorMsg[:1000]
		}
		logger.Logger.Error("job failed",
			zap.Uint64("jobId", jobID),
			zap.String("handler", j.Handler),
			zap.Error(runErr),
		)
	} else {
		logger.Logger.Info("job done",
			zap.Uint64("jobId", jobID),
			zap.String("handler", j.Handler),
			zap.String("result", truncate(result, 200)),
		)
	}

	duration := int(time.Since(start).Milliseconds())
	_ = s.repo.CreateLog(logCtx, &model.SysJobLog{
		JobID:       j.ID,
		JobName:     j.Name,
		Handler:     j.Handler,
		TriggerType: triggerType,
		Status:      status,
		Result:      result,
		ErrorMsg:    errorMsg,
		DurationMs:  duration,
	})
	_ = s.repo.UpdateLastRun(logCtx, j.ID, status, time.Now())
	return runErr
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// ParseRetainDays 从 params JSON 读 retainDays
func ParseRetainDays(params string, def int) int {
	if def <= 0 {
		def = 90
	}
	if params == "" {
		return def
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(params), &m); err != nil {
		return def
	}
	switch v := m["retainDays"].(type) {
	case float64:
		if int(v) > 0 {
			return int(v)
		}
	case int:
		if v > 0 {
			return v
		}
	}
	return def
}
