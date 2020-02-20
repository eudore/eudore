package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	app.AddHandlerExtend(func(interface{}) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			ctx.Debugf("%s extend: %v", ctx.Path(), 999)
		}
	})
	app.AddHandlerExtend("/api/", func(interface{}) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			ctx.Debugf("%s extend: %v", ctx.Path(), "api")
		}
	})
	app.AddHandlerExtend("/api/v1/", func(interface{}) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			ctx.Debugf("%s extend: %v", ctx.Path(), "api v1")
		}
	})
	app.GetFunc("/*", "")
	app.GetFunc("/api/*", "")
	app.GetFunc("/api/v1/11", "")

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/file/").Do()
	client.NewRequest("GET", "/api/11").Do()
	client.NewRequest("GET", "/api/v1/11").Do()
	client.NewRequest("GET", "/api/v2/11").Do()
	client.Stop(0)

	app.Listen(":8088")
	app.Run()
}
