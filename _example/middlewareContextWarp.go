package main

/*
在NewContextWarpFunc中间件之后的处理函数使用的eudore.Context对象为新的Context。
可以封装Context方法额外逻辑，或者重新实现一个ContextWarp中间件并加入新的自定义方法。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewContextWarpFunc(newContextParams))
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AnyFunc("/ctx", func(ctx eudore.Context) {
		index, handler := ctx.GetHandler()
		ctx.Debug(index, handler)
		ctx.SetHandler(index, handler)
	})
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debug("hello eudore")
		ctx.Info("hello eudore")
		ctx.End()
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do()
	client.NewRequest("GET", "/ctx").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

func newContextParams(ctx eudore.Context) eudore.Context {
	return contextParams{ctx}
}

type contextParams struct {
	eudore.Context
}

// GetParam 方法获取一个参数的值。
func (ctx contextParams) GetParam(key string) string {
	ctx.Debug("eudore.Context GetParam", key)
	return ctx.Context.GetParam(key)
}
