package main

/*
Router.Group 创建一个组路由深复制，新路由器具有独立的参数、中间件、处理函数扩展。
Group后新增参数、中间件、扩展均不会响应到原路由器。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()

	// 创建组路由
	apiv1 := app.Group("/api/v1")
	apiv1.AnyFunc("/*", handlerPre1, handlerParam)
	apiv1.GetFunc("/get/:name", handlerGet)

	// 组路由追加参数
	apiv2 := app.Group("/api/v2 v=v2")
	apiv2.AddMiddleware(middleware.NewLoggerFunc(app))
	apiv2.AnyFunc("/*", handlerPre1, handlerParam)
	// 获取组路由参数
	app.Info(apiv2.Params())

	app.Listen(":8088")
	app.Run()
}

func handlerGet(ctx eudore.Context) {
	ctx.Debug("Get: " + ctx.GetParam("name"))
	ctx.WriteString("Get: " + ctx.GetParam("name"))
}

func handlerPre1(ctx eudore.Context) {
	ctx.WriteString("handler pre1\n")
}

func handlerParam(ctx eudore.Context) {
	ctx.WriteString("handler params: " + ctx.Params().String())
}
