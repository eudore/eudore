package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewCore()

	// 创建熔断器并注入管理路由
	cb := middleware.NewCircuitBreaker()
	cb.InjectRoutes(app.Group("/eudore/debug/breaker"))
	app.AddMiddleware(eudore.MethodAny, "", cb.Handle)

	app.GetFunc("/*", echo)
	app.Listen(":8088")
	app.Run()
}

func echo(ctx eudore.Context) {
	if ctx.Querys().Len() > 0 {
		ctx.Fatal("test err")
		return
	}
	ctx.WriteString("route: " + ctx.GetParam("route"))
}

// 页面地址 ip:8088/eudore/debug/breaker/ui
// 每个路由访问过后才会显示
