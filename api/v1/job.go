package v1

import (
	"ops-dump/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Job module: CRUD plus enable/disable/run/run-history.

// JobFields holds the editable job attributes shared by create and update.
type JobFields struct {
	Name            string `v:"required" json:"name" dc:"名称"`
	Kind            string `v:"required" json:"kind" dc:"类型"`
	Enabled         int    `json:"enabled" dc:"是否启用"`
	SortOrder       int    `json:"sort_order" dc:"排序顺序"`
	Cron            string `v:"required" json:"cron" dc:"调度表达式"`
	DependsOn       string `json:"depends_on" dc:"依赖任务（先执行）"`
	Description     string `json:"description" dc:"描述"`
	Command         string `json:"command" dc:"命令"`
	VerifyCommand   string `json:"verify_command" dc:"校验命令"`
	OutDir          string `json:"out_dir" dc:"输出目录"`
	RetentionDays   int    `json:"retention_days" dc:"保留天数"`
	Kv              string `json:"kv" dc:"扩展参数(JSON)"`
	NotifyOnSuccess int    `json:"notify_on_success" dc:"成功时通知"`
	NotifyOnFailure int    `json:"notify_on_failure" dc:"失败时通知"`
}

type JobListReq struct {
	g.Meta `path:"jobs" method:"get" tags:"Job" summary:"任务列表"`
}
type JobListRes struct {
	List []entity.Job `json:"list" dc:"任务列表"`
}

type JobGetReq struct {
	g.Meta `path:"jobs/{id}" method:"get" tags:"Job" summary:"任务详情"`
	Id     int64 `v:"min:1#id错误" json:"id" dc:"任务ID"`
}
type JobGetRes struct {
	Job *entity.Job `json:"job" dc:"任务"`
}

type JobCreateReq struct {
	g.Meta `path:"jobs" method:"post" tags:"Job" summary:"创建任务"`
	JobFields
}
type JobCreateRes struct {
	Id int64 `json:"id" dc:"任务ID"`
}

type JobUpdateReq struct {
	g.Meta `path:"jobs/{id}" method:"put" tags:"Job" summary:"更新任务"`
	Id     int64 `v:"min:1#id错误" json:"id" dc:"任务ID"`
	JobFields
}
type JobUpdateRes struct{}

type JobDeleteReq struct {
	g.Meta `path:"jobs/{id}" method:"delete" tags:"Job" summary:"删除任务"`
	Id     int64 `v:"min:1#id错误" json:"id" dc:"任务ID"`
}
type JobDeleteRes struct{}

type JobEnableReq struct {
	g.Meta `path:"jobs/{id}/enable" method:"post" tags:"Job" summary:"启用任务"`
	Id     int64 `v:"min:1#id错误" json:"id" dc:"任务ID"`
}
type JobEnableRes struct{}

type JobDisableReq struct {
	g.Meta `path:"jobs/{id}/disable" method:"post" tags:"Job" summary:"禁用任务"`
	Id     int64 `v:"min:1#id错误" json:"id" dc:"任务ID"`
}
type JobDisableRes struct{}

type JobRunReq struct {
	g.Meta `path:"jobs/{id}/run" method:"post" tags:"Job" summary:"手动执行"`
	Id     int64 `v:"min:1#id错误" json:"id" dc:"任务ID"`
}
type JobRunRes struct {
	Status string `json:"status" dc:"执行状态"`
}

type JobRunsReq struct {
	g.Meta `path:"jobs/{id}/runs" method:"get" tags:"Job" summary:"运行记录"`
	Id     int64 `v:"min:1#id错误" json:"id" dc:"任务ID"`
}
type JobRunsRes struct {
	List []entity.JobRun `json:"list" dc:"运行记录"`
}

type JobRunsAllReq struct {
	g.Meta   `path:"runs" method:"get" tags:"Job" summary:"全部运行记录"`
	Limit    int    `v:"max:1000" json:"limit" dc:"数量限制"`
	Page     int    `json:"page" dc:"页码"`
	PageSize int    `json:"page_size" dc:"每页数量"`
	Status   string `json:"status" dc:"状态筛选"`
	JobName  string `json:"job_name" dc:"任务名称筛选"`
}
type JobRunsAllRes struct {
	List  []entity.JobRun `json:"list" dc:"运行记录"`
	Total int            `json:"total" dc:"总数"`
	Page  int            `json:"page" dc:"当前页"`
	Pages int            `json:"pages" dc:"总页数"`
}

type ClearRunsReq struct {
	g.Meta  `path:"runs/clear" method:"post" tags:"Job" summary:"清空执行历史（可按当前筛选条件）"`
	Status  string `json:"status" dc:"状态筛选"`
	JobName string `json:"job_name" dc:"任务名称筛选"`
}
type ClearRunsRes struct {
	Deleted int64 `json:"deleted" dc:"删除条数"`
}

type JobSortReq struct {
	g.Meta `path:"jobs/sort" method:"post" tags:"Job" summary:"任务排序"`
	Ids    []int64 `v:"required" json:"ids" dc:"任务ID列表（按排序顺序）"`
}
type JobSortRes struct{}

type JobExportReq struct {
	g.Meta `path:"jobs/export" method:"get" tags:"Job" summary:"导出任务"`
	Ids    []int64 `json:"ids" dc:"任务ID列表（空则导出全部）"`
}
type JobExportRes struct {
	Jobs []entity.Job `json:"jobs" dc:"任务列表"`
}

type JobImportReq struct {
	g.Meta `path:"jobs/import" method:"post" tags:"Job" summary:"导入任务"`
	Jobs   []JobFields `v:"required" json:"jobs" dc:"任务列表"`
}

type ImportResult struct {
	Name    string `json:"name" dc:"任务名称"`
	Status  string `json:"status" dc:"状态: ok/skipped/failed"`
	Message string `json:"message,omitempty" dc:"失败原因"`
}

type JobImportRes struct {
	Results []ImportResult `json:"results" dc:"导入结果"`
}
