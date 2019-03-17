package eudore_test

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware/gzip"
	"github.com/eudore/eudore/middleware/recover"
	"github.com/eudore/eudore/middleware/logger"
)

// eudore core
func ExampleNewCore() {
	// 创建App
	app := eudore.NewCore()
	app.RegisterComponent("logger-std", &eudore.LoggerStdConfig{
		Std:	true,
		Level:	eudore.LogDebug,
		Format:	"json",
	})
	// 全局级请求处理中间件
	app.AddMiddleware(
		logger.NewLogger(eudore.GetRandomString).Handle,
		gzip.NewGzip(5).Handle,
	)

	// 创建子路由器
	// apiv1 := eudore.NewRouterClone(app.Router)
	apiv1 := app.Group("/api/v1")
	// 路由级请求处理中间件
	apiv1.AddMiddleware(recover.RecoverFunc)
	{
		apiv1.GetFunc("/get/:name", handleget)
		// Api级请求处理中间件
		apiv1.AnyFunc("/*", handlepre1, handleparam)
	}
	// app注册api子路由
	// app.SubRoute("/api/v1 version:v1", apiv1)
	// 默认路由
	app.AnyFunc("/*path", func(ctx eudore.Context){
		ctx.WriteString(ctx.Method() + " " + ctx.Path())
		ctx.WriteString("\nstar param: " + " " + ctx.GetParam("path"))
	})
	// 启动server
	app.Listen(":8088")
	app.Run()
}
func handleget(ctx eudore.Context) {
	ctx.Debug("Get: " + ctx.GetParam("name"))
	ctx.WriteString("Get: " + ctx.GetParam("name"))
}
func handlepre1(ctx eudore.Context) {
	// 添加参数
	ctx.WriteString("handlepre1\n")
}
func handleparam(ctx eudore.Context) {
	ctx.WriteString(ctx.GetParam("*"))
}

// router
func ExampleNewRouterRadix() {
	r, _ := eudore.NewRouterRadix(nil)
	// Or the path is /api/v1/*path
	// 或者路径是 /api/v1/*path
	r.AnyFunc("/api/v1/*", func(ctx eudore.Context) {
		ctx.WriteString(ctx.GetParam("*"))
	})
	r.GetFunc("/api/v1/info/:name action:showname version:v1", func(ctx eudore.Context){
		// Get route additional parameters and path parameters
		// 获取路由附加参数和路径参数
		ctx.WithField("version", ctx.GetParam("version")).Info("user name is: " + ctx.GetParam("name"))
	})
}
