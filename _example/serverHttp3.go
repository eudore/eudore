package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"github.com/quic-go/quic-go/http3"
	"log/slog"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(
		middleware.NewLoggerFunc(app),
		NewServerV3HeaderFunc(),
	)
	app.GetFunc("/*", func(ctx eudore.Context) {
		type wraper interface{ Unwrap() http.ResponseWriter }
		r := ctx.Request()
		w := ctx.Response().(wraper).Unwrap()
		_, ok1 := w.(http.Flusher)
		_, ok2 := w.(http.Hijacker)
		_, ok3 := w.(http.Pusher)
		_, ok4 := w.(interface{ EnableFullDuplex() error })
		_, ok5 := w.(interface{ SetWriteDeadline(time.Time) error })
		_, ok6 := w.(interface{ SetReadDeadline(time.Time) error })
		_, ok7 := w.(wraper)

		ctx.WriteHeader(200)
		fmt.Fprintln(ctx, r.Proto, r.TransferEncoding, r.ContentLength)
		fmt.Fprintln(ctx, "http.Flusher", ok1)
		fmt.Fprintln(ctx, "http.Hijacker", ok2)
		fmt.Fprintln(ctx, "http.Pusher", ok3)
		fmt.Fprintln(ctx, "EnableFullDuplex", ok4)
		fmt.Fprintln(ctx, "SetWriteDeadline", ok5)
		fmt.Fprintln(ctx, "SetReadDeadline", ok6)
		fmt.Fprintln(ctx, "Unwrap", ok7)
	})

	server := &http3.Server{
		Port:    8088,
		Addr:    "0.0.0.0:8088",
		Handler: app,
		Logger:  slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})),
	}
	// 必须使用有效tls证书
	go server.ListenAndServeTLS("tls.crt", "tls.key")

	app.ListenTLS(":8088", "tls.crt", "tls.key")
	app.Run()
}

func NewServerV3HeaderFunc() eudore.HandlerFunc {
	// https://quic-go.net/docs/http3/server/#advertising-http3-via-alt-svc
	// 通过HeaderAltSvc 设置h3升级。
	svc := `h3=":8088"; ma=2592000`
	return func(ctx eudore.Context) {
		r := ctx.Request()
		if r.TLS != nil && r.ProtoMajor < 3 {
			ctx.SetHeader(eudore.HeaderAltSvc, svc)
		}
	}
}
