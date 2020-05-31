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
	app.SetHandler(host.NewHandler(app))

	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("start server, this default page.")
	})

	// 请求测试
	client := httptest.NewClient(host.NewHandler(app))
	client.NewRequest("GET", "/").WithHeaderValue("Host", "example.com").Do().CheckStatus(200).CheckBodyContainString("start server").Out()
	client.NewRequest("GET", "/").WithHeaderValue("Host", "eudore.cn").Do().CheckStatus(200).CheckBodyContainString("start server").Out()
	for client.Next() {
		app.Error(client.Error())
	}

	app.CancelFunc()
	app.Run()
}
