package main

import (
	"io"
	"net"
	"net/http/httputil"

	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware("global", func(ctx eudore.Context) {
		if ctx.Request().URL.Host != "" {
			defer ctx.End()
			if ctx.Method() == eudore.MethodConnect {
				// 隧道代理
				conn, err := net.Dial("tcp", ctx.Request().URL.Host)
				if err != nil {
					ctx.WriteHeader(502)
					ctx.Error(err)
					return
				}

				client, _, err := ctx.Response().Hijack()
				if err != nil {
					ctx.WriteHeader(502)
					ctx.Error(err)
					conn.Close()
					return
				}
				client.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))
				go func() {
					io.Copy(client, conn)
					client.Close()
					conn.Close()
				}()
				go func() {
					io.Copy(conn, client)
					client.Close()
					conn.Close()
				}()
			} else {
				// 反向代理
				httputil.NewSingleHostReverseProxy(ctx.Request().URL).ServeHTTP(ctx.Response(), ctx.Request())
			}
		}
	})

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
