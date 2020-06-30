package main

/*
Router.Group 返回一个新的组路由，新路由器具有独立的参数和处理函数扩展。
Group后新增参数、中间件、扩展均不会响应到原路由器。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))

	// 创建组路由
	apiv1 := app.Group("/api/v1")
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

	app.Listen(":8088")
	// app.CancelFunc()
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
