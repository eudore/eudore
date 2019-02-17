package eudore_test

import (
	"eudore"
	"eudore/middleware/gzip"
	"eudore/middleware/recover"
)

// eudore core
func ExampleNewCore() {
	// 创建App
	e := eudore.NewCore()
	// 全局级请求处理中间件
	e.AddHandler(gzip.NewGzip(5))

	// 创建子路由器
	apiv1 := eudore.NewRouterMust("", nil)
	// 路由级请求处理中间件
	apiv1.AddHandler(eudore.HandlerFunc(recover.RecoverFunc))
	{
		apiv1.GetFunc("/get", handleget)
		// Api级请求处理中间件
		apiv1.Any("/", eudore.NewMiddlewareLink(
			eudore.HandlerFunc(handlepre1),
			eudore.HandlerFunc(handleget),
		))
	}
	// app注册api子路由
	e.SubRoute("/api/v1 version:v1", apiv1)

	// 启动server
	e.Run()
}

func handlepre1(ctx eudore.Context) {
	ctx.AddParam("pre1", "1")
	ctx.AddParam("pre1", "2")
}
func handleget(ctx eudore.Context) {
	ctx.WriteJson(ctx.Params())
}

// router
func ExampleNewStdRouter() {
	r, _ := eudore.NewRouterStd(nil)
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
