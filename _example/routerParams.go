package main

/*
type RouterMethod interface {
	GetParam(string) string
	SetParam(string, string) Router
	...
}
Router可以使用GetParam和SetParam方法读写当前路器由参数。

在Router.Group时，新路由器会复制一份上级路由器参数，同时处理路径中的参数。

在Router.Macth后，默认的路由参数和匹配参数会添加到Context中。

在Router中使用键eudore.ParamAllKeys/eudore.ParamAllVals可以获取到全部参数key/val，返回多个值使用空格分割。

在Router和Context中"route"参数表示当前路由路径
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

type Parmaser interface {
	Params() eudore.Params
}

func main() {
	app := eudore.NewCore()
	app.AddMiddleware(middleware.NewLoggerFunc(app.App, "route"))

	apiv1 := app.Group("/api/v1 version=v1")
	apiv1.AnyFunc("/*", starParam)
	apiv1.GetFunc("/get/:name action=getParamName", getParamName)
	app.Debug("all param keys:", apiv1.GetParam(eudore.ParamAllKeys))
	app.Debug("all param vals:", apiv1.GetParam(eudore.ParamAllVals))
	app.Debugf("parmas: %#v", apiv1.(Parmaser).Params())

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
	app.Run()
}

func getParamName(ctx eudore.Context) {
	ctx.Debug("Get: "+ctx.GetParam("name"), "route:", ctx.GetParam("route"))
	ctx.WriteString("Get: " + ctx.GetParam("name"))
}
func starParam(ctx eudore.Context) {
	ctx.WriteString(ctx.GetParam("*"))
}
