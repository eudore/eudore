package eudore

import (
	"context"
	"database/sql"
)

// Database 定义数据库操作方法
type Database interface {
	AutoMigrate(interface{}) error
	Metadata(interface{}) interface{}
	Query(context.Context, interface{}, DatabaseStmt) error
	Exec(context.Context, DatabaseStmt) error
	Begin(ctx context.Context, opts *sql.TxOptions) (Database, error)
	Commit() error
	Rollback() error
}

// DatabaseStmt 定义数据库规则块。
type DatabaseStmt interface {
	Build(DatabaseBuilder)
}

// DatabaseBuilder 定义数据库sql构建者。
type DatabaseBuilder interface {
	Context() context.Context
	DriverName() string
	WriteStmts(...interface{})
	Result() (string, []interface{}, error)
}

// NewDatabaseStd 方法创建一个空Database。
func NewDatabaseStd(config interface{}) Database {
	return nil
}
