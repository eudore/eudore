package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/router/host"
)

func main() {
	router := host.NewRouterHost()
	router.SetRouter("example.com", eudore.NewRouterRadix())

	app := eudore.NewApp(router)
	app.AddMiddleware("global", host.AddHostHandler)
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("start server, this default page.")
	})

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").WithHeaderValue("Host", "example.com").Do().CheckStatus(200).CheckBodyContainString("start server").Out()
	client.NewRequest("GET", "/").WithHeaderValue("Host", "eudore.cn").Do().CheckStatus(200).CheckBodyContainString("start server").Out()
	for client.Next() {
		app.Error(client.Error())
	}

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
