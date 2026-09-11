// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SessionsDao is the data access object for the table od_sessions.
type SessionsDao struct {
	table    string
	group    string
	columns  SessionsColumns
	handlers []gdb.ModelHandler
}

// SessionsColumns defines and stores column names for the table od_sessions.
type SessionsColumns struct {
	Id         string
	Username   string
	Expires    string
	CreatedAt  string
}

// sessionsColumns holds the columns for the table od_sessions.
var sessionsColumns = SessionsColumns{
	Id:        "id",
	Username:  "username",
	Expires:   "expires",
	CreatedAt: "created_at",
}

// NewSessionsDao creates and returns a new DAO object for table data access.
func NewSessionsDao(handlers ...gdb.ModelHandler) *SessionsDao {
	return &SessionsDao{
		group:    gdb.DefaultGroupName,
		table:    "od_sessions",
		columns:  sessionsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SessionsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SessionsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SessionsDao) Columns() SessionsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SessionsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SessionsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SessionsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
