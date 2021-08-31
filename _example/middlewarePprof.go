package main

/*
pprof使用eudore访问net/http/pprof显示调试信息。

实际访问路径 /eudore/debug/pprof/
*/

import (
	"fmt"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
	"os"
)

func main() {
	app := eudore.NewApp()
	app.Group("/eudore/debug").AddController(middleware.NewPprofController())

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/eudore/debug/pprof/expvar").WithHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do()
	client.NewRequest("GET", "/eudore/debug/pprof/?format=json").Do()
	client.NewRequest("GET", "/eudore/debug/pprof/?format=text").Do()
	client.NewRequest("GET", "/eudore/debug/pprof/?format=html").Do()
	client.NewRequest("GET", "/eudore/debug/pprof/goroutine?debug=0").Do()
	client.NewRequest("GET", "/eudore/debug/pprof/goroutine?debug=1").Do()
	client.NewRequest("GET", "/eudore/debug/pprof/goroutine?debug=1&format=json").Do()
	client.NewRequest("GET", "/eudore/debug/pprof/goroutine?debug=1&format=text").Do()
	client.NewRequest("GET", "/eudore/debug/pprof/goroutine?debug=1&format=html").Do()
	client.NewRequest("GET", "/eudore/debug/pprof/goroutine?debug=2&format=json").Do()
	client.NewRequest("GET", "/eudore/debug/pprof/goroutine?debug=2&format=text").Do()
	client.NewRequest("GET", "/eudore/debug/pprof/goroutine?debug=2&format=html").Do()

	fmt.Println(os.Environ())
	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
