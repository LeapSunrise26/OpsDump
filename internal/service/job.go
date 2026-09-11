package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"ops-dump/internal/consts"
	"ops-dump/internal/dao"
	"ops-dump/internal/model"
	"ops-dump/internal/model/entity"
)

// JobService groups job CRUD, run history and scheduling operations.
type JobService struct{}

// NewJob creates a JobService instance.
func NewJob() *JobService { return &JobService{} }

// ListJobs returns all jobs ordered by sort_order, then name.
func (s *JobService) ListJobs(ctx context.Context) ([]entity.Job, error) {
	var jobs []entity.Job
	err := dao.Jobs.Ctx(ctx).OrderAsc("sort_order").OrderAsc("name").Scan(&jobs)
	return jobs, err
}

// GetJob returns a single job by id.
func (s *JobService) GetJob(ctx context.Context, id int64) (*entity.Job, error) {
	var job entity.Job
	err := dao.Jobs.Ctx(ctx).Where("id", id).Scan(&job)
	if err != nil {
		return nil, err
	}
	if job.Id == 0 {
		return nil, gerror.NewCode(gcode.CodeNotFound, "任务不存在")
	}
	return &job, nil
}

// CreateJob inserts a new job from the request params.
func (s *JobService) CreateJob(ctx context.Context, p model.JobParam) (int64, error) {
	if p.Kind == "" {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "kind 不能为空")
	}
	if p.Cron == "" {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "cron 不能为空")
	}
	g.Log().Infof(ctx, "create job: name=%s, kind=%s, cron=%s, out_dir=%s", p.Name, p.Kind, p.Cron, p.OutDir)
	res, err := dao.Jobs.Ctx(ctx).Insert(dataNoId(entity.Job{
		Name:             p.Name,
		Kind:             p.Kind,
		Enabled:          p.Enabled,
		SortOrder:        p.SortOrder,
		Cron:             p.Cron,
		DependsOn:        p.DependsOn,
		Description:      p.Description,
		Command:          p.Command,
		VerifyCommand:    p.VerifyCommand,
		OutDir:           p.OutDir,
		RetentionDays:    p.RetentionDays,
		Kv:               p.Kv,
		NotifyOnSuccess:  p.NotifyOnSuccess,
		NotifyOnFailure:  p.NotifyOnFailure,
		CreatedAt:        now(),
		UpdatedAt:        now(),
	}))
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err == nil {
		g.Log().Infof(ctx, "create job ok, id=%d, name=%s", id, p.Name)
	}
	return id, err
}

// UpdateJob updates an existing job.
func (s *JobService) UpdateJob(ctx context.Context, p model.JobParam) error {
	if p.Id == 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "缺少任务 id")
	}
	g.Log().Infof(ctx, "update job: id=%d, name=%s", p.Id, p.Name)
	// Build the update set explicitly so the creation timestamp is never touched.
	_, err := dao.Jobs.Ctx(ctx).Where("id", p.Id).Data(g.Map{
		"name":              p.Name,
		"kind":              p.Kind,
		"enabled":           p.Enabled,
		"sort_order":        p.SortOrder,
		"cron":              p.Cron,
		"depends_on":        p.DependsOn,
		"description":       p.Description,
		"command":           p.Command,
		"verify_command":    p.VerifyCommand,
		"out_dir":           p.OutDir,
		"retention_days":    p.RetentionDays,
		"kv":                p.Kv,
		"notify_on_success": p.NotifyOnSuccess,
		"notify_on_failure": p.NotifyOnFailure,
		"updated_at":        now(),
	}).Update()
	if err == nil {
		g.Log().Infof(ctx, "update job ok, id=%d", p.Id)
	}
	return err
}

// SetJobEnabled toggles the enabled flag and triggers scheduler reload.
func (s *JobService) SetJobEnabled(ctx context.Context, id int64, enabled int) error {
	g.Log().Infof(ctx, "set job enabled: id=%d, enabled=%d", id, enabled)
	if _, err := dao.Jobs.Ctx(ctx).Where("id", id).Data(g.Map{"enabled": enabled, "updated_at": now()}).Update(); err != nil {
		g.Log().Errorf(ctx, "set job enabled failed: id=%d, err=%v", id, err)
		return err
	}
	g.Log().Infof(ctx, "set job enabled ok, id=%d, reloading scheduler", id)
	return s.SchedulerReload(ctx)
}

// SortJobs updates sort_order for each job according to the given id sequence.
func (s *JobService) SortJobs(ctx context.Context, ids []int64) error {
	for i, id := range ids {
		if _, err := dao.Jobs.Ctx(ctx).Where("id", id).Data(g.Map{
			"sort_order": i,
			"updated_at": now(),
		}).Update(); err != nil {
			g.Log().Errorf(ctx, "sort job failed: id=%d, err=%v", id, err)
			return err
		}
	}
	g.Log().Infof(ctx, "sort jobs ok, count=%d", len(ids))
	return nil
}

// ExportJobs returns jobs by ids, or all jobs if ids is empty.
func (s *JobService) ExportJobs(ctx context.Context, ids []int64) ([]entity.Job, error) {
	var jobs []entity.Job
	q := dao.Jobs.Ctx(ctx)
	if len(ids) > 0 {
		q = q.WhereIn("id", ids)
	}
	err := q.OrderAsc("sort_order").OrderAsc("name").Scan(&jobs)
	return jobs, err
}

// ImportResult holds the result of importing a single job.
type ImportResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // ok, skipped, failed
	Message string `json:"message,omitempty"`
}

// ImportJobs inserts jobs from import data, skipping names that already exist.
func (s *JobService) ImportJobs(ctx context.Context, jobs []model.JobParam) ([]ImportResult, error) {
	results := make([]ImportResult, 0, len(jobs))
	for _, p := range jobs {
		result := ImportResult{Name: p.Name}
		// Check if name already exists
		var existing entity.Job
		scanErr := dao.Jobs.Ctx(ctx).Where("name", p.Name).Scan(&existing)
		if scanErr != nil && scanErr != sql.ErrNoRows {
			result.Status = "failed"
			result.Message = scanErr.Error()
			results = append(results, result)
			continue
		}
		if existing.Id > 0 {
			result.Status = "skipped"
			result.Message = "名称已存在"
			results = append(results, result)
			continue
		}
		if _, createErr := s.CreateJob(ctx, p); createErr != nil {
			result.Status = "failed"
			result.Message = createErr.Error()
			results = append(results, result)
			continue
		}
		result.Status = "ok"
		results = append(results, result)
	}
	imported := 0
	skipped := 0
	failed := 0
	for _, r := range results {
		switch r.Status {
		case "ok":
			imported++
		case "skipped":
			skipped++
		case "failed":
			failed++
		}
	}
	g.Log().Infof(ctx, "import jobs: imported=%d, skipped=%d, failed=%d", imported, skipped, failed)
	return results, nil
}

// DeleteJob removes a job by id.
func (s *JobService) DeleteJob(ctx context.Context, id int64) error {
	g.Log().Infof(ctx, "delete job: id=%d", id)
	_, err := dao.Jobs.Ctx(ctx).Where("id", id).Delete()
	if err == nil {
		g.Log().Infof(ctx, "delete job ok, id=%d", id)
	}
	return err
}

// ParseKv decodes the job kv JSON into a generic map.
func (s *JobService) ParseKv(job *entity.Job) map[string]any {
	kv := make(map[string]any)
	if job.Kv != "" {
		_ = json.Unmarshal([]byte(job.Kv), &kv)
	}
	return kv
}

// ListRuns returns recent run records for a job.
func (s *JobService) ListRuns(ctx context.Context, jobId int64, limit int) ([]entity.JobRun, error) {
	if limit <= 0 {
		limit = 50
	}
	var runs []entity.JobRun
	err := dao.JobRuns.Ctx(ctx).Where("job_id", jobId).
		OrderDesc("id").Limit(limit).Scan(&runs)
	return runs, err
}

// ListAllRunsParams holds parameters for ListAllRuns.
type ListAllRunsParams struct {
	Page     int
	PageSize int
	Status   string
	JobName  string
}

// ListAllRunsResult holds the result for ListAllRuns.
type ListAllRunsResult struct {
	List  []entity.JobRun
	Total int
	Page  int
	Pages int
}

// ListAllRuns returns runs with pagination and filtering.
func (s *JobService) ListAllRuns(ctx context.Context, params ListAllRunsParams) (*ListAllRunsResult, error) {
	page := params.Page
	if page <= 0 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	// Build count query
	qCount := dao.JobRuns.Ctx(ctx).
		LeftJoin("od_jobs", "od_jobs.id = od_job_runs.job_id")
	if params.Status != "" {
		qCount = qCount.Where("od_job_runs.status", params.Status)
	}
	if params.JobName != "" {
		qCount = qCount.Where("od_jobs.name", params.JobName)
	}

	// Count total
	total, err := qCount.Count()
	if err != nil {
		return nil, err
	}

	// Calculate pages
	pages := total / pageSize
	if total%pageSize > 0 {
		pages++
	}

	// Fetch page
	offset := (page - 1) * pageSize
	var runs []entity.JobRun
	q2 := dao.JobRuns.Ctx(ctx).
		LeftJoin("od_jobs", "od_jobs.id = od_job_runs.job_id").
		Fields("od_job_runs.*, od_jobs.name AS job_name")
	if params.Status != "" {
		q2 = q2.Where("od_job_runs.status", params.Status)
	}
	if params.JobName != "" {
		q2 = q2.Where("od_jobs.name", params.JobName)
	}
	err = q2.OrderDesc("od_job_runs.id").
		Limit(pageSize).
		Offset(offset).
		Scan(&runs)
	if err != nil {
		return nil, err
	}

	return &ListAllRunsResult{
		List:  runs,
		Total: total,
		Page:  page,
		Pages: pages,
	}, nil
}

// ClearRuns deletes run history filtered by status and/or job name.
// Empty filters delete the whole history. It returns the number of rows removed.
func (s *JobService) ClearRuns(ctx context.Context, status, jobName string) (int64, error) {
	q := dao.JobRuns.Ctx(ctx)
	if status != "" {
		q = q.Where("status", status)
	}
	if jobName != "" {
		var dep entity.Job
		if err := dao.Jobs.Ctx(ctx).Where("name", jobName).Scan(&dep); err != nil {
			return 0, err
		}
		if dep.Id == 0 {
			return 0, gerror.Newf("任务 %s 不存在", jobName)
		}
		q = q.Where("job_id", dep.Id)
	} else if status == "" {
		// Full-history clear (no filter). gdb refuses deletes without a WHERE
		// condition ("there should be WHERE condition statement for DELETE").
		// id is the non-negative primary key, so id >= 0 is always true and
		// satisfies the guard without narrowing the delete.
		q = q.Where("id >= 0")
	}
	r, err := q.Delete()
	if err != nil {
		return 0, err
	}
	n, err := r.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

// DashboardJob aggregates the latest run status per job for the dashboard page.
type DashboardJob struct {
	entity.Job
	LastRun      entity.JobRun `json:"last_run"`
	TodayOk      int           `json:"today_ok"`
	TodayFail    int           `json:"today_fail"`
	TodayRate    string        `json:"today_rate"`
	TodayAvgTime string        `json:"today_avg_time"`
}

// DashboardStats holds the summary statistics for the dashboard.
type DashboardStats struct {
	Total   int    `json:"total"`
	Enabled int    `json:"enabled"`
	Ok      int    `json:"ok"`
	Fail    int    `json:"fail"`
	Rate    string `json:"rate"`
}

// Dashboard returns per-job latest run status and summary stats.
func (s *JobService) Dashboard(ctx context.Context) ([]DashboardJob, *DashboardStats, error) {
	jobs, err := s.ListJobs(ctx)
	if err != nil {
		return nil, nil, err
	}
	// Get today's runs for stats
	today := time.Now().Format("2006-01-02")
	var todayRuns []entity.JobRun
	err = dao.JobRuns.Ctx(ctx).
		LeftJoin("od_jobs", "od_jobs.id = od_job_runs.job_id").
		Where("od_job_runs.finished_at LIKE ?", today+"%").
		Fields("od_job_runs.*").
		Scan(&todayRuns)
	if err != nil {
		return nil, nil, err
	}
	// Calculate summary stats
	stats := &DashboardStats{}
	var totalMs int64
	var avgCount int
	for _, run := range todayRuns {
		if run.Status == consts.StatusOk {
			stats.Ok++
		} else if run.Status == consts.StatusFailed {
			stats.Fail++
		}
		if run.ElapsedMs > 0 {
			totalMs += run.ElapsedMs
			avgCount++
		}
	}
	stats.Total = len(jobs)
	stats.Enabled = 0
	for _, j := range jobs {
		if j.Enabled == 1 {
			stats.Enabled++
		}
	}
	total := stats.Ok + stats.Fail
	if total > 0 {
		stats.Rate = fmt.Sprintf("%.0f%%", float64(stats.Ok)/float64(total)*100)
	} else {
		stats.Rate = "-"
	}

	// Build today runs map by job_id
	todayMap := make(map[int64][]entity.JobRun)
	for _, run := range todayRuns {
		todayMap[run.JobId] = append(todayMap[run.JobId], run)
	}

	out := make([]DashboardJob, 0, len(jobs))
	for _, j := range jobs {
		runs, err := s.ListRuns(ctx, j.Id, 1)
		if err != nil {
			return nil, nil, err
		}
		last := entity.JobRun{}
		if len(runs) > 0 {
			last = runs[0]
		}
		dj := DashboardJob{Job: j, LastRun: last}
		// Calculate per-job today stats
		if jobRuns, ok := todayMap[j.Id]; ok {
			var jobOk, jobFail int
			var jobTotalMs int64
			var jobAvgCount int
			for _, run := range jobRuns {
				if run.Status == consts.StatusOk {
					jobOk++
				} else if run.Status == consts.StatusFailed {
					jobFail++
				}
				if run.ElapsedMs > 0 {
					jobTotalMs += run.ElapsedMs
					jobAvgCount++
				}
			}
			dj.TodayOk = jobOk
			dj.TodayFail = jobFail
			jobTotal := jobOk + jobFail
			if jobTotal > 0 {
				dj.TodayRate = fmt.Sprintf("%.0f%%", float64(jobOk)/float64(jobTotal)*100)
			} else {
				dj.TodayRate = "-"
			}
			if jobAvgCount > 0 {
				avgMs := jobTotalMs / int64(jobAvgCount)
				if avgMs >= 1000 {
					dj.TodayAvgTime = fmt.Sprintf("%.1fs", float64(avgMs)/1000)
				} else {
					dj.TodayAvgTime = fmt.Sprintf("%dms", avgMs)
				}
			} else {
				dj.TodayAvgTime = "-"
			}
		} else {
			dj.TodayRate = "-"
			dj.TodayAvgTime = "-"
		}
		out = append(out, dj)
	}
	return out, stats, nil
}
