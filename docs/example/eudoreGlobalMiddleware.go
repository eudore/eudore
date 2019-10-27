package main

/*
Eudore全局中间件会在路由匹配前执行，可以影响路由匹配数据。
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewEudore()
	app.RegisterInit("init-router", 0x015, func(app *eudore.Eudore) error {
		// 添加全局中间件修改请求，请求方法和路径固定位PUT和/。
		app.AddGlobalMiddleware(func(ctx eudore.Context) {
			ctx.Request().Method = "PUT"
			ctx.Request().URL.Path = "/"
		})
		// 添加路由详细。
		app.GetFunc("/*", func(ctx eudore.Context) {
			ctx.WriteString("hello eudore get")
		})
		app.AnyFunc("/*", func(ctx eudore.Context) {
			ctx.WriteString("hello eudore\n")
			ctx.WriteString("path is " + ctx.Path())
		})
		return nil
	})
	app.RegisterInit("init-listen", 0x016, func(app *eudore.Eudore) error {
		app.Listen(":8088")
		return nil
	})
	app.Run()
}
