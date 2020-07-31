package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/pprof"
)

func main() {
	app := eudore.NewApp()
	pprof.Init(app.Group("/eudore/debug"))
	app.AnyFunc("/eudore/debug/pprof/look/* godoc=/eudore/debug/pprof/godoc", pprof.NewLook(app))

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/eudore/debug/show/").Do().OutBody()
	client.NewRequest("GET", "/eudore/debug/pprof/look/?d=1").Do().OutBody()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
