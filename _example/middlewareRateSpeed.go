package main

import (
	"context"
	"net/http"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	RateSpeedTimeout()
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewRateSpeedFunc(16*1024, 64*1024, app.Context))
	app.PostFunc("/post", func(ctx eudore.Context) {
		ctx.Debug(string(ctx.Body()))
	})
	app.AnyFunc("/srv", func(ctx eudore.Context) {
		ctx.WriteString("rate speed 16kB")
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("POST", "/post").WithBody("return body").Do().CheckStatus(200)
	client.NewRequest("PUT", "/srv").Do().CheckStatus(200)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()

}

func RateSpeedTimeout() {
	app := eudore.NewApp()
	app.SetHandler(http.TimeoutHandler(app, 2*time.Second, ""))

	// /done限速512B
	app.PostFunc("/done", func(ctx eudore.Context) {
		c, cannel := context.WithCancel(ctx.GetContext())
		ctx.WithContext(c)
		cannel()
	}, middleware.NewRateSpeedFunc(512, 1024, app.Context), func(ctx eudore.Context) {
		ctx.Debug(string(ctx.Body()))
	})

	// 测试数据限速16B
	app.AddMiddleware(middleware.NewRateSpeedFunc(16, 128, app.Context))
	app.AnyFunc("/get", func(ctx eudore.Context) {
		for i := 0; i < 10; i++ {
			ctx.WriteString("rate speed =16B\n")
		}
	})
	app.PostFunc("/post", func(ctx eudore.Context) {
		ctx.Debug(string(ctx.Body()))
	})

	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(http.TimeoutHandler(app, 2*time.Second, ""))
	client.NewRequest("GET", "/get").Do().CheckStatus(200)
	client.NewRequest("POST", "/post").WithBody("read body is to long,body太大，会中间件超时无法完全读取。").Do().CheckStatus(200)
	client.NewRequest("POST", "/done").WithBody("hello").Do().CheckStatus(200)
	// app.CancelFunc()
	app.Run()
}
