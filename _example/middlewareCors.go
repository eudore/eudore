package main

/*
Cors中间件具有两个参数
第一个参数是一个字符串数组，保存全部运行的Origin。
第二个参数是一个map，保存option请求匹配后返回的额外Header。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	middleware.NewCorsFunc(nil, map[string]string{
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
		"Access-Control-Expose-Headers":    "X-Request-Id",
		"access-control-max-age":           "1000",
	})

	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(middleware.NewCorsFunc([]string{"www.*.com", "example.com", "127.0.0.1:*"}, map[string]string{
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
		"Access-Control-Expose-Headers":    "X-Request-Id",
		"access-control-allow-methods":     "GET, POST, PUT, DELETE, HEAD",
		"access-control-max-age":           "1000",
	}))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("OPTIONS", "/1").Do()
	client.NewRequest("OPTIONS", "/2").WithHeaderValue("Origin", "http://"+httptest.HTTPTestHost).Do()
	client.NewRequest("OPTIONS", "/3").WithHeaderValue("Origin", "http://localhost").Do()
	client.NewRequest("OPTIONS", "/4").WithHeaderValue("Origin", "http://127.0.0.1:8088").Do()
	client.NewRequest("OPTIONS", "/5").WithHeaderValue("Origin", "http://127.0.0.1:8089").Do()
	client.NewRequest("OPTIONS", "/6").WithHeaderValue("Origin", "http://example.com").Do()
	client.NewRequest("OPTIONS", "/6").WithHeaderValue("Origin", "http://www.eudore.cn").Do()
	client.NewRequest("GET", "/1").Do()
	client.NewRequest("GET", "/2").WithHeaderValue("Origin", "http://"+httptest.HTTPTestHost).Do()
	client.NewRequest("GET", "/3").WithHeaderValue("Origin", "http://localhost").Do()
	client.NewRequest("GET", "/4").WithHeaderValue("Origin", "http://127.0.0.1:8088").Do()
	client.NewRequest("GET", "/5").WithHeaderValue("Origin", "http://127.0.0.1:8089").Do()
	client.NewRequest("GET", "/6").WithHeaderValue("Origin", "http://example.com").Do()
	client.NewRequest("GET", "/6").WithHeaderValue("Origin", "http://www.eudore.cn").Do()
	for client.Next() {
		app.Error(client.Error())
	}

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
