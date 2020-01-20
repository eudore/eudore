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
	app := eudore.NewCore()
	app.AddMiddleware(middleware.NewLoggerFunc(app.App, "action", "ram", "route", "resource", "browser"))

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/get").Do().CheckStatus(200).CheckBodyContainString("get", "/*path")
	client.NewRequest("GET", "/get/ha").Do().CheckStatus(200).CheckBodyContainString("/get/:name")
	client.NewRequest("GET", "/get/eudore").Do().CheckStatus(200).CheckBodyContainString("/get/eudore")
	client.NewRequest("PUT", "/get/eudore").Do().CheckStatus(200).CheckBodyContainString("any", "/*path")
	for client.Next() {
		app.Error(client.Error())
	}
	client.Stop(0)

	// 启动server
	app.Listen(":8088")
	app.Run()
}
