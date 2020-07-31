package main

/*
httptest作为默认测试客户端，用于发送http客户端请求
*/

import (
	"crypto/tls"
	"net/http"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "action", "ram", "route", "resource", "browser"))
	app.AnyFunc("/*", eudore.HandlerEmpty)
	app.Listen(":8088")
	app.ListenTLS(":8089", "", "")

	// 请求测试
	client := httptest.NewClient(app)
	client.Client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.AddQuery("name", "eudore")
	// 使用http.Handler接口处理构造请求
	client.NewRequest("GET", "/get").Do().CheckStatus(200)
	client.NewRequest("GET", "127.0.0.1:8080").Do().CheckStatus(500)
	// 使用http.Client发送请求
	client.NewRequest("GET", "http://127.0.0.1:8088").Do().CheckStatus(200)
	client.NewRequest("GET", "https://127.0.0.1:8089").Do().CheckStatus(200)

	// app.CancelFunc()
	app.Run()
}
