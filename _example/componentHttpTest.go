package main

/*
待设计完善
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "action", "ram", "route", "resource", "browser"))
	app.AnyFunc("/*", func(ctx eudore.Context) {})
	app.Listen(":8088")

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/get").Do().CheckStatus(404)
	client.NewRequest("GET", "127.0.0.1:8080").Do().CheckStatus(500)
	client.NewRequest("GET", "http://127.0.0.1:8088").Do().CheckStatus(200).Out()
	for client.Next() {
		app.Error(client.Error())
	}

	app.CancelFunc()
	app.Run()
}
