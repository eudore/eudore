package main

/*
eudore.RouterStd允许注册Test方法查看添加路径的切割方法和处理函数,用于验证路由规则注册结果。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp(
		eudore.NewRouterStd(eudore.NewRouterCoreDebug(nil)),
		eudore.NewRouterStd(eudore.NewRouterCoreDebug(eudore.NewRouterCoreStd())),
		eudore.Renderer(eudore.RenderJSON),
	)

	api := app.Group("/api/{v 1} version=v1")
	api.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	api.AddHandler("test", "/get/:name action=GetName", eudore.HandlerEmpty)
	api.AddHandler("test", "/get/{{}} action=GetName", eudore.HandlerEmpty)
	app.AddHandler("TEST", "/api/v:v/user/*name", eudore.HandlerEmpty)
	api.AddHandler("GET", "/ui", eudore.HandlerEmpty)
	api.AddHandler("GET", "/get/:name action=GetName", eudore.HandlerEmpty)
	api.AddHandler("GET", "/get/{{}} action=GetName", eudore.HandlerEmpty)
	app.AddHandler("GET", "/api/v:v/user/*name", eudore.HandlerEmpty)
	api.AddHandler("GET", "/ui")
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/eudore/debug/router/data").Do().OutBody()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
