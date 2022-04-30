package main

/*
Eudore.Context实现并发安全需要对几乎所有操作加锁，成本太大使用http.TimeoutHandler仅对全局进行超时限制。
*/

import (
	"net/http"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.SetHandler(http.TimeoutHandler(app, 3*time.Second/10, ""))
	app.AddMiddleware(
		middleware.NewLoggerFunc(app, "route"),
		middleware.NewRecoverFunc(),
		// middleware.NewTimeoutFunc(3*time.Second/10),
	)
	app.AnyFunc("/*", func(ctx eudore.Context) {
		time.Sleep(time.Duration(eudore.GetStringInt64(ctx.GetParam("*"))) * time.Second / 10)
		ctx.WriteString("hello")
		ctx.WriteString("eudore")
	})
	app.AnyFunc("/h/*", func(ctx eudore.Context) {
		time.Sleep(time.Duration(eudore.GetStringInt64(ctx.GetParam("*"))) * time.Second / 10)
		ctx.SetHeader("X-MetaData", "timeout")
		ctx.WriteHeader(200)
		ctx.WriteString("hello")
		ctx.WriteString("eudore")
	})
	app.AnyFunc("/panic", func(ctx eudore.Context) {
		time.Sleep(time.Second / 10)
		var n int
		ctx.Debug(11 / n)
	})

	client := httptest.NewClient(http.TimeoutHandler(app, 3*time.Second/10, ""))
	client.NewRequest("PUT", "/1").Do().CheckStatus(200)
	client.NewRequest("PUT", "/2").Do().CheckStatus(200)
	client.NewRequest("PUT", "/4").Do().CheckStatus(503)
	client.NewRequest("PUT", "/5").Do().CheckStatus(503)
	client.NewRequest("PUT", "/h/1").Do().CheckStatus(200)
	client.NewRequest("PUT", "/h/2").Do().CheckStatus(200)
	client.NewRequest("PUT", "/h/4").Do().CheckStatus(503)
	client.NewRequest("PUT", "/h/5").Do().CheckStatus(503)
	client.NewRequest("PUT", "/panic").Do().CheckStatus(503)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
