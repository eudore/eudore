package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AnyFunc("/r", func(ctx eudore.Context) {
		ctx.Redirect(301, "/")
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.WithHeaderValue(eudore.HeaderXParentID, "parent-id")
	client.NewRequest("GET", "/1?a=1").Do()
	client.NewRequest("GET", "/r").WithHeaderValue(eudore.HeaderXRequestID, "request-id").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
