package main

/*
app.AddMiddleware 方法如果第一个参数为字符串"global",
则作为全局请求中间件添加给App(使用DefaultHandlerExtend创建请求处理函数),
否则等同于调用app.Rputer.AddMiddleware方法。

全局中间件会在路由匹配之前执行。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware("global", func(ctx eudore.Context) {
		// 强制修改请求方法为PUT
		ctx.Request().Method = "PUT"
	})
	app.AddMiddleware(middleware.NewLoggerFunc(app))
	app.GetFunc("/*", eudore.HandlerEmpty)

	app.Listen(":8088")
	app.Run()
}
