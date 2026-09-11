// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// NotifyConfigDao is the data access object for the table od_notify_config.
type NotifyConfigDao struct {
	table    string
	group    string
	columns  NotifyConfigColumns
	handlers []gdb.ModelHandler
}

// NotifyConfigColumns defines and stores column names for the table od_notify_config.
type NotifyConfigColumns struct {
	Id        string
	Channel   string
	Enabled   string
	Config    string
	UpdatedAt string
}

// notifyConfigColumns holds the columns for the table od_notify_config.
var notifyConfigColumns = NotifyConfigColumns{
	Id:        "id",
	Channel:   "channel",
	Enabled:   "enabled",
	Config:    "config",
	UpdatedAt: "updated_at",
}

// NewNotifyConfigDao creates and returns a new DAO object for table data access.
func NewNotifyConfigDao(handlers ...gdb.ModelHandler) *NotifyConfigDao {
	return &NotifyConfigDao{
		group:    gdb.DefaultGroupName,
		table:    "od_notify_config",
		columns:  notifyConfigColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *NotifyConfigDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *NotifyConfigDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *NotifyConfigDao) Columns() NotifyConfigColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *NotifyConfigDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *NotifyConfigDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *NotifyConfigDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
