package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
	"time"
)

func main() {
	app := eudore.NewApp()
	// 创建熔断器并注入管理路由
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(middleware.NewBreaker().InjectRoutes(app.Group("/eudore/debug")).NewBreakFunc())
	app.GetFunc("/*", echo)

	client := httptest.NewClient(app)
	// 错误请求
	for i := 0; i < 15; i++ {
		client.NewRequest("GET", "/1?a=1").Do()
	}
	// 除非熔断后访问
	for i := 0; i < 15; i++ {
		time.Sleep(time.Millisecond * 500)
		client.NewRequest("GET", "/1").Do()
	}
	client.NewRequest("GET", "/eudore/debug/breaker/ui").Do()
	middleware.BreakerStaticHTML = ""
	client.NewRequest("GET", "/eudore/debug/breaker/ui").Do()
	client.NewRequest("GET", "/eudore/debug/breaker/data").WithHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do().OutBody()
	client.NewRequest("GET", "/eudore/debug/breaker/0").Do()
	client.NewRequest("GET", "/eudore/debug/breaker/100").Do()
	client.NewRequest("PUT", "/eudore/debug/breaker/0/state/0").Do()
	client.NewRequest("PUT", "/eudore/debug/breaker/0/state/3").Do()
	client.NewRequest("PUT", "/eudore/debug/breaker/3/state/3").Do()

	app.Listen(":8088")
	// app.CancelFunc()
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
