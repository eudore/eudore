package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	// 设置日志级别Info，使用Options方法加载Logger刷新pinrt函数使用的日志级别
	app.SetLevel(eudore.LogInfo)
	app.Options(app.Logger)

	app.AddMiddleware(func(ctx eudore.Context) {
		// 如果路由规则是"/api/v1/user",设置日志级别为Debug
		if ctx.GetParam("route") == "/api/v1/user" {
			log := ctx.Logger().WithFields(nil)
			log.SetLevel(eudore.LogDebug)
			ctx.SetLogger(log)
		}
	})
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	
	app.AnyFunc("/api/v1/user", func(ctx eudore.Context) {
		ctx.Debug("Get User")
	})
	app.AnyFunc("/api/v1/meta", func(ctx eudore.Context) {
		ctx.Debug("Get Meta")
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/api/v1/user").Do()
	client.NewRequest("GET", "/api/v1/meta").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
