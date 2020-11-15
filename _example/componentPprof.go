package main

/*
pprof使用eudore访问net/http/pprof显示调试信息。

实际访问路径 /eudore/debug/pprof/
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/pprof"
)

func main() {
	app := eudore.NewApp()
	pprof.Init(app.Group("/eudore/debug godoc=https://golang.org"))

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/eudore/debug/pprof/expvar").WithHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do().OutBody()
	client.NewRequest("GET", "/eudore/debug/pprof/goroutine?debug=1").Do().OutBody()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
