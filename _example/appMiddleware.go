package main

/*
app.AddMiddleware 第一参数为字符串"global"，则中间件为全局中间件会在路由匹配之前执行，否在作为路由器中间件添加。
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
