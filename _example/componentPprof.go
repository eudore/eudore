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
	httptest.NewClient(app).Stop(0)
	pprof.RoutesInject(app.Group("/eudore/debug"))

	app.Listen(":8088")
	app.Run()
}
