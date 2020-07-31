package main

import (
	"fmt"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "action", "ram", "route", "resource", "browser"))
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debug(ctx.GetCookie("name"))
		ctx.SetCookieValue("resp", "hello", 0)
	})

	// 请求测试
	client := httptest.NewClient(app)
	client.AddCookie("/", "name", "eudore")
	client.NewRequest("GET", "/get").Do()
	fmt.Println(client.GetCookie("/", "name"))
	fmt.Println(client.GetCookie("/", "resp"))

	app.CancelFunc()
	app.Run()
}
