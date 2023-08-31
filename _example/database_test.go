package eudore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/eudore/eudore"
)

func TestDatabase(*testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyDatabase, &databaseTest{})
	app.SetValue(eudore.ContextKeyDatabaseRuntime, eudore.NewDatabaseRuntime)
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))

	app.GetFunc("/hello", func(ctx eudore.Context) {
		ctx.Query(nil, nil)
		ctx.Exec(nil)

		ctx.SetValue(eudore.ContextKeyDatabase, &databaseTest{Err: fmt.Errorf("test error")})
		ctx.Query(nil, nil)
		ctx.SetHeader(eudore.HeaderXTraceID, "id")
		ctx.Exec(nil)
	})

	app.NewRequest(nil, "GET", "/hello",
		strings.NewReader("trace"),
		eudore.NewClientCheckStatus(200),
	)

	app.CancelFunc()
	app.Run()
}

type databaseTest struct {
	eudore.Database
	Err error
}

func (db *databaseTest) Query(ctx context.Context, data interface{}, stmt eudore.DatabaseStmt) error {
	builder := &databaseBuilder{}
	stmt.Build(builder)
	return db.Err
}

func (db *databaseTest) Exec(ctx context.Context, stmt eudore.DatabaseStmt) error {
	builder := &databaseBuilder{}
	stmt.Build(builder)
	return db.Err
}

type databaseBuilder struct {
}

func (builder *databaseBuilder) Context() context.Context {
	return nil
}
func (builder *databaseBuilder) DriverName() string {
	return ""
}
func (builder *databaseBuilder) Metadata(interface{}) interface{} {
	return nil
}
func (builder *databaseBuilder) WriteStmts(...interface{}) {}
func (builder *databaseBuilder) Result() (string, []interface{}, error) {
	return "", nil, nil
}
