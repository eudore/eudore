package eudore

import (
	"context"
)

// Database 定义数据库操作方法
type Database interface {
	WithContext(context.Context) Database
	AddHook(interface{})
	AutoMigrate(interface{}) error
	Query(interface{}, Stmt) error
	Exec(Stmt) error
}

// DatabaseContext 定义数据上下文
type DatabaseContext interface {
	Context() context.Context
	WriteString(string)
	WriteValue(interface{})
	WriteStmt(Stmt)
}

// Stmt 定义sql statement接口
type Stmt interface {
	Init(DatabaseContext) error
	Build(DatabaseContext)
}

// NewDatabaseStd 方法创建一个空Database。
func NewDatabaseStd(config interface{}) Database {
	return nil
}
