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
	// app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
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
	client.AddHeaderValue(eudore.HeaderAcceptEncoding, "gzip")
	client.NewRequest("GET", "/eudore/debug/dump/connect").Do()
	go func() {
		client.NewRequest("GET", "/eudore/debug/dump/connect").WithWebsocket(ReadDumpMessage).Do()
	}()
	go func() {
		time.Sleep(time.Millisecond * 10)
		client.NewRequest("GET", "/eudore/debug/dump/connect").WithWebsocket(closeDumpMessage).Do()
		client.NewRequest("GET", "/").Do()
		client.NewRequest("GET", "/").Do()
		client.NewRequest("GET", "/eudore/debug/dump/connect").WithWebsocket(ReadDumpMessage).Do()
	}()
	go func() {
		client.NewRequest("GET", "/eudore/debug/dump/connect?path=/echo").WithWebsocket(ReadDumpMessage).Do()
	}()
	go func() {
		client.NewRequest("GET", "/eudore/debug/dump/connect").WithWebsocket(closeDumpMessage).Do()
		client.NewRequest("GET", "/eudore/debug/dump/connect").WithWebsocket(closeDumpMessage).Do()
		client.NewRequest("GET", "/eudore/debug/dump/connect").WithWebsocket(closeDumpMessage).Do()
		client.NewRequest("GET", "/eudore/debug/dump/connect").WithWebsocket(closeDumpMessage).Do()
	}()
	go func() {
		client.NewRequest("GET", "/eudore/debug/dump/connect?path=/echo&query-name=eudore*").WithWebsocket(ReadDumpMessage).Do()
	}()
	go func() {
		client.NewRequest("GET", "/eudore/debug/dump/connect?param-route=/echo").WithWebsocket(ReadDumpMessage).Do()
	}()
	go func() {
		client.NewRequest("GET", "/eudore/debug/dump/connect?header-origin=https://*").WithWebsocket(ReadDumpMessage).Do()
	}()
	go func() {
		ticker := time.Tick(time.Millisecond * 100)
		for {
			select {
			case <-ticker:
				return
			default:
				time.Sleep(time.Millisecond * 1)
				go client.NewRequest("GET", "/eudore/debug/dump/connect").WithWebsocket(closeDumpMessage).Do()
				go client.NewRequest("GET", "/").Do()
			}
		}
	}()
	go func() {
		ticker := time.Tick(time.Millisecond * 100)
		for {
			select {
			case <-ticker:
				return
			default:
				time.Sleep(time.Millisecond * 1)
				go client.NewRequest("GET", "/eudore/debug/dump/connect").WithWebsocket(ReadDumpMessage).Do()
				go client.NewRequest("GET", "/").Do()
			}
		}
	}()
	go func() {
		time.Sleep(time.Millisecond * 20)
		client.NewRequest("GET", "/").Do()
		client.NewRequest("GET", "/echo").Do()
		client.NewRequest("GET", "/bigbody").Do()
		client.NewRequest("GET", "/gzip").Do()
		time.Sleep(time.Millisecond * 40)
		client.NewRequest("GET", "/gziperr").Do()
		time.Sleep(time.Millisecond * 40)
		// app.CancelFunc()
	}()

	app.Listen(":8088")
	app.Run()
}

func closeDumpMessage(conn net.Conn) {
	time.Sleep(time.Millisecond * 4)
	conn.Close()
}
func ReadDumpMessage(conn net.Conn) {
	io.Copy(ioutil.Discard, conn)
}
