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
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	// RouterStd方法扩展
	eudore.RouterAllMethod = append(eudore.RouterAllMethod, "MOVE", "COPY", "LOCK", "UNLOCK")
	eudore.RouterAnyMethod = append(eudore.RouterAnyMethod, "LOCK", "UNLOCK")

	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	// 通配符覆盖
	app.AnyFunc("/eudore/debug/look/*", middleware.NewLookFunc(app))
	app.AnyFunc("/*path", eudore.HandlerEmpty)
	app.AddHandler("LOCK", "/*", eudore.HandlerEmpty)
	app.AddHandler("MOVE", "/*", eudore.HandlerEmpty)
	app.AddHandler("LOCK", "/dav/*", eudore.HandlerEmpty)
	app.AddHandler("UNLOCK", "/dav/*", eudore.HandlerEmpty)
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("LOCK2", "/").Do().Out()
	client.NewRequest("LOCK", "/dav/1").Do().Out()
	client.NewRequest("COPY", "/dav/1").Do().Out()

	// 反注册 单元测试
	app.AddHandler("LOCK", "/* register=off", eudore.HandlerEmpty)
	app.AddHandler("ANY", "/* register=off", eudore.HandlerEmpty)
	app.AddHandler("MOVE", "/* register=off", eudore.HandlerEmpty)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
