package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AddHandlerExtend(func(interface{}) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			ctx.Debugf("%s extend: %v", ctx.Path(), "default")
		}
	})

	g1 := app.Group("")
	g1.AddHandlerExtend(func(interface{}) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			ctx.Debugf("%s extend: %v", ctx.Path(), "g1")
		}
	})

	g2 := app.Group("")
	g2.AddHandlerExtend(func(interface{}) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			ctx.Debugf("%s extend: %v", ctx.Path(), "g2")
		}
	})

	app.GetFunc("/*", "")
	g1.GetFunc("/api/*", "")
	g2.GetFunc("/api/v1/11", "")

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/file/").Do()
	client.NewRequest("GET", "/api/11").Do()
	client.NewRequest("GET", "/api/v1/11").Do()
	client.NewRequest("GET", "/api/v2/11").Do()

	app.CancelFunc()
	app.Run()
}
