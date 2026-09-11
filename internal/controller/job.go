package controller

import (
	"context"

	v1 "ops-dump/api/v1"
	"ops-dump/internal/model"
	"ops-dump/internal/service"
)

// JobController handles job CRUD and lifecycle operations.
type JobController struct{}

func NewJob() *JobController { return &JobController{} }

func (c *JobController) List(ctx context.Context, req *v1.JobListReq) (res *v1.JobListRes, err error) {
	jobs, err := service.NewJob().ListJobs(ctx)
	if err != nil {
		return
	}
	res = &v1.JobListRes{List: jobs}
	return
}

func (c *JobController) Get(ctx context.Context, req *v1.JobGetReq) (res *v1.JobGetRes, err error) {
	job, err := service.NewJob().GetJob(ctx, req.Id)
	if err != nil {
		return
	}
	res = &v1.JobGetRes{Job: job}
	return
}

func (c *JobController) Create(ctx context.Context, req *v1.JobCreateReq) (res *v1.JobCreateRes, err error) {
	id, err := service.NewJob().CreateJob(ctx, toJobParam(req.JobFields, 0))
	if err == nil {
		err = service.NewJob().SchedulerReload(ctx)
	}
	if err != nil {
		return
	}
	res = &v1.JobCreateRes{Id: id}
	return
}

func (c *JobController) Update(ctx context.Context, req *v1.JobUpdateReq) (res *v1.JobUpdateRes, err error) {
	err = service.NewJob().UpdateJob(ctx, toJobParam(req.JobFields, req.Id))
	if err == nil {
		err = service.NewJob().SchedulerReload(ctx)
	}
	return
}

func (c *JobController) Delete(ctx context.Context, req *v1.JobDeleteReq) (res *v1.JobDeleteRes, err error) {
	err = service.NewJob().DeleteJob(ctx, req.Id)
	if err == nil {
		err = service.NewJob().SchedulerReload(ctx)
	}
	return
}

func (c *JobController) Enable(ctx context.Context, req *v1.JobEnableReq) (res *v1.JobEnableRes, err error) {
	err = service.NewJob().SetJobEnabled(ctx, req.Id, 1)
	return
}

func (c *JobController) Disable(ctx context.Context, req *v1.JobDisableReq) (res *v1.JobDisableRes, err error) {
	err = service.NewJob().SetJobEnabled(ctx, req.Id, 0)
	return
}

func (c *JobController) Run(ctx context.Context, req *v1.JobRunReq) (res *v1.JobRunRes, err error) {
	status, err := service.NewJob().ManualRun(ctx, req.Id)
	if err != nil {
		return
	}
	res = &v1.JobRunRes{Status: status}
	return
}

func (c *JobController) Runs(ctx context.Context, req *v1.JobRunsReq) (res *v1.JobRunsRes, err error) {
	runs, err := service.NewJob().ListRuns(ctx, req.Id, 50)
	if err != nil {
		return
	}
	res = &v1.JobRunsRes{List: runs}
	return
}

func (c *JobController) AllRuns(ctx context.Context, req *v1.JobRunsAllReq) (res *v1.JobRunsAllRes, err error) {
	result, err := service.NewJob().ListAllRuns(ctx, service.ListAllRunsParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
		JobName:  req.JobName,
	})
	if err != nil {
		return
	}
	res = &v1.JobRunsAllRes{
		List:  result.List,
		Total: result.Total,
		Page:  result.Page,
		Pages: result.Pages,
	}
	return
}

func (c *JobController) ClearRuns(ctx context.Context, req *v1.ClearRunsReq) (res *v1.ClearRunsRes, err error) {
	n, err := service.NewJob().ClearRuns(ctx, req.Status, req.JobName)
	if err != nil {
		return
	}
	res = &v1.ClearRunsRes{Deleted: n}
	return
}

func (c *JobController) Sort(ctx context.Context, req *v1.JobSortReq) (res *v1.JobSortRes, err error) {
	err = service.NewJob().SortJobs(ctx, req.Ids)
	return
}

func (c *JobController) Export(ctx context.Context, req *v1.JobExportReq) (res *v1.JobExportRes, err error) {
	jobs, err := service.NewJob().ExportJobs(ctx, req.Ids)
	if err != nil {
		return
	}
	res = &v1.JobExportRes{Jobs: jobs}
	return
}

func (c *JobController) Import(ctx context.Context, req *v1.JobImportReq) (res *v1.JobImportRes, err error) {
	params := make([]model.JobParam, 0, len(req.Jobs))
	for _, j := range req.Jobs {
		params = append(params, toJobParam(j, 0))
	}
	results, err := service.NewJob().ImportJobs(ctx, params)
	if err != nil {
		return
	}
	importResults := make([]v1.ImportResult, 0, len(results))
	for _, r := range results {
		importResults = append(importResults, v1.ImportResult{
			Name:    r.Name,
			Status:  r.Status,
			Message: r.Message,
		})
	}
	res = &v1.JobImportRes{Results: importResults}
	return
}

func toJobParam(f v1.JobFields, id int64) model.JobParam {
	return model.JobParam{
		Id:              id,
		Name:            f.Name,
		Kind:            f.Kind,
		Enabled:         f.Enabled,
		SortOrder:       f.SortOrder,
		Cron:            f.Cron,
		DependsOn:       f.DependsOn,
		Description:     f.Description,
		Command:         f.Command,
		VerifyCommand:   f.VerifyCommand,
		OutDir:          f.OutDir,
		RetentionDays:   f.RetentionDays,
		Kv:              f.Kv,
		NotifyOnSuccess: f.NotifyOnSuccess,
		NotifyOnFailure: f.NotifyOnFailure,
	}
}
