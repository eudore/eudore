package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	routerdata := map[string]interface{}{
		"/api/:v/*": func(ctx eudore.Context) {
			ctx.Request().URL.Path = "/api/v3/" + ctx.GetParam("*")
		},
		"GET /api/:v/*": func(ctx eudore.Context) {
			ctx.WriteHeader(403)
			ctx.End()
		},
	}

	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewLoggerFunc(app, "route", "*"))
	app.AddMiddleware(middleware.NewRouterFunc(routerdata))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/api/v1/user").Do()
	client.NewRequest("PUT", "/api/v1/user").Do()
	client.NewRequest("PUT", "/api/v2/user").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
