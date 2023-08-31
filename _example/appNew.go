package main

/*
eudore.App对象的简单组装各类对象，实现Value/SetValue、Listen和Run方法。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})
	app.AnyFunc("/data", func(ctx eudore.Context) interface{} {
		// 返回interface{}并直接Render
		return map[string]interface{}{
			"aa": 11,
			"bb": 22,
		}
	})

	app.Listen(":8088")
	app.Run()
}
