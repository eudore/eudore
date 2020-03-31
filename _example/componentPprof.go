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
	app := eudore.NewCore()
	pprof.RoutesInject(app.Group("/eudore/debug"))

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/eudore/debug/pprof/goroutine?debug=1").Do().OutBody()

	app.Run()
}
