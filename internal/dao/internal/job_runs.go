// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// JobRunsDao is the data access object for the table od_job_runs.
type JobRunsDao struct {
	table    string
	group    string
	columns  JobRunsColumns
	handlers []gdb.ModelHandler
}

// JobRunsColumns defines and stores column names for the table od_job_runs.
type JobRunsColumns struct {
	Id           string
	JobId        string
	Status       string
	StartedAt    string
	FinishedAt   string
	ElapsedMs    string
	ExitCode     string
	Error        string
	Command      string
	Output       string
	VerifyStatus string
	VerifyOutput string
	SizeMb       string
	DiskFreeMb   string
	CreatedAt    string
}

// jobRunsColumns holds the columns for the table od_job_runs.
var jobRunsColumns = JobRunsColumns{
	Id:           "id",
	JobId:        "job_id",
	Status:       "status",
	StartedAt:    "started_at",
	FinishedAt:   "finished_at",
	ElapsedMs:    "elapsed_ms",
	ExitCode:     "exit_code",
	Error:        "error",
	Command:      "command",
	Output:       "output",
	VerifyStatus: "verify_status",
	VerifyOutput: "verify_output",
	SizeMb:       "size_mb",
	DiskFreeMb:   "disk_free_mb",
	CreatedAt:    "created_at",
}

// NewJobRunsDao creates and returns a new DAO object for table data access.
func NewJobRunsDao(handlers ...gdb.ModelHandler) *JobRunsDao {
	return &JobRunsDao{
		group:    gdb.DefaultGroupName,
		table:    "od_job_runs",
		columns:  jobRunsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *JobRunsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *JobRunsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *JobRunsDao) Columns() JobRunsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *JobRunsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *JobRunsDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *JobRunsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
