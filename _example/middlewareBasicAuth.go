package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	data := map[string]string{"user": "pw"}
	// map保存用户密码
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(middleware.NewBasicAuthFunc("Eudore", data))
	app.AnyFunc("/*", eudore.HandlerEmpty)
	icon := app.Group("")
	icon.AddMiddleware(middleware.NewBasicAuthFunc("", data))
	icon.AnyFunc("/favicon.ico", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/1").Do()
	client.NewRequest("GET", "/2").WithHeaderValue("Authorization", "Basic dXNlcjpwdw==").Do()
	for client.Next() {
		app.Error(client.Error())
	}

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
