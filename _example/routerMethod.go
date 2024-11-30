package main

/*
eudore.RouterCoreStd允许扩展容易注册方法，可以正确处理405场景。

RouterAllMethod 为RouterStd允许注册的全部方法。
RouterAnyMethod 为AnyFunc方法注册的全部方法。
NotFound, 404 	方法注册全局404处理方法
MethodNotAllowed, 405 方法注册全局405处理方法，默认实现响应返回Allow和X-Match-Route Header。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	// 定义Router允许的方法
	// 必须在NewRouter之前修改默认定义
	eudore.DefaultRouterAllMethod = append(eudore.DefaultRouterAllMethod, "MOVE", "COPY", "LOCK", "UNLOCK")
	eudore.DefaultRouterAnyMethod = append(eudore.DefaultRouterAnyMethod, "LOCK", "UNLOCK")

	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app))
	// 注册自定义方法
	app.AnyFunc("/eudore/debug/look/*", middleware.NewLookFunc(app))
	app.AnyFunc("/*path", eudore.HandlerEmpty)
	app.AddHandler("LOCK", "/*", eudore.HandlerEmpty)
	app.AddHandler("MOVE", "/*", eudore.HandlerEmpty)
	app.AddHandler("LOCK", "/dav/*", eudore.HandlerEmpty)
	app.AddHandler("UNLOCK", "/dav/*", eudore.HandlerEmpty)
	app.AddHandler("none", "/dav/*", eudore.HandlerEmpty)
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)

	// ---------- Any方法优先级 ----------
	// 通配符覆盖
	app.GetFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("method is get\n")
	})
	// Any方法不会覆盖Get方法。
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("method is any\n")
	})
	// 非Any方法覆盖Any方法。
	app.PostFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("method is post\n")
	})

	app.Listen(":8088")
	app.Run()
}
