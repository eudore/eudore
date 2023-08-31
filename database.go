package eudore

import (
	"context"
	"database/sql"
)

// Database 定义数据库操作方法。
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
	Metadata(interface{}) interface{}
	WriteStmts(...interface{})
	Result() (string, []interface{}, error)
}

// NewDatabase 方法创建一个空Database。
func NewDatabase(interface{}) Database {
	return nil
}

type stmtContextRuntime struct {
	Context Context
	Stmt    DatabaseStmt
}

func NewDatabaseRuntime(ctx Context, stmt DatabaseStmt) DatabaseStmt {
	return stmtContextRuntime{ctx, stmt}
}

var conetxtIDKeys = [...]string{HeaderXTraceID, HeaderXRequestID}

func (stmt stmtContextRuntime) Build(builder DatabaseBuilder) {
	h := stmt.Context.Response().Header()
	for _, key := range conetxtIDKeys {
		id := h.Get(key)
		if id != "" {
			builder.WriteStmts("-- "+id+"\r\n", stmt.Stmt)
			return
		}
	}
}
