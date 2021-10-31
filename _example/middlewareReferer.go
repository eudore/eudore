package main

/*
检查请求Referer Header值是否有效

	""                         =>    其他值未匹配时使用的默认值。
	"origin"                   =>    请求Referer和Host同源情况下，检查host为referer前缀，origin检查在其他值检查之前。
	"*"                        =>    任意域名端口
	"www.eudore.cn/*"          =>    www.eudore.cn域名全部请求，不指明http或https时为同时包含http和https
	"www.eudore.cn/api/*"      =>    www.eudore.cn域名全部/api/前缀的请求
	"https://www.eudore.cn/*"  =>    www.eudore.cn仅匹配https。
*/

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
		"http://126.0.0.1:*/*":     true,
		"http://127.0.0.1:*/*":     true,
		"http://128.0.0.1:*/*":     true,
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
	client.AddHeaderValue(eudore.HeaderHost, "www.eudore.cn")
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "").Do().CheckStatus(200)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "http://www.eudore.cn/").Do().CheckStatus(200)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderReferer, "http://www.example.com/").Do().CheckStatus(403)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
