package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	// 创建熔断器并注入管理路由
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(middleware.NewCircuitBreaker(app.Group("/eudore/debug/breaker")).NewBreakFunc())
	app.GetFunc("/*", echo)

	client := httptest.NewClient(app)
	// 错误请求
	for i := 0; i < 15; i++ {
		client.NewRequest("GET", "/1?a=1").Do()
	}
	// 除非熔断后访问
	for i := 0; i < 5; i++ {
		client.NewRequest("GET", "/1").Do()
	}
	for client.Next() {
		app.Error(client.Error())
	}

	app.CancelFunc()
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
