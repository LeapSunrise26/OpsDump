package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gctx"

	"ops-dump/internal/dao"
	"ops-dump/internal/model/entity"
)

var (
	schedMu    sync.Mutex
	schedEntry = map[string]*gcron.Entry{}
)

// normalizeCron makes gcron accept standard 5-field crons by prepending "0 ".
// gcron uses 6 fields (seconds minutes hours day month week), while users are
// used to the classic 5-field form (minutes hours day month week).
func normalizeCron(expr string) string {
	fields := strings.Fields(expr)
	if len(fields) == 5 {
		return "0 " + expr
	}
	return expr
}

// SchedulerReload (re)registers cron tasks for all enabled jobs. It is invoked at
// startup and after any job create/update/delete/enable/disable change.
func (s *JobService) SchedulerReload(ctx context.Context) error {
	schedMu.Lock()
	defer schedMu.Unlock()

	g.Log().Info(ctx, "scheduler: reloading cron jobs")
	for name, e := range schedEntry {
		e.Close()
		delete(schedEntry, name)
	}
	schedEntry = map[string]*gcron.Entry{}

	var jobs []entity.Job
	if err := dao.Jobs.Ctx(ctx).Where("enabled", 1).Scan(&jobs); err != nil {
		g.Log().Errorf(ctx, "scheduler: query enabled jobs failed: %v", err)
		return err
	}
	g.Log().Infof(ctx, "scheduler: found %d enabled jobs", len(jobs))
	// Jobs referenced by another job's depends_on are driven by that job and
	// must not register their own cron schedule (avoid duplicate execution).
	depNames := map[string]bool{}
	for i := range jobs {
		if jobs[i].DependsOn != "" {
			depNames[jobs[i].DependsOn] = true
		}
	}
	for i := range jobs {
		job := jobs[i]
		if depNames[job.Name] {
			g.Log().Infof(ctx, "job %d(%s) is a dependency of another job; cron skipped", job.Id, job.Name)
			continue
		}
		entryName := fmt.Sprintf("job-%d", job.Id)
		// Use an independent context so cron callbacks outlive the request
		// context that triggered SchedulerReload (avoid "context canceled").
		entry, err := gcron.AddSingleton(gctx.New(), normalizeCron(job.Cron), func(ctx context.Context) {
			s.executeJob(ctx, job.Id)
		}, entryName)
		if err != nil {
			g.Log().Errorf(ctx, "register cron %s(%s) failed: %+v", job.Name, job.Cron, err)
			continue
		}
		schedEntry[entryName] = entry
		g.Log().Infof(ctx, "scheduler: registered job %d(%s) cron=%s", job.Id, job.Name, job.Cron)
	}
	g.Log().Infof(ctx, "scheduler: reload complete, %d cron jobs registered", len(schedEntry))
	return nil
}

// executeJob loads the freshest job by id and runs it.
func (s *JobService) executeJob(ctx context.Context, jobId int64) {
	g.Log().Infof(ctx, "cron trigger: loading job %d", jobId)
	var job entity.Job
	if err := dao.Jobs.Ctx(ctx).Where("id", jobId).Scan(&job); err != nil || job.Id == 0 {
		g.Log().Errorf(ctx, "load job %d failed: %+v", jobId, err)
		return
	}
	g.Log().Infof(ctx, "cron trigger: job %s(id=%d) loaded, executing", job.Name, job.Id)
	s.RunJob(ctx, &job)
}

// ManualRun triggers a job immediately in the background.
func (s *JobService) ManualRun(ctx context.Context, id int64) (string, error) {
	g.Log().Infof(ctx, "manual run: job id=%d", id)
	job, err := s.GetJob(ctx, id)
	if err != nil {
		g.Log().Errorf(ctx, "manual run: get job %d failed: %v", id, err)
		return "", err
	}
	g.Log().Infof(ctx, "manual run: job %s(id=%d) starting in background", job.Name, job.Id)
	// Run asynchronously so the HTTP response returns immediately;
	// use a fresh context so the job outlives the request.
	go s.RunJob(gctx.New(), job)
	return "running", nil
}
