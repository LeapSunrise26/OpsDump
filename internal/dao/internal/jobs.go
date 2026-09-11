// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// JobsDao is the data access object for the table od_jobs.
type JobsDao struct {
	table    string
	group    string
	columns  JobsColumns
	handlers []gdb.ModelHandler
}

// JobsColumns defines and stores column names for the table od_jobs.
type JobsColumns struct {
	Id            string
	Name          string
	Kind          string
	Enabled       string
	Cron          string
	DependsOn     string
	Description   string
	Container     string
	Database      string
	Command       string
	VerifyCommand string
	MinFreeMb     string
	OutDir        string
	RetentionDays string
	Kv            string
	CreatedAt     string
	UpdatedAt     string
}

// jobsColumns holds the columns for the table od_jobs.
var jobsColumns = JobsColumns{
	Id:            "id",
	Name:          "name",
	Kind:          "kind",
	Enabled:       "enabled",
	Cron:          "cron",
	DependsOn:     "depends_on",
	Description:   "description",
	Container:     "container",
	Database:      "database",
	Command:       "command",
	VerifyCommand: "verify_command",
	MinFreeMb:     "min_free_mb",
	OutDir:        "out_dir",
	RetentionDays: "retention_days",
	Kv:            "kv",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewJobsDao creates and returns a new DAO object for table data access.
func NewJobsDao(handlers ...gdb.ModelHandler) *JobsDao {
	return &JobsDao{
		group:    gdb.DefaultGroupName,
		table:    "od_jobs",
		columns:  jobsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *JobsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *JobsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *JobsDao) Columns() JobsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *JobsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *JobsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *JobsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
