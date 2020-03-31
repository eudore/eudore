package main

/*
eudore兼容net/http库，所以可以直接使用net/http/httputil.NewSingleHostReverseProxy方法，直接使用标准库创建一个http反向代理处理函数。

该example额外go一个8089端口app显示访问信息，然后访问8088端口请求全部都反向代理到了8089，访问8088显示结果均为8089内容。
*/

import (
	"net/http/httputil"
	"net/url"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	addr, _ := url.Parse("http://localhost:8089")
	app.AnyFunc("/*", httputil.NewSingleHostReverseProxy(addr))

	go func() {
		app := eudore.NewCore()
		app.AnyFunc("/*", func(ctx eudore.Context) {
			ctx.WriteString("host: " + ctx.Host())
			ctx.WriteString("\r\nrealip: " + ctx.RealIP())
		})
		app.Listen(":8089")
		app.Run()
	}()

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do().Out()
	client.Stop(0)

	app.Listen(":8088")
	app.Run()
}
