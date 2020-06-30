package main

import (
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(
		middleware.NewLoggerFunc(app, "route"),
		middleware.NewRecoverFunc(),
		middleware.NewTimeoutFunc(3*time.Second/10),
	)
	app.AnyFunc("/*", func(ctx eudore.ContextData) {
		time.Sleep(time.Duration(ctx.GetParamInt64("*")) * time.Second / 10)
		ctx.WriteString("hello")
		ctx.WriteString("eudore")
	})
	app.AnyFunc("/h/*", func(ctx eudore.ContextData) {
		time.Sleep(time.Duration(ctx.GetParamInt64("*")) * time.Second / 10)
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

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/1").Do().CheckStatus(200)
	client.NewRequest("PUT", "/2").Do().CheckStatus(200)
	client.NewRequest("PUT", "/3").Do().CheckStatus(200)
	client.NewRequest("PUT", "/4").Do().CheckStatus(200)
	client.NewRequest("PUT", "/5").Do().CheckStatus(200)
	client.NewRequest("PUT", "/h/1").Do().CheckStatus(200)
	client.NewRequest("PUT", "/h/2").Do().CheckStatus(200)
	client.NewRequest("PUT", "/h/3").Do().CheckStatus(200)
	client.NewRequest("PUT", "/h/4").Do().CheckStatus(200)
	client.NewRequest("PUT", "/h/5").Do().CheckStatus(200)
	client.NewRequest("PUT", "/panic").Do().CheckStatus(200)
	for client.Next() {
		app.Error(client.Error())
	}

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
