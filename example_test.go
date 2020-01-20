package eudore_test

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

// eudore core
func ExampleNewCore() {
	// 创建App
	app := eudore.NewCore()
	// 全局级请求处理中间件
	app.AddMiddleware(middleware.NewLoggerFunc(app.App))

	// 创建子路由器
	apiv1 := app.Group("/api/v1 version=v1")
	// 路由级请求处理中间件
	apiv1.AddMiddleware(middleware.NewRecoverFunc())
	{
		// Api级请求处理中间件, 常量优先于通配符
		apiv1.AnyFunc("/*", handlepre1, handleparam)
		apiv1.GetFunc("/get/:name", handleget)
	}
	// 默认路由
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString(ctx.Method() + " " + ctx.Path())
		ctx.WriteString("\nstar param: " + " " + ctx.GetParam("path"))
	})
	// 启动server
	app.Listen(":8088")
	app.Run()
}
func handleget(ctx eudore.Context) {
	ctx.Debug("Get: " + ctx.GetParam("name"))
	ctx.WriteString("Get: " + ctx.GetParam("name"))
}
func handlepre1(ctx eudore.Context) {
	// 添加参数
	ctx.WriteString("handlepre1\n")
	ctx.WriteString("version: " + ctx.GetParam("version") + "\n")
}
func handleparam(ctx eudore.Context) {
	ctx.WriteString(ctx.GetParam("*"))
}

// router
func ExampleNewRouterRadix() {
	r := eudore.NewRouterRadix()
	// Or the path is /api/v1/*path
	// 或者路径是 /api/v1/*path
	r.AnyFunc("/api/v1/*", func(ctx eudore.Context) {
		ctx.WriteString(ctx.GetParam("*"))
	})
	r.GetFunc("/api/v1/info/:name action=showname version=v1", func(ctx eudore.Context) {
		// Get route additional parameters and path parameters
		// 获取路由附加参数和路径参数
		ctx.WithField("route version", ctx.GetParam("version")).Info("user name is: " + ctx.GetParam("name"))
	})
}
