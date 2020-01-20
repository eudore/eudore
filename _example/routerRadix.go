package main

/*
RouterRadix是eudore默认路由，使用基数数算法实现。

具有路由匹配优先级： 常量匹配 > 变量匹配 > 通配符匹配
方法优先级： 具体方法 > Any方法
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.WriteString("route: " + ctx.GetParam("route") + "\n")
	})
	app.GetFunc("/get/:name", func(ctx eudore.Context) {
		ctx.WriteString("get name: " + ctx.GetParam("name") + "\n")
	})
	// /get/eudore是常量匹配优先于/get/:name
	app.GetFunc("/get/eudore", func(ctx eudore.Context) {
		ctx.WriteString("get eudore\n")
	})
	app.GetFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("get path: /" + ctx.GetParam("path") + "\n")
	})
	// Any方法不会覆盖Get方法。
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("any path: /" + ctx.GetParam("path") + "\n")
	})

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
