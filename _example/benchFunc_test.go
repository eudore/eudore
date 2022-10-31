package eudore_test

import (
	"reflect"
	"testing"
)

type Stmt interface {
	Init(DatabaseContext) error
	Build(DatabaseContext)
}
type DatabaseContext interface{}

type StmtSelect struct {
	Name string
}

func (stmt *StmtSelect) Init(DatabaseContext) error {
	_ = stmt.Name
	return nil
}
func (stmt *StmtSelect) Build(DatabaseContext) {

}

func BenchmarkFuncReflect(b *testing.B) {
	fn := reflect.ValueOf(func(ctx DatabaseContext, stmt *StmtSelect) {
		stmt.Init(ctx)
		stmt.Build(ctx)
	})
	var ctx DatabaseContext = 0
	var stmt = &StmtSelect{"eudore"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		reflect.TypeOf(stmt)
		fn.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(stmt)})
	}
}

func BenchmarkFuncAassertions(b *testing.B) {
	fn := func(ctx DatabaseContext, stmt Stmt) {
		stmtSelect := stmt.(*StmtSelect)
		stmtSelect.Init(ctx)
		stmtSelect.Build(ctx)
	}
	var ctx DatabaseContext = 0
	var stmt = &StmtSelect{"eudore"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		reflect.TypeOf(stmt)
		fn(ctx, stmt)
	}
}
