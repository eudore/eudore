package main

/*
// Context 对象相关内容。
type Context interface {
	Request() *RequestReader
	Response() ResponseWriter
	GetHeader(name string) string
	SetHeader(string, string)
	...
}

Request().Header 为请求header
Response().Header() 为响应header
ctx.GetHeader(key) => ctx.Request().Header.Get(key)
ctx.SetHeader(key, val) = > ctx.Response().Header().Set(key, val)
*/

import (
	"fmt"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.AnyFunc("/get", func(ctx eudore.Context) {
		// 遍历请求header
		for k, v := range ctx.Request().Header {
			fmt.Fprintf(ctx, "%s: %s\n", k, v)
		}
		// 获取一个请求header
		ctx.SetHeader("name", "eudore")
		ctx.Infof("user-agent: %s", ctx.GetHeader("User-Agent"))
	})

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/get").WithHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationForm).Do().CheckStatus(200).Out()

	for client.Next() {
		app.Error(client.Error())
	}

	app.Run()
}
