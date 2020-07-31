package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
	"net/http"
)

func main() {
	app := eudore.NewApp()
	// 参数是压缩等级
	app.AddMiddleware(middleware.NewGzipFunc(5))
	app.AddMiddleware(middleware.NewGzipFunc(10))
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Push("/stat", nil)
		ctx.Response().Push("/stat", nil)
		ctx.Response().Push("/stat", &http.PushOptions{})
		ctx.Response().Push("/stat", &http.PushOptions{Header: make(http.Header)})
		ctx.WriteString("gzip")
		ctx.Response().Flush()
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/1").Do()
	client.AddHeaderValue(eudore.HeaderAcceptEncoding, "gzip")
	client.NewRequest("GET", "/1").Do().Out()
	client.NewRequest("GET", "https://localhost:8088/1").Do().Out()

	app.ListenTLS(":8088", "", "")
	// app.CancelFunc()
	app.Run()
}
