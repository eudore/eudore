package main

/*
app.AddMiddleware 方法如果第一个参数为字符串"global",
则作为全局请求中间件添加给App(使用DefaultHandlerExtend创建请求处理函数),
否则等同于调用app.Router.AddMiddleware方法。

全局中间件会在路由匹配之前执行。
*/

import (
	"errors"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	if app.Parse() != nil {
		return
	}

	app.AddMiddleware("global",
		middleware.NewLoggerFunc(app),
		middleware.NewRequestIDFunc(nil),
		middleware.NewRecoveryFunc(),
	)
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debug(ctx.Request().RemoteAddr)
		ctx.WriteString("hello eudore")
	})
	app.AnyFunc("/err", func(ctx eudore.Context) error {
		return errors.New("errors")
	})
	app.AnyFunc("/data", func(ctx eudore.Context) any {
		// any并直接Render
		return map[string]any{
			"aa": 11,
			"bb": 22,
		}
	})
	app.GetRequest("/")

	app.Listen(":8088")
	app.Run()
}
