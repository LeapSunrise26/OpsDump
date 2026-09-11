package service

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"ops-dump/internal/consts"
	"ops-dump/internal/dao"
	"ops-dump/internal/model/entity"
)

var jobLocks sync.Map // jobId(int64) -> *sync.Mutex

func lockFor(id int64) *sync.Mutex {
	li, _ := jobLocks.LoadOrStore(id, &sync.Mutex{})
	return li.(*sync.Mutex)
}

// substituteCommand replaces template tokens with concrete job values.
func substituteCommand(kv map[string]any, command string) string {
	replacements := make([]string, 0, len(kv)*2)
	for key, value := range kv {
		if values, ok := value.([]any); ok {
			if len(values) == 0 {
				continue
			}
			value = values[0]
		}
		replacements = append(replacements, "<"+key+">", fmt.Sprint(value))
	}
	return strings.NewReplacer(replacements...).Replace(command)
}

func dirSizeMB(dir string) float64 {
	var total int64
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			if fi, e := d.Info(); e == nil {
				total += fi.Size()
			}
		}
		return nil
	})
	return float64(total) / (1 << 20)
}

// unsafeCleanTopLevels lists top-level directories that must never be treated
// as a cleanup root when BACKUP_DIR is not configured (defensive fallback).
var unsafeCleanTopLevels = map[string]bool{
	"home": true, "root": true, "Users": true,
	"etc": true, "usr": true, "var": true, "opt": true, "srv": true,
	"tmp": true, "mnt": true, "media": true, "boot": true, "proc": true,
	"sys": true, "dev": true, "data": true, "run": true, "bin": true,
	"sbin": true, "lib": true, "lib64": true,
}

// isSafeCleanDir reports whether dir is safe for retention cleanup.
// Primary guard: when BACKUP_DIR is configured, cleanup is only allowed
// inside it (dir == BACKUP_DIR or a descendant), so a mistyped out_dir
// (/, /home, /var, ...) is refused before anything is removed.
// Fallback when BACKUP_DIR is not configured: require an absolute path at
// least 3 levels deep whose top level is not a system directory.
func isSafeCleanDir(dir string) bool {
	if dir == "" || !filepath.IsAbs(dir) {
		return false
	}
	abs := filepath.Clean(dir)
	if bk := Env.Get("BACKUP_DIR"); bk != "" {
		babs, err := filepath.Abs(bk)
		if err != nil {
			return false
		}
		babs = filepath.Clean(babs)
		return abs == babs || strings.HasPrefix(abs, babs+string(filepath.Separator))
	}
	vol := filepath.VolumeName(abs)
	rel := strings.TrimPrefix(abs, vol)
	segs := strings.FieldsFunc(rel, func(r rune) bool { return r == '/' || r == '\\' })
	if len(segs) < 3 {
		return false
	}
	return !unsafeCleanTopLevels[segs[0]]
}

// cleanRetention deletes FILES in dir older than retentionDays days.
// Directories are deliberately skipped: job outputs are expected to be single
// archive files (.tar / .sql.gz), so directory-shaped leftovers are never
// auto-removed. Existing directory leftovers must be cleaned up manually.
func cleanRetention(ctx context.Context, dir string, retentionDays int) {
	if !isSafeCleanDir(dir) {
		g.Log().Errorf(ctx, "retention: refuse cleanup of unsafe dir %q (must be inside BACKUP_DIR when configured)", dir)
		return
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	entries, err := os.ReadDir(dir)
	if err != nil {
		g.Log().Warningf(ctx, "retention: read dir %s failed: %v", dir, err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue // directory-shaped leftovers: never auto-removed
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			path := filepath.Join(dir, entry.Name())
			if err := os.Remove(path); err != nil {
				g.Log().Warningf(ctx, "retention: remove %s failed: %v", path, err)
			} else {
				g.Log().Infof(ctx, "retention: removed %s", path)
			}
		}
	}
}

// RunJob executes a single job (command + optional verify) and records the result.
// It returns the final run status. A job's depends_on dependency is executed first;
// the current job runs only when the dependency succeeds.
func (s *JobService) RunJob(ctx context.Context, job *entity.Job) string {
	return s.runJob(ctx, job, map[int64]bool{})
}

// runJob implements RunJob with cycle detection along the dependency chain.
func (s *JobService) runJob(ctx context.Context, job *entity.Job, chain map[int64]bool) string {
	if job == nil || job.Id == 0 {
		g.Log().Warning(ctx, "runJob: job is nil or id=0")
		return consts.StatusFailed
	}
	if chain[job.Id] {
		g.Log().Errorf(ctx, "job %s(id=%d) dependency cycle detected", job.Name, job.Id)
		return consts.StatusFailed
	}
	chain[job.Id] = true
	defer delete(chain, job.Id)

	lock := lockFor(job.Id)
	lock.Lock()
	defer lock.Unlock()

	start := time.Now()
	g.Log().Infof(ctx, "job %s(id=%d) start, out_dir=%s, retention_days=%d", job.Name, job.Id, job.OutDir, job.RetentionDays)

	// depends_on dependency: run the dependency job FIRST so its run id precedes
	// this job's id in history (execution order == id order).
	if job.DependsOn != "" {
		g.Log().Infof(ctx, "job %s depends on %s, running dependency first", job.Name, job.DependsOn)
		var dep entity.Job
		if err := dao.Jobs.Ctx(ctx).Where("name", job.DependsOn).Scan(&dep); err != nil {
			g.Log().Errorf(ctx, "job %s query dependency %s failed: %v", job.Name, job.DependsOn, err)
			return s.dependencyFailed(ctx, job, fmt.Sprintf("依赖任务 %s 查询失败: %v", job.DependsOn, err))
		}
		if dep.Id == 0 {
			g.Log().Errorf(ctx, "job %s dependency %s not found", job.Name, job.DependsOn)
			return s.dependencyFailed(ctx, job, fmt.Sprintf("依赖任务 %s 不存在", job.DependsOn))
		}
		depStatus := s.runJob(ctx, &dep, chain)
		if depStatus != consts.StatusOk {
			g.Log().Errorf(ctx, "job %s dependency %s failed with status %s", job.Name, job.DependsOn, depStatus)
			return s.dependencyFailed(ctx, job, fmt.Sprintf("依赖任务 %s 执行结果 %s", job.DependsOn, depStatus))
		}
		g.Log().Infof(ctx, "job %s dependency %s completed successfully", job.Name, job.DependsOn)
	}

	kv := s.ParseKv(job)
	commandTemplate := substituteCommand(kv, job.Command)
	// Replace <output_dir> with out_dir if set.
	if job.OutDir != "" {
		commandTemplate = strings.ReplaceAll(commandTemplate, "<output_dir>", job.OutDir)
	}
	command := Env.Resolve(commandTemplate)
	runId, createErr := createRun(ctx, job, Env.ResolveForRecord(commandTemplate))
	if createErr != nil {
		g.Log().Errorf(ctx, "create run failed: %+v", createErr)
		return consts.StatusFailed
	}
	fail := func(err error, output string) {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		finishRun(ctx, runId, consts.StatusFailed, "", start, output, "", 0, 0, msg)
		g.Log().Errorf(ctx, "job %s failed: %+v", job.Name, err)
		NewNotify().Notify(ctx, job.Name, "failed", job.NotifyOnSuccess == 1, job.NotifyOnFailure == 1, err)
	}

	// Execute command.
	g.Log().Infof(ctx, "job %s exec", job.Name)
	res, execErr := Env.Exec(ctx, command)
	output := truncate(res.Output)

	if execErr != nil {
		g.Log().Errorf(ctx, "job %s exec error: %+v", job.Name, execErr)
		fail(execErr, output)
		return consts.StatusFailed
	}
	if res.ExitCode != 0 {
		g.Log().Errorf(ctx, "job %s exit code %d", job.Name, res.ExitCode)
		fail(gerror.Newf("exit code %d: %s", res.ExitCode, strings.TrimSpace(output)), output)
		return consts.StatusFailed
	}
	g.Log().Infof(ctx, "job %s exec ok, output length: %d", job.Name, len(res.Output))

	// Calculate size and disk free using out_dir or fallback to BACKUP_DIR.
	sizeDir := job.OutDir
	if sizeDir == "" {
		sizeDir = Env.Get("BACKUP_DIR")
	}
	size := dirSizeMB(sizeDir)
	freeMB, _, _ := diskFreeMB(sizeDir)

	verifyStatus := consts.StatusSkipped
	if job.VerifyCommand != "" {
		vc := substituteCommand(kv, job.VerifyCommand)
		if job.OutDir != "" {
			vc = strings.ReplaceAll(vc, "<output_dir>", job.OutDir)
		}
		g.Log().Infof(ctx, "job %s verify", job.Name)
		vres, verr := Env.Exec(ctx, vc)
		if verr != nil || vres.ExitCode != 0 {
			verifyStatus = consts.StatusFailed
			verifyOutput := truncate(vres.Output)
			errMsg := fmt.Sprintf("verify failed: exit code %d", vres.ExitCode)
			if verr != nil {
				errMsg = fmt.Sprintf("verify failed: %v", verr)
			} else if vres.ExitCode != 0 {
				errMsg = fmt.Sprintf("verify failed: exit code %d, output: %s", vres.ExitCode, verifyOutput)
			}
			g.Log().Errorf(ctx, "job %s %s", job.Name, errMsg)
			finishRun(ctx, runId, consts.StatusFailed, verifyStatus, start, output,
				verifyOutput, size, freeMB, errMsg)
			NewNotify().Notify(ctx, job.Name, "verify-failed", job.NotifyOnSuccess == 1, job.NotifyOnFailure == 1, gerror.New(errMsg))
			return consts.StatusFailed
		}
		verifyStatus = consts.StatusOk
	}

	finishRun(ctx, runId, consts.StatusOk, verifyStatus, start, output, "", size, freeMB, "")
	g.Log().Infof(ctx, "job %s ok, elapsed %s", job.Name, time.Since(start))

	// Send success notification
	NewNotify().Notify(ctx, job.Name, "ok", job.NotifyOnSuccess == 1, job.NotifyOnFailure == 1, nil)

	// Retention cleanup: delete files older than retention_days in out_dir.
	if job.OutDir != "" && job.RetentionDays > 0 {
		g.Log().Infof(ctx, "job %s start retention cleanup, dir=%s, days=%d", job.Name, job.OutDir, job.RetentionDays)
		cleanRetention(ctx, job.OutDir, job.RetentionDays)
		g.Log().Infof(ctx, "job %s retention cleanup done", job.Name)
	}

	g.Log().Infof(ctx, "job %s(id=%d) finished, status=ok, elapsed=%s", job.Name, job.Id, time.Since(start))
	return consts.StatusOk
}

// dependencyFailed records a failed run for a job whose dependency did not
// succeed. It is created AFTER the dependency's own run, so the id order in
// history matches the actual execution order.
func (s *JobService) dependencyFailed(ctx context.Context, job *entity.Job, msg string) string {
	runId, err := createRun(ctx, job, job.Command)
	if err != nil {
		g.Log().Errorf(ctx, "create run failed: %+v", err)
		return consts.StatusFailed
	}
	finishRun(ctx, runId, consts.StatusFailed, "", time.Now(), "", msg, 0, 0, msg)
	g.Log().Errorf(ctx, "job %s failed: %s", job.Name, msg)
	NewNotify().Notify(ctx, job.Name, "failed", job.NotifyOnSuccess == 1, job.NotifyOnFailure == 1, gerror.New(msg))
	return consts.StatusFailed
}

// createRun inserts a running record and returns its id.
func createRun(ctx context.Context, job *entity.Job, command string) (int64, error) {
	res, err := dao.JobRuns.Ctx(ctx).Insert(map[string]any{
		"job_id":     job.Id,
		"status":     consts.StatusRunning,
		"started_at": now(),
		"command":    command,
		"error":      "",
		"created_at": now(),
	})
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// finishRun finalizes a run record.
func finishRun(ctx context.Context, runId int64, status, verifyStatus string, start time.Time,
	output, verifyOutput string, sizeMb, diskFreeMb float64, errMsg string) {
	data := map[string]any{
		"status":        status,
		"finished_at":   now(),
		"elapsed_ms":    time.Since(start).Milliseconds(),
		"output":        output,
		"verify_status": verifyStatus,
		"verify_output": verifyOutput,
		"size_mb":       sizeMb,
		"disk_free_mb":  diskFreeMb,
		"error":         errMsg,
	}
	if verifyStatus == "" {
		data["verify_status"] = consts.StatusSkipped
	}
	_, _ = dao.JobRuns.Ctx(ctx).Where("id", runId).Data(data).Update()
}

func truncate(s string) string {
	const (
		headLen = 4096
		tailLen = 4096
	)
	if len(s) <= headLen+tailLen {
		return s
	}
	return s[:headLen] + "\n...(truncated)...\n" + s[len(s)-tailLen:]
}

// dependencyOk no longer exists: depends_on now executes the dependency job first
// (see RunJob/runJob). Keeping this marker out of the file.
