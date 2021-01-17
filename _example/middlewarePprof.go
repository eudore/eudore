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
	app.Group("/eudore/debug2 godoc=localhost:6020").AddController(middleware.NewPprofController())

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/eudore/debug/pprof/expvar").WithHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do().OutBody()
	client.NewRequest("GET", "/eudore/debug/pprof/goroutine?debug=1").Do().OutBody()
	client.NewRequest("GET", "/eudore/debug/pprof/goroutine?debug=2").Do().OutBody()
	client.NewRequest("GET", "/eudore/debug/pprof/goroutine?debug=1&m=txt").Do().OutBody()
	client.NewRequest("GET", "/eudore/debug2/pprof/goroutine?debug=1").Do().OutBody()

	fmt.Println(os.Environ())
	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
