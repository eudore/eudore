package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyRouter, eudore.NewRouterStd(eudore.NewRouterCoreHost(nil)))

	app.AnyFunc("/* host=eudore.com", echoHandleHost)
	app.AnyFunc("/* host=eudore.com:8088", echoHandleHost)
	app.AnyFunc("/* host=eudore.cn", echoHandleHost)
	app.AnyFunc("/* host=eudore.*", echoHandleHost)
	app.AnyFunc("/* host=example.com", echoHandleHost)
	app.AnyFunc("/* host=www.*.cn", echoHandleHost)
	app.AnyFunc("/api/* host=*", echoHandleHost)
	app.AnyFunc("/api/* host=eudore.com,eudore.cn", echoHandleHost)
	app.AnyFunc("/*", echoHandleHost)

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do().CheckStatus(200).CheckBodyString("")
	client.NewRequest("GET", "/").WithHeaderValue("Host", "eudore.cn").Do().CheckStatus(200).CheckBodyString("eudore.cn")
	client.NewRequest("GET", "/").WithHeaderValue("Host", "eudore.com").Do().CheckStatus(200).CheckBodyString("eudore.com")
	client.NewRequest("GET", "/").WithHeaderValue("Host", "eudore.com:8088").Do().CheckStatus(200).CheckBodyString("eudore.com")
	client.NewRequest("GET", "/").WithHeaderValue("Host", "eudore.com:8089").Do().CheckStatus(200).CheckBodyString("eudore.com")
	client.NewRequest("GET", "/").WithHeaderValue("Host", "eudore.net").Do().CheckStatus(200).CheckBodyString("eudore.*")
	client.NewRequest("GET", "/").WithHeaderValue("Host", "www.eudore.cn").Do().CheckStatus(200).CheckBodyString("www.*.cn")
	client.NewRequest("GET", "/").WithHeaderValue("Host", "example.com").Do().CheckStatus(200).CheckBodyString("example.com")
	client.NewRequest("GET", "/").WithHeaderValue("Host", "www.example").Do().CheckStatus(200).CheckBodyString("")
	client.NewRequest("GET", "/api/v1").WithHeaderValue("Host", "example.com").Do().CheckStatus(200).CheckBodyString("*")
	client.NewRequest("GET", "/api/v1").WithHeaderValue("Host", "eudore.com").Do().CheckStatus(200).CheckBodyString("eudore.com,eudore.cn")

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

func echoHandleHost(ctx eudore.Context) {
	ctx.WriteString(ctx.GetParam("host"))
}
