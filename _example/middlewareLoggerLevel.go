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

	app.AddMiddleware(middleware.NewLoggerLevelFunc(nil))
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))

	app.AnyFunc("/api/v1/user", func(ctx eudore.Context) {
		ctx.Debug("Get User")
	})
	app.AnyFunc("/api/v1/meta", func(ctx eudore.Context) {
		ctx.Info("Get Meta", ctx.GetQuery("debug"))
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/api/v1/user").Do()
	client.NewRequest("GET", "/api/v1/meta?debug=0").Do()
	client.NewRequest("GET", "/api/v1/meta?debug=1").Do()
	client.NewRequest("GET", "/api/v1/meta?debug=5").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
