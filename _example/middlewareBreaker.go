package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)

	// 创建熔断器并注入管理路由
	app.AddMiddleware(middleware.NewCircuitBreaker(app.Group("/eudore/debug/breaker")).NewBreakFunc())

	app.GetFunc("/*", echo)
	app.Listen(":8088")
	app.Run()
}

func echo(ctx eudore.Context) {
	if len(ctx.Querys()) > 0 {
		ctx.Fatal("test err")
		return
	}
	ctx.WriteString("route: " + ctx.GetParam("route"))
}

// 页面地址 ip:8088/eudore/debug/breaker/ui
// 每个路由访问过后才会显示
