package main

import (
	"net"
	"net/http"
	"time"

	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyServer, eudore.NewServer(&eudore.ServerConfig{
		// 设置配置
		ReadTimeout:  eudore.TimeDuration(4 * time.Second),
		WriteTimeout: eudore.TimeDuration(12 * time.Second),
		IdleTimeout:  eudore.TimeDuration(60 * time.Second),
	}))

	// 自定义Handler，默认是app
	app.Server.SetHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.Debug("warp", r.RequestURI)
		app.ServeHTTP(w, r)
	}))

	ln, err := net.Listen("TCP", ":8089")
	if err != nil {
		app.Error(err)
	} else {
		go app.Server.Serve(ln)
	}

	app.Listen(":8088")
	app.Run()
}
