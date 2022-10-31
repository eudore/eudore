package main

import (
	"io/ioutil"

	"github.com/andybalholm/brotli"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(
		middleware.NewCompressFunc("br", func() interface{} { return brotli.NewWriter(ioutil.Discard) }),
		middleware.NewCompressGzipFunc(5),
		middleware.NewCompressDeflateFunc(5),
	)
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("eudore compress")
	})

	// 默认浏览器需要https协议才会使用br压缩
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAcceptEncoding, "br"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAcceptEncoding, "deflate"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAcceptEncoding, "gzip"))

	app.ListenTLS(":8088", "", "")
	app.Run()
}
