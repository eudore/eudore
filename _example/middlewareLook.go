package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	config := make(map[interface{}]interface{})
	app := eudore.NewApp()
	app.Set("conf", config)
	config[true] = 1
	config[1] = 11
	config[uint(1)] = 11
	config[1.0] = 11
	config[complex(1, 1)] = 11

	var i interface{}
	config[i] = 0

	app.AnyFunc("/eudore/debug/look/* godoc=/eudore/debug/pprof/godoc", middleware.NewLookFunc(app))

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/eudore/debug/look/?d=1").Do().OutBody()
	client.NewRequest("GET", "/eudore/debug/look/?all=1").Do().OutBody()
	client.NewRequest("GET", "/eudore/debug/look/?format=txt").Do().OutBody()
	client.NewRequest("GET", "/eudore/debug/look/?format=t2").Do().OutBody()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
