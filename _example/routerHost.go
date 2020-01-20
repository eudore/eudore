package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/router/host"
)

func main() {
	rh := host.NewRouterHost()
	rh.SetRouter("example.com", eudore.NewRouterRadix())

	app := eudore.NewEudore(rh)
	eudore.Set(app.Router, "print", eudore.NewPrintFunc(app.App))
	app.AddGlobalMiddleware(host.InitAddHost)
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("start fasthttp server, this default page.")
	})

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").WithHeaderValue("Host", "example.com").Do().CheckStatus(200).CheckBodyContainString("start fasthttp server, this default page.").Out()
	client.NewRequest("GET", "/").WithHeaderValue("Host", "eudore.cn").Do().CheckStatus(200).CheckBodyContainString("start fasthttp server, this default page.").Out()
	for client.Next() {
		app.Error(client.Error())
	}
	client.Stop(0)

	app.Listen(":8088")
	app.Run()
}
