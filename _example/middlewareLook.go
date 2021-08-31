package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	config := make(map[interface{}]interface{})
	app := eudore.NewApp(eudore.Renderer(eudore.RenderJSON))
	app.Logger = app.WithField("key", "look").WithFields(nil, nil)
	app.Set("conf", config)
	config[true] = 1
	config[1] = 11
	config[uint(1)] = 11
	config[1.0] = 11.0
	config[complex(1, 1)] = complex(1, 1)

	var i interface{}
	config[i] = 0

	app.AnyFunc("/eudore/debug/look/* godoc=/eudore/debug/pprof/godoc", middleware.NewLookFunc(app))

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/eudore/debug/look/?d=3").Do()
	client.NewRequest("GET", "/eudore/debug/look/?all=1").Do()
	client.NewRequest("GET", "/eudore/debug/look/?format=text").Do()
	client.NewRequest("GET", "/eudore/debug/look/?format=json").Do()
	client.NewRequest("GET", "/eudore/debug/look/?format=t2").Do()
	client.NewRequest("GET", "/eudore/debug/look/Config/Keys/2").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
