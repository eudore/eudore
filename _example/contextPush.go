package main

import (
	"crypto/tls"
	"net"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"golang.org/x/net/http2"
)

func main() {
	app := eudore.NewApp()
	app.GetFunc("/", func(ctx eudore.Context) {
		ctx.Debug(ctx.Request().Proto)
		ctx.Push("/css/1.css", nil)
		ctx.Push("/css/2.css", nil)
		ctx.Push("/css/3.css", nil)
		ctx.Push("/favicon.ico", nil)
		ctx.WriteString(`<!DOCTYPE html>
<html>
<head>
	<title>push</title>
	<link href='/css/1.css' rel="stylesheet">
	<link href='/css/2.css' rel="stylesheet">
	<link href='/css/3.css' rel="stylesheet">
</head>
<body>
push test
</body>
</html>`)
	})
	app.GetFunc("/hijack", func(ctx eudore.Context) {
		conn, _, err := ctx.Response().Hijack()
		if err == nil {
			conn.Close()
		}
	})
	app.GetFunc("/css/*", func(ctx eudore.Context) {
		ctx.WriteString("*{}")
	})
	app.ListenTLS(":8088", "", "")

	client := httptest.NewClient(app)
	client.Client.Transport = &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return tls.Dial(network, addr, cfg)
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.NewRequest("GET", "/").Do().CheckStatus(200).Out()
	client.NewRequest("GET", "https://localhost:8088/").Do().CheckStatus(200).Out()
	client.NewRequest("GET", "https://localhost:8088/hijack").Do().CheckStatus(200).Out()
	for client.Next() {
		app.Error(client.Error())
	}

	app.CancelFunc()
	app.Run()
}
