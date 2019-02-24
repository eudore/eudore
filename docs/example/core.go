// curl -XGET http://localhost:8088/api/v1/
// curl -XGET http://localhost:8088/api/v1/get/eudore
package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware/logger"
	// "github.com/eudore/eudore/middleware/gzip"
	"github.com/eudore/eudore/middleware/recover"
)

// eudore core
func main() {
	// 创建App
	app := eudore.NewCore()
	app.RegisterComponent("logger-std", &eudore.LoggerStdConfig{
		Std:	true,
		Level:	eudore.LogDebug,
		Format:	"json",
	})
	// 全局级请求处理中间件
	app.AddHandler(
		logger.NewLogger(eudore.GetRandomString),
		// gzip.NewGzip(5),
	)

	// 创建子路由器
	// apiv1 := eudore.NewRouterClone(app.Router)
	apiv1 := app.Group("/api/v1")
	// 路由级请求处理中间件
	apiv1.AddHandler(eudore.HandlerFunc(recover.RecoverFunc))
	{
		apiv1.GetFunc("/get/:name", handleget)
		// Api级请求处理中间件
		apiv1.Any("/*", eudore.NewMiddlewareLink(
			eudore.HandlerFunc(handlepre1),
			eudore.HandlerFunc(handleparam),
		))
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
	// 将ctx的参数以Json格式返回
	// ctx.WriteJson(ctx.Params())
	// 将ctx的参数根据请求格式返回
	// ctx.WriteRender(ctx.Params())
}