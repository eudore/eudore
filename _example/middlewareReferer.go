package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewRefererFunc(map[string]bool{
		"":                         true,
		"origin":                   false,
		"www.eudore.cn/*":          true,
		"www.eudore.cn/api/*":      false,
		"www.example.com/*":        true,
		"www.example.com/*/*":      false,
		"www.example.com/*/2":      true,
		"http://127.0.0.1/*":       true,
		"http://127.0.0.1:*/*":     true,
		"http://localhost/api/*":   true,
		"http://localhost:*/api/*": true,
	}))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "").Do().CheckStatus(200)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderHost, "www.eudore.cn").WithHeaderValue(eudore.HeaderReferer, "http://www.eudore.cn/").Do().CheckStatus(403)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderHost, "www.eudore.cn").WithTLS().WithHeaderValue(eudore.HeaderReferer, "https://www.eudore.cn/").Do().CheckStatus(403)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "http://www.eudore.cn/").Do().CheckStatus(200)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "http://www.example.com").Do().CheckStatus(200)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "http://www.example.com/").Do().CheckStatus(200)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "http://www.example.com/1").Do().CheckStatus(200)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "http://www.example.com/1/1").Do().CheckStatus(403)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "http://www.example.com/1/2").Do().CheckStatus(200)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "http://127.0.0.1/1").Do().CheckStatus(200)
	for client.Next() {
		app.Error(client.Error())
	}

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()

	referer2()
}

func referer2() {
	app := eudore.NewApp()
	// 仅允许无referer或与host相同的同源referer值。
	app.AddMiddleware(middleware.NewRefererFunc(map[string]bool{
		"":       true,
		"origin": true,
		"*":      false,
	}))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.WithHeaderValue(eudore.HeaderHost, "www.eudore.cn")
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "").Do().CheckStatus(200)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "http://www.eudore.cn/").Do().CheckStatus(200)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "http://www.example.com/").Do().CheckStatus(403)

	for client.Next() {
		app.Error(client.Error())
	}

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
