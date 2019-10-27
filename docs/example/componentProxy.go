package main

/*
eudore.HandlerProxy实现参考net/http/httputil.NewSingleHostReverseProxy
*/

import (
	"net/http/httputil"
	"net/url"

	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	addr, _ := url.Parse("http://localhost:8089")
	app.AnyFunc("/*", httputil.NewSingleHostReverseProxy(addr).ServeHTTP)

	go func() {
		app := eudore.NewCore()
		app.AnyFunc("/*", func(ctx eudore.Context) {
			ctx.WriteString("host: " + ctx.Host())
			ctx.WriteString("\r\nrealip: " + ctx.RealIP())
		})
		app.Listen(":8089")
		app.Run()
	}()

	app.Listen(":8088")
	app.Run()
}
