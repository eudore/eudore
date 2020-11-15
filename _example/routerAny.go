package main

/*
路由方法优先级： 具体方法 > Any方法
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.WriteString("route: " + ctx.GetParam("route") + "\n")
	})
	// 通配符覆盖
	app.GetFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("method is get\n")
	})
	// Any方法不会覆盖Get方法。
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("method is any\n")
	})
	app.PostFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("method is post\n")
	})

	// 遍历覆盖
	app.GetFunc("/get/:val", func(ctx eudore.Context) {
		ctx.WriteString("method is get\n")
	})
	app.AnyFunc("/get/:val", func(ctx eudore.Context) {
		ctx.WriteString("method is any\n")
	})
	app.PostFunc("/get/:val", func(ctx eudore.Context) {
		ctx.WriteString("method is post\n")
	})
	app.AddHandler("444", "", eudore.HandlerRouter404)
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/get").Do().CheckStatus(200).CheckBodyContainString("get").OutBody()
	client.NewRequest("POST", "/get").Do().CheckStatus(200).CheckBodyContainString("post").OutBody()
	client.NewRequest("PUT", "/get").Do().CheckStatus(200).CheckBodyContainString("any").OutBody()
	client.NewRequest("GET", "/get/1").Do().CheckStatus(200).CheckBodyContainString("get").OutBody()
	client.NewRequest("POST", "/get/2").Do().CheckStatus(200).CheckBodyContainString("post").OutBody()
	client.NewRequest("PUT", "/get/3").Do().CheckStatus(200).CheckBodyContainString("any").OutBody()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
