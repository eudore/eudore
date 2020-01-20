package main

/*
Eudore全局中间件会在路由匹配前执行，可以影响路由匹配数据。

全局中间件会在路由匹配前执行，不会存在路由详细、也可以修改基础信息影响路由匹配。

ServeHTTP时先设置请求上下文的处理函数是全部全局中间件函数处理请求。
最后一个全局中间件函数是app.HandleContext，该方法才会匹配路由请求，然后重新调用ctx.SetHandler方法设置多个请求处理函数是路由匹配后的结果。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewEudore()
	httptest.NewClient(app).Stop(0)
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
		return app.Listen(":8088")
	})
	app.Run()
}
