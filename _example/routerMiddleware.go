package main

/*
Router.AddMiddleware 会使用当前路由参数注册到路由核心，该行为是全局级别的，一次注册一直存在。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	// 第一个参数为字符串就是指定的路径。
	app.AddMiddleware("/api/", func(ctx eudore.Context) {
		ctx.WriteString("middleware /api/\n")
	})

	// 创建组路由
	apiv1 := app.Group("/api/v1")
	apiv1.AddMiddleware(func(ctx eudore.Context) {
		ctx.WriteString("group /api/v1 middleware/\n")
	})
	apiv1.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("request /api/v1")
	})

	// 默认路由
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString(ctx.Method() + " " + ctx.Path())
		ctx.WriteString("\nstar param: " + " " + ctx.GetParam("path"))
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do().Out()
	client.NewRequest("GET", "/api/v1/").Do().Out()
	client.NewRequest("GET", "/api/v1/eudore").Do().Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
