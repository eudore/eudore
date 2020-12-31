package main

/*
AddMiddleware 方法如果第一个参数为字符串"global",则作为全局请求中间件添加给App(使用DefaultHandlerExtend创建请求处理函数),否则等同于调用app.Rputer.AddMiddleware方法。
func (app *App) AddMiddleware(hs ...interface{}) error {
	if len(hs) > 1 {
		name, ok := hs[0].(string)
		if ok && name == "global" {
			handler := DefaultHandlerExtend.NewHandlerFuncs("", hs[1:])
			app.Info("Register app global middleware:", handler)
			app.HandlerFuncs = HandlerFuncsCombine(app.HandlerFuncs[0:len(app.HandlerFuncs)-1], handler)
			app.HandlerFuncs = HandlerFuncsCombine(app.HandlerFuncs, HandlerFuncs{app.serveContext})
			return nil
		}
	}
	return app.Router.AddMiddleware(hs...)
}
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware("global", func(ctx eudore.Context) {
		ctx.Request().Method = "GET"
	})
	app.AddMiddleware(middleware.NewLoggerFunc(app))
	app.GetFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("xxx", "/1").Do()
	client.NewRequest("POST", "/1").Do()
	client.NewRequest("PUT", "/1").Do()
	client.NewRequest("OPTIONS", "/1").Do()
	client.NewRequest("OPTIONS", "/1").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
