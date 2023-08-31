package main

import (
	"crypto/tls"
	"net"
	"net/http"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"golang.org/x/net/http2"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(
		middleware.NewLoggerFunc(app),
		middleware.NewCompressMixinsFunc(nil),
	)
	app.GetFunc("/", func(ctx eudore.Context) {
		ctx.Push("/css/app.css", &http.PushOptions{
			Header: http.Header{eudore.HeaderAuthorization: {"00"}},
		})
		ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTMLCharsetUtf8)
		ctx.WriteString(`<!DOCTYPE html>
<html>
<head><title>push</title><link href='/css/app.css' rel="stylesheet"></head>
<body>push test, push css is red font.</body>
</html>`)
	})
	app.GetFunc("/css/*", func(ctx eudore.Context) {
		if ctx.GetHeader(eudore.HeaderAuthorization) == "" {
			ctx.WriteHeader(eudore.StatusUnauthorized)
			return
		}
		ctx.WithField("header", ctx.Request().Header).Debug()
		ctx.SetHeader(eudore.HeaderContentType, "text/css")
		ctx.WriteString("*{color: red;}")
	})
	app.Listen(":8088")
	app.ListenTLS(":8089", "", "")

	client := app.WithClient(&http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return tls.Dial(network, addr, cfg)
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})
	client.NewRequest(nil, "GET", "https://localhost:8089/", eudore.NewClientCheckStatus(200))

	app.Run()
}
