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
	app.AddMiddleware(middleware.NewBasicAuthFunc(data))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/1").Do()
	// 全局设置basic auth信息
	client.AddBasicAuth("user", "pw")
	client.NewRequest("GET", "/2").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
