package main

/*
type RouterMethod interface {
	Params() *Params
	...
}
Router可以使用Params方法获取当前路器由参数。

在Router.Group时，新路由器会复制一份上级路由器参数，同时处理路径中的参数。

在Router.Macth后，默认的路由参数和匹配参数会添加到Context中。

在Router和Context中"route"参数表示当前路由路径
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))

	apiv1 := app.Group("/api/v1 version=v1")
	apiv1.AnyFunc("/*", starParam)
	apiv1.GetFunc("/get/:name action=getParamName", getParamName)
	app.Debug("all param:", apiv1.Params())

	// 默认路由
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString(ctx.Method() + " " + ctx.Path())
		ctx.WriteString("\nstar param: " + " " + ctx.GetParam("path"))
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/api/v1/").Do().Out()
	client.NewRequest("GET", "/api/v1/get/eudore").Do().Out()
	client.NewRequest("GET", "/api/v1/set/eudore").Do().Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

func getParamName(ctx eudore.Context) {
	ctx.Debug("Get: "+ctx.GetParam("name"), "route:", ctx.GetParam("route"))
	ctx.WriteString("Get: " + ctx.GetParam("name"))
}
func starParam(ctx eudore.Context) {
	ctx.WriteString(ctx.GetParam("*"))
}
