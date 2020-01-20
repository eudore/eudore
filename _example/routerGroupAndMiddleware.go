package main

/*
Router.Group 返回一个新的组路由，新路由器具有独立的参数和处理函数扩展。
Router.AddMiddleware 会使用当前路由参数注册到路由核心，该行为是全局级别的，一次注册一直存在。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewCore()
	app.AddMiddleware(middleware.NewLoggerFunc(app.App, "route"))

	// 创建组路由
	apiv1 := app.Group("/api/v1")
	apiv1.AddMiddleware(middleware.NewRecoverFunc())
	apiv1.AnyFunc("/*", handlepre1, handleparam)
	apiv1.GetFunc("/get/:name", handleget)

	// 默认路由
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString(ctx.Method() + " " + ctx.Path())
		ctx.WriteString("\nstar param: " + " " + ctx.GetParam("path"))
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/api/v1/").Do().Out()
	client.NewRequest("GET", "/api/v1/get/eudore").Do().Out()
	client.NewRequest("GET", "/api/v1/set/eudore").Do().Out()
	for client.Next() {
		app.Error(client.Error())
	}
	client.Stop(0)

	// 启动server
	app.Listen(":8088")
	app.Run()
}

func handleget(ctx eudore.Context) {
	ctx.Debug("Get: " + ctx.GetParam("name"))
	ctx.WriteString("Get: " + ctx.GetParam("name"))
}
func handlepre1(ctx eudore.Context) {
	ctx.WriteString("handlepre1\n")
}
func handleparam(ctx eudore.Context) {
	ctx.WriteString("handleparam: " + ctx.GetParam("*"))
}
