package main

import (
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(middleware.NewDumpFunc(app.Group("/eudore/debug")))
	app.AnyFunc("/gzip", middleware.NewGzipFunc(5), func(ctx eudore.Context) {
		ctx.WriteString("gzip body")
	})
	app.AnyFunc("/gziperr", func(ctx eudore.Context) {
		ctx.SetHeader(eudore.HeaderContentEncoding, "gzip")
		ctx.WriteString("gzip body")
	})
	app.AnyFunc("/echo", func(ctx eudore.Context) {
		ctx.Write(ctx.Body())
	})
	app.AnyFunc("/bigbody", func(ctx eudore.Context) {
		body := []byte("0123456789abcdef0123456789abcdef0123456789abcdefx")
		for i := 0; i < 1000; i++ {
			ctx.Write(body)
		}
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.WithHeaderValue(eudore.HeaderAcceptEncoding, "gzip")
	client.NewRequest("GET", "/eudore/debug/dump/connect").WithWebsocket(ReadDumpMessage).Do().Out()
	go func() {
		client.NewRequest("ws", "/eudore/debug/dump/connect").WithWebsocket(ReadDumpMessage).Do().Out()
	}()
	go func() {
		client.NewRequest("ws", "/eudore/debug/dump/connect?path=/echo").WithWebsocket(ReadDumpMessage).Do().Out()
	}()
	go func() {
		client.NewRequest("ws", "/eudore/debug/dump/connect?path=/echo&query-name=eudore*").WithWebsocket(ReadDumpMessage).Do().Out()
	}()
	go func() {
		client.NewRequest("ws", "/eudore/debug/dump/connect?param-route=/echo").WithWebsocket(ReadDumpMessage).Do().Out()
	}()
	go func() {
		client.NewRequest("ws", "/eudore/debug/dump/connect?header-origin=https://*").WithWebsocket(ReadDumpMessage).Do().Out()
	}()
	go func() {
		time.Sleep(time.Second / 5)
		client.NewRequest("GET", "/").Do()
		client.NewRequest("GET", "/echo").Do()
		client.NewRequest("GET", "/bigbody").Do()
		client.NewRequest("GET", "/gzip").Do()
		client.NewRequest("GET", "/gziperr").Do()
		// app.CancelFunc()
	}()

	app.Listen(":8088")
	app.Run()
}

func ReadDumpMessage(conn net.Conn) {
	io.Copy(ioutil.Discard, conn)
}
