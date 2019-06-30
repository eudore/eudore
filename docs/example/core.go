// curl -XGET http://localhost:8088/api/v1/
// curl -XGET http://localhost:8088/api/v1/get/eudore
// curl -XGET http://localhost:8088/api/v1/set/eudore
package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware/logger"
	"github.com/eudore/eudore/middleware/recover"
)

// eudore core
func main() {
	// 创建App
	app := eudore.NewCore()
	// 全局级请求处理中间件
	app.AddMiddleware("ANY", "",
		logger.NewLogger(),
	)

	// 创建子路由器
	apiv1 := app.Group("/api/v1 version:v1")
	// 路由级请求处理中间件
	apiv1.AddMiddleware("ANY", "", recover.RecoverFunc)
	{
		// Api级请求处理中间件, 常量优先于通配符
		apiv1.AnyFunc("/*", handlepre1, handleparam)
		apiv1.GetFunc("/get/:name", handleget)
	}
	// 默认路由
	app.AnyFunc("/*path", func(ctx eudore.Context) {
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
	ctx.WriteString("version: " + ctx.GetParam("version") + "\n")
}
func handleparam(ctx eudore.Context) {
	ctx.WriteString(ctx.GetParam("*"))
}
